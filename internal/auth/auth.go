package auth

import (
	"errors"
	"fmt"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/b-camacho/microjournal/internal/server"
	"github.com/b-camacho/microjournal/internal/server/middleware"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

const CookieName = "session"

type Env struct {
	store db.PStore
	s securecookie.SecureCookie
}

func (env *Env) AuthenticateUser(email, password string) (*db.User, error) {
	user, err := env.store.FindUser("email", email)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("No user with email %s", email))
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Incorrect password", email))
	}
	return user, nil
}

func (env *Env) SerialiseUser(user db.User) int {
	return user.Id
}

func (env *Env) DeserialiseUser(id int) (*db.User, error) {
	return env.store.FindUser("id", id)
}


type LoginPayload struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func LoginValidate(payload interface{}) bool {
	loginPayload, ok := payload.(LoginPayload)
	if !ok || strings.Index(loginPayload.Email, "@") == -1 {
		return false // TODO: actual email regex match
	}
	return true
}

func (env *Env) HandleLogin() http.HandlerFunc {
	payload := LoginPayload{}
	fn := func(w http.ResponseWriter, r *http.Request) {
		body := r.Context().Value(middleware.CtxBody).(LoginPayload)
		user, err := env.AuthenticateUser(body.Email, body.Password)
		if err != nil {
			server.GenericError(w, r)
			return
		}
		id := env.SerialiseUser(*user)
		encoded, _ := env.s.Encode(CookieName, id)
		cookie := &http.Cookie{
			Name:  CookieName,
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, cookie)

		w.Write([]byte("ok"))
	}

	return middleware.ParseBody(http.HandlerFunc(fn), payload, LoginValidate).ServeHTTP
}




