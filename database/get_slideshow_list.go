package database

import (
	"context"
	"database/sql"
	"fmt"
	"photo_fetch/config"
)

type SlideshowServer struct {
	db           *sql.DB
	immichURL    string
	immichAPIKey string
}

func (s *SlideshowServer) getSlideshowList(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf(`
			SELECT id 
			FROM asset
			WHERE type = 'IMAGE'
			ORDER BY RANDOM() 
			LIMIT %d;`, config.SLIDESHOW_BATCH_SIZE)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var links []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		link := fmt.Sprintf("%s/api/assets/%s/thumbnail?size=preview&apiKey=%s", s.immichURL, id, s.immichAPIKey)
		links = append(links, link)
	}
	return links, nil
}

func ReturnSlideshowList(ctx context.Context, immichURL string, immichAPIKey string) ([]string, error) {
	db, err := ConnectToDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	server := &SlideshowServer{
		db:           db,
		immichURL:    immichURL,
		immichAPIKey: immichAPIKey,
	}

	return server.getSlideshowList(ctx)
}
