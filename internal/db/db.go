package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DBHandler struct {
	Path string
	DB   *sql.DB
}

func NewDB(path string) *DBHandler {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("Error opening db file at '", path, "' ", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging '", path, "' file", err)
	}
	return &DBHandler{Path: path, DB: db}
}
