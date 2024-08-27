package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/JeanLeonHenry/mymedia/internal/api"
	_ "modernc.org/sqlite"
)

type DBHandler struct {
	Path string
	DB   *sql.DB
}

func NewDB(path string) *DBHandler {
	// FIX: dbh should ensure the the existence of db instead of panicing?
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("Error opening db file at '", path, "' ", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging '", path, "' file", err)
	}
	return &DBHandler{Path: path, DB: db}
}

// checkDB looks up the db for a media record with case-insensitive matching titles and a year within tolerance of year
func (dbh *DBHandler) CheckDB(title string, year int, tolerance int, debug bool) bool {
	dBQuery := "SELECT title, year, id, media_type, overview, director FROM media WHERE lower(media.title)=lower(?) AND ABS(media.year-?)<=?"
	rows := dbh.DB.QueryRow(dBQuery, title, year, tolerance)
	var titleDB, media_type, overview, director string
	var yearDB, id int
	if err := rows.Scan(&titleDB, &yearDB, &id, &media_type, &overview, &director); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Fatal(" Query error: ", err)
		}
		// found no match, check is complete
		return false
	}
	result := api.Media{
		ID:        id,
		MediaType: media_type,
		Director:  director,
		Overview:  overview,
	}
	dateFromYear := strconv.Itoa(yearDB) + "-01-01"
	if media_type == api.MediaTypeTV {
		result.FirstAirDate = dateFromYear
		result.Name = titleDB
	} else {
		result.ReleaseDate = dateFromYear
		result.Title = titleDB
	}
	out := result.String()
	if debug {
		out = result.Dump()
	}
	fmt.Printf("✓ Found %v in DB.\n", out)
	return true
}
func (dbh *DBHandler) WriteToDB(media api.Media, path string) {
	// TODO: handle the case where media.ID is already in the DB
	dbInsert := "INSERT INTO media(id, media_type, title, year, overview, director, poster, path) VALUES(?,?,?,?,?,?,?,?)"
	dbh.DB.Exec(dbInsert, media.ID, media.MediaType, media.GetTitle(), media.GetYear(), media.Overview, media.Director, media.PosterData, path)
}
