package auth

import (
	"context"
	"fmt"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

const CookieName = "session"

type Env struct {
	store db.PStore
	s     *securecookie.SecureCookie
}

const SessionMaxAge = 315576000

func (env *Env) AuthenticateUser(email, password string) (*db.User, error) {
	user, err := env.store.FindUser("email", email)
	if err != nil {
		return nil, fmt.Errorf("no user with email %s", email)
	}
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		return nil, fmt.Errorf("incorrect password")
	}
	return user, nil
}

func (env *Env) SerializeUser(user *db.User) *http.Cookie {
	encoded, err := env.s.Encode(CookieName, user.Id)
	if err != nil {
		log.Println(err.Error())
	}
	cookie := &http.Cookie{
		Name:  CookieName,
		Value: encoded,
		Path:  "/",
		MaxAge: SessionMaxAge,
	}
	err = env.store.CreateSession(user.Id)
	if err != nil {
		log.Println(err.Error())
	}
	return cookie
}

func (env *Env) DeserializeUser(cookie *http.Cookie) (*db.User, error) {
	var uid int
	err := env.s.Decode(CookieName, cookie.Value, &uid)
	if err != nil {
		return nil, err
	}
	valid := env.store.ExistsSession(uid)
	if !valid {
		return nil, fmt.Errorf("this session expired")
	}
	return env.store.FindUser("id", uid)
}

func (env *Env) RequireAuthentication(
	exclusions []string,
	onerr func(error, http.ResponseWriter)) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var excluded bool
			for _, excl := range exclusions { // these are exempt from auth
				if r.URL.Path == excl {
					excluded = true
				}
			}
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				if excluded {
					next.ServeHTTP(w, r)
					return
				}
				onerr(err, w)
				return
			}
			user, err := env.DeserializeUser(cookie)
			if err != nil {
				if excluded {
					next.ServeHTTP(w, r)
					return
				}
				onerr(err, w)
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "user", user))
			next.ServeHTTP(w, r)
		})
	}
}

func Init(store db.PStore, blockKey, hashKey []byte) Env {
	sc := securecookie.New(hashKey, blockKey)
	return Env{store, sc}
}
