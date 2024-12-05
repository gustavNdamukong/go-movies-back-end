package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dbrepo"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = 8080

type application struct {
	DSN    string
	Domain string
	DB     repository.DatabaseRepo // NOTES: this should ideally be a pool of possibly different DB connections, hence we
	// make it an interface. Now, every entity assigned to DB must fullfil the requirements
	// of the repository.DatabaseRepo interface. This is how interfaces work in Go.
	// To use a different type of DB, just change the value u assign to DB below to the required STRUCT that must
	// meet the requirements of the repository.DatabaseRepo interface.
	auth         Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
	APIKey       string
}

func main() {
	// set application config
	var app application

	// create command line flags (to read from CLI)
	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=movies sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.StringVar(&app.JWTSecret, "jwt-secret", "verysecret", "signing secret")
	flag.StringVar(&app.JWTIssuer, "jwt-issuer", "example.com", "signing issuer")
	flag.StringVar(&app.JWTAudience, "jwt-audience", "example.com", "signing audience")
	flag.StringVar(&app.CookieDomain, "cookie-domain", "localhost", "cookie domain")
	flag.StringVar(&app.Domain, "domain", "example.com", "domain")
	flag.StringVar(&app.APIKey, "api-key", "12bf99076cbd076fc0c74b70b2b1000a", "api key")
	flag.Parse()

	// connect to DB
	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	// NOTES: this is a handy way to assign a struct to a var (or in this case another struct property)
	//	while updating its value at the same time
	app.DB = &dbrepo.PostgresDBRepo{DB: conn}
	// NOTES: defer means: 'run this defer line only when the function in which defer is (in this case main()) ends'
	// this is coz we dont wanna close the DB session when we are still using it.
	defer app.DB.Connection().Close() // This will also work: defer conn.Close()

	// populated the app.auth key with the right JWT auth stuff
	app.auth = Auth{
		Issuer:   app.JWTIssuer,
		Audience: app.JWTAudience,
		Secret:   app.JWTSecret,
		// we want the token to initially expire every 15 minutes, so the user will be logged out,
		// unless we refresh the token-which we will.
		TokenExpiry:   time.Minute * 15,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		/////CookieName:    "__Host-refresh_token",
		CookieName:   "gus_refresh_token",
		CookieDomain: app.CookieDomain,
	}

	log.Println("Running application on port", port)

	// start a web server
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}

}
