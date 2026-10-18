package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	flag.Parse()
	url := os.Getenv("NUVRO_DATABASE_URL")
	if url == "" {
		log.Fatal("NUVRO_DATABASE_URL is required")
	}
	database, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatal("database configuration rejected")
	}
	defer func() { _ = database.Close() }()
	if dialectErr := goose.SetDialect("postgres"); dialectErr != nil {
		log.Fatal(dialectErr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	switch *direction {
	case "up":
		err = goose.UpContext(ctx, database, "db/migrations")
	case "down":
		err = goose.DownContext(ctx, database, "db/migrations")
	default:
		log.Fatal("direction must be up or down")
	}
	if err != nil {
		log.Fatal("migration failed; inspect database using authorized operator tooling")
	}
}
