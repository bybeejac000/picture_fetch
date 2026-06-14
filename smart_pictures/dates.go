package smart_pictures

import (
	"context"
	"photo_fetch/database"
	"time"
)

type birthdayPersons struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ThumbnailPath string `json:"thumbnailPath"`
	FaceAssetId   string `json:"faceAssetId"`
}

func GetBirthdays() ([]birthdayPersons, error) {
	db, err := database.ConnectToDB()
	if err != nil {
		return nil, err
	}

	query := `SELECT id
			,p."name" 
			,p."thumbnailPath" 
			,p."faceAssetId" 
			FROM public.person p
			WHERE p."birthDate" IS NOT NULL
			AND EXTRACT(MONTH FROM p."birthDate") = EXTRACT(MONTH FROM CURRENT_DATE)
			AND EXTRACT(DAY FROM p."birthDate") = EXTRACT(DAY FROM CURRENT_DATE);`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var birthdays []birthdayPersons
	for rows.Next() {
		var id, name, thumbnailPath, faceAssetId string
		if err := rows.Scan(&id, &name, &thumbnailPath, &faceAssetId); err != nil {
			return nil, err
		}
		var birthdayPerson birthdayPersons
		birthdayPerson.ID = id
		birthdayPerson.Name = name
		birthdayPerson.ThumbnailPath = thumbnailPath
		birthdayPerson.FaceAssetId = faceAssetId
		birthdays = append(birthdays, birthdayPerson)
	}

	return birthdays, nil
}
