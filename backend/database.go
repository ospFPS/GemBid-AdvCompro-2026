package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var DB *pgxpool.Pool

func ConnectDatabase() {
	// Load .env from project root when running inside /backend
	_ = godotenv.Load("../.env")

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)

	if err != nil {
		log.Fatal(
			"Unable to create database pool: ",
			err,
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatal(
			"Unable to connect to Supabase: ",
			err,
		)
	}

	DB = db

	log.Println(
		"Connected to Supabase PostgreSQL",
	)
}