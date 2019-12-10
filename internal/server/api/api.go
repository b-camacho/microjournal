package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

type Env struct {
	store db.PStore
	auth auth.Env
}

// HelloResponse is the JSON representation for a customized message
type HelloResponse struct {
	Message string `json:"message"`
}

func jsonResponse(w http.ResponseWriter, data interface{}, c int) {
	dj, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	fmt.Fprintf(w, "%s", dj)
}

// HelloName returns a personalized JSON message
func HelloName(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	response := HelloResponse{
		Message: fmt.Sprintf("Hello %s!", u.Email),
	}
	jsonResponse(w, response, http.StatusOK)
}


func (env *Env) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/login" || r.URL.Path == "/api/v1/register"{ // these are exempt from auth
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		user, err := env.auth.DeserialiseUser(cookie)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		next.ServeHTTP(w, r)
	})
}

// NewRouter returns an HTTP handler that implements the routes for the API
func NewRouter(store db.PStore, auth auth.Env) http.Handler {
	env := Env{store, auth}

	r := chi.NewRouter()

	r.Use(env.RequireAuthentication)
	r.Post("/login", auth.HandleLogin)
	r.Get("/email", HelloName)

	return r
}
