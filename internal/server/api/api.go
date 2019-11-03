package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/b-camacho/microjournal/internal/server"
	"github.com/gorilla/securecookie"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

type Env struct {
	store *db.PStore
	auth *auth.Env
	s *securecookie.SecureCookie
}

// ValidBearer is a hardcoded bearer token for demonstration purposes.
const ValidBearer = "123456"

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

// HelloWorld returns a basic "Hello World!" message
func HelloWorld(w http.ResponseWriter, r *http.Request) {
	response := HelloResponse{
		Message: "Hello world!",
	}
	jsonResponse(w, response, http.StatusOK)
}

// HelloName returns a personalized JSON message
func HelloName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	response := HelloResponse{
		Message: fmt.Sprintf("Hello %s!", name),
	}
	jsonResponse(w, response, http.StatusOK)
}

// RequireAuthentication is an example middleware handler that checks for a
// hardcoded bearer token. This can be used to verify session cookies, JWTs
// and more.
func (env *Env) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			server.GenericError(w, r)
			return
		}
		uid := 0
		err = env.s.Decode(auth.CookieName, cookie.Value, &uid)
		if err != nil {
			server.GenericError(w, r)
			return
		}
		user, err := env.auth.DeserialiseUser(uid)
		if err != nil {
			server.GenericError(w, r)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "user", user))

		next.ServeHTTP(w, r)
	})
}

// NewRouter returns an HTTP handler that implements the routes for the API
func NewRouter(store *db.PStore, auth *auth.Env, s *securecookie.SecureCookie) http.Handler {
	env := Env{store, auth, s}

	r := chi.NewRouter()

	r.Use(env.RequireAuthentication)

	// Register the API routes
	r.Get("/", HelloWorld)
	r.Get("/{name}", HelloName)
	r.Post("/login", auth.HandleLogin())
	return r
}
