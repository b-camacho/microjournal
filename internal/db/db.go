package db

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

type PStore interface {
	FindUser(string, interface{}) (*User, error)
	CreateUser(string, []byte) (*User, error)
}

func Init(connStr string) DB {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	DB := DB{conn}
	return DB
}

func (db *DB) FindUser(colName string, colValue interface{}) (*User, error) {
	query := `SELECT id, email, password, created_at, updated_at, deleted_at
			FROM users WHERE $1 = $2`

	var u User
	notFound := db.conn.
		QueryRow(query, colName, colValue).
		Scan(&u.Id, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if notFound != nil {
		return nil, notFound
	} else {
		return &u, nil
	}
}

func (db *DB) CreateUser(email string, password []byte) (*User, error) {
	if _, missing := db.FindUser("email", email); missing != nil {
		return nil, errors.New("This email address is taken")
	}
	hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	statement := `INSERT INTO users(email, password)
	VALUES($1, $2)`
	_, err = db.conn.Exec(statement, email, hash)
	if err != nil {
		return nil, err
	}
	return db.FindUser("email", email)
}
