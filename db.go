package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func openDB() error {
	db, err := sql.Open("sqlite3", "./prisma/dev.db")
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func closeDB() error {
	return DB.Close()
}

func TodoPost() error {
	insertTodo:= "INSERT INTO Todo(todo, done) VALUES(?, ?)"
	_, err := DB.Exec(insertTodo, "todo1", false)
	if err != nil {
		return err
	}
	return nil
}
