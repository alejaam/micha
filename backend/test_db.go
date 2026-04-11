package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://micha:micha_dev_password@localhost:5432/micha")
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	rows, err := pool.Query(ctx, "SELECT id, household_id, name, slug FROM categories LIMIT 10")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, hid, name, slug string
		if err := rows.Scan(&id, &hid, &name, &slug); err != nil {
			log.Fatalf("Scan failed: %v\n", err)
		}
		fmt.Printf("Cat: %s, Household: %s, Name: %s, Slug: %s\n", id, hid, name, slug)
	}

	hrows, err := pool.Query(ctx, "SELECT id, name FROM households LIMIT 10")
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer hrows.Close()

	for hrows.Next() {
		var id, name string
		if err := hrows.Scan(&id, &name); err != nil {
			log.Fatalf("Scan failed: %v\n", err)
		}
		fmt.Printf("Household: %s, Name: %s\n", id, name)
	}
}
