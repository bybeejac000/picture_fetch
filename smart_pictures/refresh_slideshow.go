package smart_pictures

import (
	"context"
	"fmt"
	"photo_fetch/config"
	"photo_fetch/database"
	"strings"
)

func getSlideshowList(ctx context.Context) ([]string, error) {
	birthdays, err := GetBirthdays()
	if err != nil {
		return nil, fmt.Errorf("failed to get birthdays: %v", err)
	}
	db, err := database.ConnectToDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	var assetIds []string

	if len(birthdays) > 0 {
		placeholders := make([]string, len(birthdays))
		args := make([]interface{}, len(birthdays))
		for i, id := range birthdays {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = id.ID
		}

		query := fmt.Sprintf(`
        SELECT a."id"
        FROM asset a
        JOIN asset_face af ON af."assetId" = a.id
        WHERE a.type = 'IMAGE'
        AND af."personId" IN (%s)
        ORDER BY RANDOM()
        LIMIT %d;
    `, strings.Join(placeholders, ","), config.SLIDESHOW_BATCH_SIZE)

		rows, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("failed to query birthday assets: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			assetIds = append(assetIds, id)
		}

	} else {
		query := fmt.Sprintf(`
        SELECT id
        FROM asset
        WHERE type = 'IMAGE'
        ORDER BY RANDOM()
        LIMIT %d;`, config.SLIDESHOW_BATCH_SIZE)

		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("query failed: %v", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			assetIds = append(assetIds, id)
		}
	}

	return assetIds, nil

}

func ReturnSlideshowList(ctx context.Context, immichURL string, immichAPIKey string) ([]string, error) {

	assetIds, err := getSlideshowList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get slideshow list: %v", err)
	}

	for i, id := range assetIds {
		assetIds[i] = fmt.Sprintf("%s/api/assets/%s/thumbnail?size=preview&apiKey=%s", immichURL, id, immichAPIKey)
	}
	return assetIds, nil
}
