package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openDB(dsn string) (*sql.DB, error) {
	// NOTES: Whats in the DSN is:
	/*
		host=localhost
		port=5432
		user=postgres
		password=postgres
		dbname=movies
		sslmode=disable
		timezone=UTC connect_timeout=5
	*/

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *application) connectToDB() (*sql.DB, error) {
	connection, err := openDB(app.DSN)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Postgres DB")
	return connection, nil
}
