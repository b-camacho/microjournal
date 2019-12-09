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
	u := r.Context().Value("user").(db.User)
	response := HelloResponse{
		Message: fmt.Sprintf("Hello %s!", u.Email),
	}
	jsonResponse(w, response, http.StatusOK)
}


func (env *Env) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, cookieErr := r.Cookie(auth.CookieName)
		user, deserErr := env.auth.DeserialiseUser(cookie)
		if cookieErr != nil || deserErr != nil {
			http.Error(w, "Auth Failed", http.StatusUnauthorized)
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

	r.Post("/login", auth.HandleLogin())

	// Register the API routes
	r.Get("/", HelloWorld)
	r.Get("/email", HelloName)

	return r
}
