package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// set up an app
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.addIPToContext)
	// middleware to load and save session
	mux.Use(app.Session.LoadAndSave)

	// register routes
	mux.Get("/", app.Home)
	mux.Post("/login", app.Login)

	// any access to /user/* requires authentication
	mux.Route("/user", func(mux chi.Router) {
		mux.Use(app.Auth)
		mux.Get("/profile", app.Profile)
		mux.Post("/upload-profile-pic", app.UploadProfilePic)
	})

	// static assets
	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
