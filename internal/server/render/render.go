package render

import (
	"encoding/json"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Env struct {
	store db.PStore
	auth auth.Env
	template *template.Template
}

type RenderParams struct {
	flash string
}
func ReadTemplate(dir, name string) *template.Template {
	f, err := os.Open(dir + "/components")
	if err != nil { panic(err) }
	compNames, err := f.Readdirnames(0)
	for i, s := range compNames {
		compNames[i] = dir + "/components/" + s
	}
	compNames = append(compNames, dir + "/" + name)
	if err != nil { panic(err) }
	tmpl := template.Must(template.ParseGlob(dir + "/*.tmpl"))
	template.Must(tmpl.ParseGlob(dir + "/**/*.tmpl"))
	return tmpl
}
func (env *Env) renderResponse(w http.ResponseWriter, templateName string, templateData interface{}) {
	err := env.template.ExecuteTemplate(w, templateName, templateData)
	if err != nil {
		log.Fatal(err.Error())
	}
}
// NewRouter returns an HTTP handler that implements the routes for the API
func NewRouter(store db.PStore, auth auth.Env) http.Handler {
	tmpl := ReadTemplate("internal/templates", "home.tmpl")

	env := Env{store, auth, tmpl}

	r := chi.NewRouter()

	authMiddleware := env.auth.RequireAuthentication(
		[]string{"/login", "/register", "/about", "/"},
		func(err error, w http.ResponseWriter) {
			w.WriteHeader(401)
			err = tmpl.ExecuteTemplate(w,"login", RenderParams{"You need to log in before accessing that page"})
			if err != nil {
				log.Println(err.Error())
			}
			},
		)
	r.Use(authMiddleware)
	r.Get("/", env.GetHome)
	r.Get("/entries", env.GetEntries)
	r.Get("/login", env.GetLogin)
	r.Get("/register", env.GetRegister)
	r.Post("/entry", env.PostEntry)
	r.Post("/register", env.PostRegister)
	r.Post("/login", env.PostLogin)
	//r.Post("/login", env.PostLogin)
	//r.Post("/register", env.PostRegister)
	//
	//r.Get("/post", env.GetPosts)
	//r.Post("/post", env.CreatePost)

	return r
}

func (env *Env) GetHome(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "home", "")
}

func (env *Env) GetLogin(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "login", "")
}

func (env *Env) GetRegister(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "register", "")
}

func (env *Env) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := env.auth.HandleLogin(w, r)
	if err != nil {
		log.Printf("failed auth attempt: %s ", err.Error())

		http.Redirect(w, r, "/login", 401)
	}
	http.Redirect(w, r, "/entries", 200)
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
	http.Redirect(w, r, "/entries", 200)
}

type PostsResponse struct {
	Posts []db.Post
}

func (env *Env) GetEntries(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	data := PostsResponse{
		Posts: env.store.FindPosts(u.Id),
	}
	env.renderResponse(w, "entries", data)
}

type PostPayload struct {
	Body string `json:"body"`
	Title string `json:"title"`
}

func (env *Env) PostEntry(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, "/entries", 200)
}

