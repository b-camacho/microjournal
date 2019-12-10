package api

import (
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

// NewRouter returns an HTTP handler that implements the routes for the API
func NewRouter(store db.PStore, auth auth.Env) http.Handler {
	env := Env{store, auth}

	r := chi.NewRouter()

	authMiddleware := env.auth.RequireAuthentication(
		[]string{"/api/v1/login", "/api/v1/register"},
		func(err error, w http.ResponseWriter) {http.Error(w, err.Error(), 401)},
		)

	r.Use(authMiddleware)
	r.Post("/login", env.PostLogin)
	r.Post("/register", env.PostRegister)

	r.Get("/post", env.GetPosts)
	r.Post("/post", env.CreatePost)

	return r
}

func (env *Env) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := env.auth.HandleLogin(w, r)
	if err != nil {
		log.Printf("failed auth attempt: %s ", err.Error())
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
	}
	return
}

type RegisterPayload struct {
	Email string `json:"email"`
	Password []byte `json:"password"`
}
func (env *Env) PostRegister(w http.ResponseWriter,	r *http.Request) {
	var payload RegisterPayload
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if decoder.Decode(&payload) != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	user, err := env.store.CreateUser(payload.Email, payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	cookie := env.auth.SerialiseUser(user)
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

type PostsResponse struct {
	Posts []db.Post
}

func (env *Env) GetPosts(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	response := PostsResponse{
		Posts: env.store.FindPosts(u.Id),
	}
	jsonResponse(w, response, http.StatusOK)
}

type PostPayload struct {
	Body string `json:"body"`
	Title string `json:"title"`
}

func (env *Env) CreatePost(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	var payload PostPayload
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if decoder.Decode(&payload) != nil {
		http.Error(w, "json body could not be parsed", http.StatusBadRequest)
		return
	}
	err := env.store.CreatePost(u.Id, payload.Body, payload.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

