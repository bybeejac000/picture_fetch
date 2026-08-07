package slideshow

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"photo_fetch/internal/config"
	"photo_fetch/internal/store"
)

type Service struct {
	db    *sql.DB
	store *store.Store
	cfg   *config.Config
}

func NewService(db *sql.DB, store *store.Store, cfg *config.Config) *Service {
	return &Service{db: db, store: store, cfg: cfg}
}

type person struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ThumbnailPath string `json:"thumbnailPath"`
	FaceAssetID   string `json:"faceAssetId"`
}

type assetRef struct {
	ID   string
	Type string
}

func (s *Service) Refresh(ctx context.Context) error {
	links, err := s.list(ctx)
	if err != nil {
		return fmt.Errorf("getting slideshow list: %w", err)
	}
	if len(links) == 0 {
		return fmt.Errorf("no photos found for slideshow")
	}
	if err := s.store.ClearList(ctx, s.cfg.PhotosListKey); err != nil {
		return fmt.Errorf("clearing list in redis: %w", err)
	}
	if err := s.store.PushList(ctx, links, s.cfg.PhotosListKey); err != nil {
		return fmt.Errorf("pushing list to redis: %w", err)
	}
	return nil
}

func (s *Service) list(ctx context.Context) ([]string, error) {
	assets, err := s.assetIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting slideshow list: %w", err)
	}

	links := make([]string, 0, len(assets))
	for _, asset := range assets {
		switch asset.Type {
		case "VIDEO":
			links = append(links, fmt.Sprintf("%s/api/assets/%s/video/playback?apiKey=%s", s.cfg.ImmichURL, asset.ID, s.cfg.ImmichROAPIKey))
		default:
			links = append(links, fmt.Sprintf("%s/api/assets/%s/thumbnail?size=preview&apiKey=%s", s.cfg.ImmichURL, asset.ID, s.cfg.ImmichROAPIKey))
		}
	}
	return links, nil
}

func (s *Service) assetIDs(ctx context.Context) ([]assetRef, error) {
	birthdays, err := s.birthdays(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting birthdays: %w", err)
	}

	var assets []assetRef

	if len(birthdays) > 0 {
		placeholders := make([]string, len(birthdays))
		args := make([]interface{}, len(birthdays))
		for i, p := range birthdays {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = p.ID
		}

		userIDArgPos := len(args) + 1
		limitArgPos := len(args) + 2
		args = append(args, s.cfg.ImmichUserIds[0], s.cfg.SlideshowBatchSize)

		query := fmt.Sprintf(`
		SELECT * FROM (
			SELECT DISTINCT a."id", a."type"
			FROM asset a
			JOIN asset_face af ON af."assetId" = a.id
			WHERE a.type IN ('IMAGE', 'VIDEO')
			AND af."personId" IN (%s)
			AND (
				a."ownerId" = $%d
				OR a."ownerId" IN (
					SELECT "sharedById" FROM partner WHERE "sharedWithId" = $%d
				)
				OR a.id IN (
					SELECT aa."assetId" FROM album_asset aa
					JOIN album_user au ON au."albumId" = aa."albumId"
					WHERE au."userId" = $%d
				)
			)
			AND a."originalPath" NOT LIKE '%%.lrdata%%'
			AND a."originalPath" NOT LIKE '%%.lrcat%%'
			AND a."originalPath" NOT LIKE '%%.lrprev%%'
		) sub
		ORDER BY RANDOM()
		LIMIT $%d;
	`, strings.Join(placeholders, ","), userIDArgPos, userIDArgPos, userIDArgPos, limitArgPos)

		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil || rows.Err() != nil {
			return nil, fmt.Errorf("querying birthday assets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var asset assetRef
			if err := rows.Scan(&asset.ID, &asset.Type); err != nil {
				return nil, err
			}
			assets = append(assets, asset)
		}
	} else {

		query := `
			SELECT a.id, a.type
			FROM asset a
			WHERE a.type IN ('VIDEO', 'IMAGE')
			AND (
				a."ownerId" = $1
				OR a."ownerId" IN (
					SELECT "sharedById" FROM partner WHERE "sharedWithId" = $1
				)
				OR a.id IN (
					SELECT aa."assetId" FROM album_asset aa
					JOIN album_user au ON au."albumId" = aa."albumId"
					WHERE au."userId" = $1
				)
			)
			AND a."originalPath" NOT LIKE '%.lrdata%'
			AND a."originalPath" NOT LIKE '%.lrcat%'
			AND a."originalPath" NOT LIKE '%.lrprev%'

			ORDER BY RANDOM()
			LIMIT $2;
		`

		// Pass parameters directly into your database driver:
		rows, err := s.db.QueryContext(ctx, query, s.cfg.ImmichUserIds[0], s.cfg.SlideshowBatchSize)
		if err != nil || rows.Err() != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var asset assetRef
			if err := rows.Scan(&asset.ID, &asset.Type); err != nil {
				return nil, err
			}
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

func (s *Service) birthdays(ctx context.Context) ([]person, error) {
	query := `SELECT id
			,p."name"
			,p."thumbnailPath"
			,p."faceAssetId"
			FROM public.person p
			WHERE p."birthDate" IS NOT NULL
			AND EXTRACT(MONTH FROM "birthDate") = EXTRACT(MONTH FROM NOW() AT TIME ZONE 'America/Denver')
			AND EXTRACT(DAY FROM "birthDate") = EXTRACT(DAY FROM NOW() AT TIME ZONE 'America/Denver');`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Querying birthdays: %w", err)
	}
	defer rows.Close()

	var birthdays []person
	for rows.Next() {
		var p person
		if err := rows.Scan(&p.ID, &p.Name, &p.ThumbnailPath, &p.FaceAssetID); err != nil {
			return nil, fmt.Errorf("Scanning birthday row: %w", err)
		}
		fmt.Printf("Happy Birthday to %s", p.Name)
		birthdays = append(birthdays, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating birthday rows: %w", err)
	}

	return birthdays, nil
}
