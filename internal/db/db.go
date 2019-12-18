package db

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"

	_ "github.com/lib/pq"
)

type PStore interface {
	FindUser(string, interface{}) (*User, error)
	CreateUser(string, []byte) (*User, error)
	FindPosts(int, int, int) ([]*Post, int)
	CreatePost(int, string, string) error
	CreateSession(int) error
	ExistsSession(int) bool
	InvalidateSession(int) error
}

type DB struct {
	conn *sql.DB
}

func Init(connStr string) PStore {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = conn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	DB := DB{conn}
	return &DB
}

func (db *DB) FindUser(colName string, colValue interface{}) (*User, error) {
	if colName != "id" && colName != "email" {
		return nil, errors.New(fmt.Sprintf("not a valid columnd name: %s", colName))
	}

	query := fmt.Sprintf(`SELECT id, email, password, created_at, updated_at, deleted_at
			FROM users WHERE %s = $1`, colName)

	var u User
	var deleted_at sql.NullTime
	notFound := db.conn.
		QueryRow(query, colValue).
		Scan(&u.Id, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt, &deleted_at)

	if notFound != nil || deleted_at.Valid {
		return nil, notFound
	} else {
		return &u, nil
	}
}

func (db *DB) CreateUser(email string, password []byte) (*User, error) {
	if _, missing := db.FindUser("email", email); missing == nil {
		return nil, errors.New("This email address is taken")
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err.Error())
		return nil, fmt.Errorf("Could not create user")
	}

	statement := `INSERT INTO users(email, password) VALUES($1, $2)`
	_, err = db.conn.Exec(statement, email, hash)
	if err != nil {
		return nil, err
	}
	return db.FindUser("email", email)
}

func (db *DB) FindPosts(userId, offset, limit int) ([]*Post, int) {
	query := fmt.Sprintf("SELECT id, title, body, created_at, updated_at, deleted_at FROM POSTS " +
		"WHERE user_id = $1 AND deleted_at IS NULL " +
		"ORDER BY created_at DESC " +
		"LIMIT %d OFFSET %d", limit, offset)

	postRows, err := db.conn.
		Query(query, userId)
	if err != nil {
		log.Fatal(err.Error())
		return []*Post{}, 0
	}

	var rowCnt int
	err = db.conn.
		QueryRow(`SELECT COUNT(1) FROM POSTS WHERE user_id = $1 AND deleted_at IS NULL`, userId).
		Scan(&rowCnt)
	if err != nil {
		log.Fatal(err.Error())
		return []*Post{}, 0
	}

	posts := make([]*Post, 0)
	defer postRows.Close()

	for postRows.Next() {
		var post Post
		var deleted_at sql.NullTime
		err = postRows.Scan(&post.Id, &post.Title, &post.Body, &post.CreatedAt, &post.UpdatedAt, &deleted_at)
		if err == nil || !deleted_at.Valid {
			posts = append(posts, &post)
		}
		if err != nil {
			log.Println(err.Error())
		}
	}


	return posts, rowCnt
}

func (db *DB) CreatePost(userId int, title, body string) error {
	_, err := db.conn.
		Exec(`INSERT INTO posts VALUES (DEFAULT, $1, $2, $3)`, userId, title, body)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (db *DB) CreateSession(userId int) error {
	_, err := db.conn.
		Exec("INSERT INTO sessions VALUES " +
			"(DEFAULT, $1, DEFAULT, DEFAULT, DEFAULT) " +
			"ON CONFLICT (user_id) DO UPDATE SET updated_at = now(), valid = true", userId)
	return err
}

func (db *DB) ExistsSession(userId int) bool {
	var hasSession int

	db.conn.
		QueryRow("SELECT COUNT(1) FROM sessions WHERE user_id = $1 AND valid = TRUE", userId).Scan(&hasSession)
	if hasSession == 1 {
		return true
	}
	return false
}

func (db *DB) InvalidateSession(userId int) error {
	_, err := db.conn.
		Exec("UPDATE sessions SET valid = false WHERE user_id = $1", userId)
	return err
}

