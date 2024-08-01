package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"database/sql"

	fzf "github.com/junegunn/fzf/src"

	_ "modernc.org/sqlite"
)

func main() {
	inputChan := make(chan string)
	go func() {
		db, err := sql.Open("sqlite", "/media/jean-leon/MyData/Videos/media.db")
		if err != nil {
			log.Fatal(err)
		}
		query := "SELECT title, year, overview, director, path FROM media ORDER BY title, year ASC"
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		for rows.Next() {
			var title, overview, director, path string
			var year int
			if err := rows.Scan(&title, &year, &overview, &director, &path); err != nil {
				log.Fatal(err)
			}
			// title, year, overview, director, path := res[0], res[1], res[2], res[3], res[4]
			// if director != "" {
			// 	director = " -- " + director
			// }
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

	// Build fzf.Options
	options, err := fzf.ParseOptions(
		true, // whether to load defaults ($FZF_DEFAULT_OPTS_FILE and $FZF_DEFAULT_OPTS)
		[]string{"-d=\t", "--with-nth=1"},
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
