package gembid

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func isPGError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

func ConnectDatabase() {
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), databaseURL)

	if err != nil {
		log.Fatal("Unable to create database pool: ", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("Unable to connect to Supabase: ", err)
	}

	DB = db

	log.Println("Connected to Supabase PostgreSQL")
}
