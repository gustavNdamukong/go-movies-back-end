package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create a router mux (multiplexer)
	mux := chi.NewRouter()

	// here is where you define your routes & any middleware

	// the Recoverer logs issues with a debug trace, if the server runs into errors,
	// then return the appropriate 500 error & keeps the server running so your app
	// doesn't grind to a halt
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	// NOTES: we say app.Home coz Home() is a receiver func of the application struct, witten in 'cmd/api/handlers.go'
	mux.Get("/", app.Home)

	mux.Post("/authenticate", app.authenticate)
	mux.Get("/refresh", app.refreshToken)
	mux.Get("/logout", app.logout)

	mux.Get("/movies", app.AllMovies)
	mux.Get("/movies/{id}", app.GetMovie)

	mux.Get("/genres", app.AllGenres)
	mux.Get("/movies/genres/{id}", app.AllMoviesByGenre)

	// Note that we will have only one route for GraphQL queries
	mux.Post("/graph", app.MoviesGraphQL)

	// restrict the app.authRequired token access validation to "/admin" routes
	mux.Route("/admin", func(mux chi.Router) {
		mux.Use(app.authRequired)

		mux.Get("/movies", app.MovieCatalog)
		mux.Get("/movies/{id}", app.MovieForEdit)
		mux.Put("/movies/0", app.InsertMovie)
		mux.Patch("/movies/{id}", app.UpdateMovie)
		mux.Delete("/movies/{id}", app.DeleteMovie)
	})

	return mux
}
