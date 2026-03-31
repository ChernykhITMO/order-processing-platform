package main

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	var (
		dsn     = flag.String("db", "", "postgres dsn")
		dir     = flag.String("dir", "./migrations", "migrations directory")
		command = flag.String("command", "up", "goose command")
	)
	flag.Parse()

	if *dsn == "" {
		log.Fatal("db dsn is required")
	}

	db, err := sql.Open("pgx", *dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	if err := goose.Run(*command, db, *dir); err != nil {
		log.Fatalf("goose %s failed: %v", *command, err)
	}
}
