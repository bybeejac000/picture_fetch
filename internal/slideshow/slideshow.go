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
	assetIDs, err := s.assetIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting slideshow list: %w", err)
	}

	for i, id := range assetIDs {
		assetIDs[i] = fmt.Sprintf("%s/api/assets/%s/thumbnail?size=preview&apiKey=%s", s.cfg.ImmichURL, id, s.cfg.ImmichROAPIKey)
	}
	return assetIDs, nil
}

func (s *Service) assetIDs(ctx context.Context) ([]string, error) {
	birthdays, err := s.birthdays(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting birthdays: %w", err)
	}

	var assetIDs []string

	if len(birthdays) > 0 {
		placeholders := make([]string, len(birthdays))
		args := make([]interface{}, len(birthdays))
		for i, p := range birthdays {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = p.ID
		}

		query := fmt.Sprintf(`
        SELECT a."id"
        FROM asset a
        JOIN asset_face af ON af."assetId" = a.id
        WHERE a.type = 'IMAGE'
        AND af."personId" IN (%s)
        ORDER BY RANDOM()
        LIMIT %d;
    `, strings.Join(placeholders, ","), s.cfg.SlideshowBatchSize)

		rows, err := s.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("querying birthday assets: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			assetIDs = append(assetIDs, id)
		}
	} else {
		query := fmt.Sprintf(`
        SELECT id
        FROM asset
        WHERE type = 'IMAGE'
        ORDER BY RANDOM()
        LIMIT %d;`, s.cfg.SlideshowBatchSize)

		rows, err := s.db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			assetIDs = append(assetIDs, id)
		}
	}

	return assetIDs, nil
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

	return birthdays, nil
}
