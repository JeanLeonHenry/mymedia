package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"database/sql"

	fzf "github.com/junegunn/fzf/src"
	"github.com/profclems/go-dotenv"

	_ "modernc.org/sqlite"
)

func main() {
	// Load config
	dotenv.SetConfigFile(path.Join(os.Getenv("HOME"), ".config/mymedia/config.env"))
	dbPath := dotenv.GetString("DB_PATH")
	if dbPath == "" {
		log.Fatal("db path is empty, check config file.")
	}
	inputChan := make(chan string)
	go func() {
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Fatal("Error opening db file at '", dbPath, "' ", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatal("Error pinging '", dbPath, "' file", err)
		}
		query := "SELECT title, year, overview, director, path FROM media ORDER BY title, year ASC"
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal("Query error : ", err)
		}
		for rows.Next() {
			var title, overview, director, path string
			var year int
			if err := rows.Scan(&title, &year, &overview, &director, &path); err != nil {
				log.Fatal(err)
			}
			if director != "" {
				director = " -- " + director
			}
			s := fmt.Sprintf("%v\t%v (%v)%v\t%v\t%v", title, title, year, director, overview, path)
			inputChan <- s
		}
		close(inputChan)
	}()

	outputChan := make(chan string)
	go func() {
		for s := range outputChan {
			path := strings.FieldsFunc(s, func(r rune) bool { return r == '\t' })[3]
			fmt.Println(path)
		}
	}()

	exit := func(code int, err error) {
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(code)
	}

	// BUG: This preview fails with fish shell because of parentheses.
	// should work fine with bash
	previewCmd := "'echo {2};echo;echo {3}|fold -w ${FZF_PREVIEW_COLUMNS} -s'"
	// Build fzf.Options
	options, err := fzf.ParseOptions(
		true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
		[]string{"--delimiter=\\t", "--with-nth=1", "--preview=" + previewCmd},
	)
	if err != nil {
		exit(fzf.ExitError, err)
	}

	// Set up input and output channels
	options.Input = inputChan
	options.Output = outputChan

	// Run fzf
	code, err := fzf.Run(options)
	exit(code, err)
}
