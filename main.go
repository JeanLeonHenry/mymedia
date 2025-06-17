package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"

	"github.com/JeanLeonHenry/mymedia/cmd"
	"github.com/JeanLeonHenry/mymedia/config"
	"github.com/JeanLeonHenry/mymedia/db"
)

//go:embed schema.sql
var ddl string

func main() {
	cfg := config.New()
	cfg.Check()
	if !cfg.IsValid {
		log.Fatalf("Config is not valid.\nConfig was : %+v", cfg)
	}

	ctx := context.Background()

	// INFO: will create the file if it doesn't exist
	DB, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		log.Fatalln(err)
	}

	// INFO: will create the tables if they don't exist
	if _, err := DB.ExecContext(ctx, ddl); err != nil {
		log.Println(err)
	}
	queries := db.New(DB)
	cmd.InitCLI(queries, ctx, cfg)
	cmd.Execute()
}
