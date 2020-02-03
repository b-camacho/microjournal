package server

import (
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/config"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/b-camacho/microjournal/internal/server/api"
	"github.com/b-camacho/microjournal/internal/server/render"
	"net/http"
	"path/filepath"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// NewRouter returns a new HTTP handler that implements the main server routes
func NewRouter(conf config.Config, authProvider auth.Env, store db.PStore) http.Handler {
	router := chi.NewRouter()

	// Set up our middleware with sane defaults
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.DefaultCompress)
	router.Use(middleware.Timeout(60 * time.Second))

	// Set up static file serving
	staticPath, _ := filepath.Abs("internal/static")
	fmt.Println(staticPath)
	fs := http.FileServer(http.Dir(staticPath))
	router.Handle("/static*", http.StripPrefix("/static/", fs))

	// Set up rendered site handlers
	router.Mount("/", render.NewRouter(store, authProvider))

	// Set up REST API
	router.Mount("/api/v1/", api.NewRouter(store, authProvider))

	return router
}