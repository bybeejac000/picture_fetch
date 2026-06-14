package database

import (
	"database/sql"
)

type SlideshowServer struct {
	db           *sql.DB
	immichURL    string
	immichAPIKey string
}
