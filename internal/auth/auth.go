package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
)

const CookieName = "session"

type Env struct {
	store db.PStore
	s     *securecookie.SecureCookie
}

func (env *Env) AuthenticateUser(email, password string) (*db.User, error) {
	user, err := env.store.FindUser("email", email)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("No user with email %s", email))
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Incorrect password"))
	}
	return user, nil
}

func (env *Env) SerialiseUser(user *db.User) *http.Cookie {
	encoded, err := env.s.Encode(CookieName, user.Id)
	if err != nil {
		log.Fatal(err.Error())
	}
	cookie := &http.Cookie{
		Name:  CookieName,
		Value: encoded,
		Path:  "/",
	}
	return cookie
}

func (env *Env) DeserialiseUser(cookie *http.Cookie) (*db.User, error) {
	uid := 0
	err := env.s.Decode(CookieName, cookie.Value, &uid)
	if err != nil {
		return nil, err
	}
	return env.store.FindUser("id", uid)
}

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginValidate(payload *LoginPayload) bool {
	if strings.Index(payload.Email, "@") == -1 {
		return false // TODO: actual email regex match
	}
	return true
}

func (env *Env) HandleLogin(w http.ResponseWriter, r *http.Request) {
	payload := LoginPayload{}
	decoder := json.NewDecoder(r.Body)
	if decoder.Decode(&payload) != nil || !LoginValidate(&payload) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := env.AuthenticateUser(payload.Email, payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	cookie := env.SerialiseUser(user)

	http.SetCookie(w, cookie)

	w.Write([]byte("ok"))

}

func Init(store db.PStore, blockKey, hashKey []byte) Env {
	sc := securecookie.New(hashKey, blockKey)
	return Env{store, sc}
}
