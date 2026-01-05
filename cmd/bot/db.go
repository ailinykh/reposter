package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/ailinykh/reposter/v3/internal/database"
)

func NewDB(logger *slog.Logger) *sql.DB {
	db, err := database.New(logger,
		database.WithURL(os.Getenv("DATABASE_URL")),
		database.WithMigrations(database.Migrations),
	)
	if err != nil {
		panic(err)
	}
	return db
}
