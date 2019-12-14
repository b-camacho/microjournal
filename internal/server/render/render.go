package render

import (
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Env struct {
	store    db.PStore
	auth     auth.Env
	template *template.Template
	perPage int
}

type RenderParams struct {
	LoggedIn bool
	Flash string
}

func parseTemplates(dir, name string) *template.Template {
	f, err := os.Open(dir + "/components")
	if err != nil {
		panic(err)
	}
	compNames, err := f.Readdirnames(0)
	for i, s := range compNames {
		compNames[i] = dir + "/components/" + s
	}
	compNames = append(compNames, dir+"/"+name)
	if err != nil {
		panic(err)
	}
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
	tmpl := parseTemplates("internal/templates", "home.tmpl")

	env := Env{store, auth, tmpl, 100}

	r := chi.NewRouter()

	authMiddleware := env.auth.RequireAuthentication(
		[]string{"/login", "/register", "/about", "/"},
		func(err error, w http.ResponseWriter) {
			w.WriteHeader(401)
			err = tmpl.ExecuteTemplate(w, "login", RenderParams{Flash: "You need to sign in before accessing that page"})
			if err != nil {
				log.Println(err.Error())
			}
		},
	)

	r.Use(authMiddleware)
	r.Get("/", env.GetHome)
	r.Get("/register", env.GetRegister)
	r.Get("/login", env.GetLogin)
	r.Get("/entries", env.GetEntries)
	//r.Get("/register", env.GetRegister)
	r.Post("/entry", env.PostEntry)
	r.Post("/register", env.PostRegister)
	r.Post("/login", env.PostLogin)

	r.Get("/logout", env.GetLogout)
	//r.Post("/login", env.PostLogin)
	//r.Post("/register", env.PostRegister)
	//
	//r.Get("/post", env.GetPosts)
	//r.Post("/post", env.CreatePost)

	return r
}

func (env *Env) GetHome(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "login", RenderParams{false, ""})
}

func (env *Env) GetLogin(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "login", RenderParams{false, ""})
}

func (env *Env) GetRegister(w http.ResponseWriter, r *http.Request) {
	env.renderResponse(w, "register", RenderParams{false, ""})
}

func (env *Env) PostLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		env.template.ExecuteTemplate(w, "login", RenderParams{Flash:"Sign in failed, please try again."})
	}
	email, password := r.PostForm.Get("email"), r.PostForm.Get("password")
	user, err := env.auth.AuthenticateUser(email, password)
	if err != nil {
		log.Printf("failed auth attempt: %s ", err.Error())
		w.WriteHeader(401)
		env.template.ExecuteTemplate(w, "login", RenderParams{Flash:"The username or password are not correct."})
		return
	}

	cookie := env.auth.SerialiseUser(user)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/entries", http.StatusSeeOther)
}

type RegisterPayload struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func (env *Env) PostRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		env.template.ExecuteTemplate(w, "login", RenderParams{Flash:"Sign in failed, please try again."})
		return
	}
	email, password := r.PostForm.Get("email"), r.PostForm.Get("password")
	user, err := env.store.CreateUser(email, []byte(password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	cookie := env.auth.SerialiseUser(user)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/entries", http.StatusSeeOther)
}

func (env *Env) GetEntries(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	pageNoStr := r.URL.Query().Get("page")
	pageNo := 0
	if pageNoStr != "" {
		var err error
		pageNo, err = strconv.Atoi(pageNoStr)
		if err != nil {
			env.renderResponse(w, "error", RenderParams{true, "An error has occured"})
			return
		}
	}

	posts, postCnt := env.store.FindPosts(u.Id, pageNo * env.perPage, (pageNo + 1) * env.perPage)
	entries := make([]Entry, 0)
	for _, post := range posts {
		entries = append(entries, toEntry(post))
	}

	pageCnt := postCnt / env.perPage
	if postCnt % env.perPage != 0 {
		pageCnt += 1
	}
	pageNumbers := make([]Page, pageCnt)
	for i := range pageNumbers {
		pageNumbers[i] = Page{i + 1, i == pageNo}
	}
	data := EntriesResp{
		Entries: entries,
		DayIdx: len(entries),
		Pages: pageNumbers,
	}
	env.renderResponse(w, "entries", data)
}

func (env *Env) PostEntry(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user").(*db.User)
	err := r.ParseForm()
	if err != nil {
		env.template.ExecuteTemplate(w, "login", RenderParams{true,"Could not create post"})
		return
	}
	err = env.store.CreatePost(u.Id, r.PostForm.Get("title"), r.PostForm.Get("body"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/entries", http.StatusSeeOther)
}

func (env *Env) GetLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:  auth.CookieName,
		Value: "",
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func niceDuration(elapsed time.Duration) string {
	if elapsed.Hours() < 1 {
		if elapsed.Minutes() < 1 {
			return "just now"
		}
		if elapsed.Minutes() < 2 {
			return "a minute ago"
		}
		return fmt.Sprintf("%d minutes ago", int(math.Round(elapsed.Minutes())))
	}
	if elapsed.Hours() < 24 {
		if elapsed.Hours() < 2 {
			return "an hour ago"
		}
		return fmt.Sprint("%d hours ago", elapsed.Hours())
	}
	if elapsed.Hours() / 24 < 2 {
		return "yesterday"
	}
	return fmt.Sprintf("%d days ago", int(math.Round(elapsed.Hours() / 24)))
}

func niceElapsed(from time.Time) string {
	now := time.Now()
	elapsed := now.Sub(from)
	return niceDuration(elapsed)
}

func toEntry(post *db.Post) Entry {
	friendlyDate := niceElapsed(post.CreatedAt)
	return Entry{
		Created_at: friendlyDate,
		Title:      post.Title,
		Body:       post.Body,
	}
}

type Entry struct {
	Created_at string
	Title string
	Body string
}
type Page struct {
	Idx int
	Current bool
}
type EntriesResp struct {
	Entries []Entry
	DayIdx int
	Pages []Page
}
