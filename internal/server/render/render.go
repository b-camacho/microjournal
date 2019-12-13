package render

import (
	"encoding/json"
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/go-chi/chi"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type Env struct {
	store    db.PStore
	auth     auth.Env
	template *template.Template
}

type RenderParams struct {
	flash string
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

	env := Env{store, auth, tmpl}

	r := chi.NewRouter()

	//authMiddleware := env.auth.RequireAuthentication(
	//	[]string{"/login", "/register", "/about", "/"},
	//	func(err error, w http.ResponseWriter) {
	//		w.WriteHeader(401)
	//		err = tmpl.ExecuteTemplate(w, "login", RenderParams{"You need to log in before accessing that page"})
	//		if err != nil {
	//			log.Println(err.Error())
	//		}
	//	},
	//)
	//r.Use(authMiddleware)
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
	err := r.ParseForm()
	if err != nil {
		env.template.ExecuteTemplate(w, "login", RenderParams{"log in failed"})
	}
	email, password := r.PostForm.Get("email"), r.PostForm.Get("password")
	user, err := env.auth.AuthenticateUser(email, password)
	if err != nil {
		log.Printf("failed auth attempt: %s ", err.Error())
		w.WriteHeader(401)
		env.template.ExecuteTemplate(w, "login", RenderParams{"incorrect username or password"})
		return
	}

	cookie := env.auth.SerialiseUser(user)
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/entries", 200)
}

type RegisterPayload struct {
	Email    string `json:"email"`
	Password []byte `json:"password"`
}

func (env *Env) PostRegister(w http.ResponseWriter, r *http.Request) {
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

type EntriesResp struct {
	Entries []Entry
	DayIdx int
}

func (env *Env) GetEntries(w http.ResponseWriter, r *http.Request) {
	//u := r.Context().Value("user").(*db.User)
	//data := PostsResponse{
	//	Posts: env.store.FindPosts(u.Id),
	//}
	env.renderResponse(w, "entries", EntriesResp{[]Entry{
		{
			Created_at: "yesterday",
			Title:  "Lorem Ipsum",
			Body:   "Illo corrupti sint perferendis. Illum voluptatem nobis qui. Facilis exercitationem est sapiente nihil aut ipsum. Omnis est nisi dicta et nesciunt iusto.",
		}, {
			Created_at: "just now",
			Title:  "Lorem Ipsum 2",
			Body:   ` sunt sunt fugit eius sint numquam eos ad earum facilis quis enim non officia animi tempore quia enim fuga quos atque reiciendis dolor quia itaque voluptatem quas velit quo voluptas voluptatem soluta voluptate accusantium incidunt esse quo ut quibusdam quas tempora possimus perspiciatis omnis natus accusantium est quaerat ab eveniet `,
		},
	},
	2})
}

type PostPayload struct {
	Body  string `json:"body"`
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
