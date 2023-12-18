package main

import (
	"database/sql"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func openDB() error {
	db, err := sql.Open("postgres", "postgresql://postgres:postgres@postgres:5432/postgres?sslmode=disable")
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func closeDB() error {
	return DB.Close()
}
