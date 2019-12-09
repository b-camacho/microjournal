package db

import "time"

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	Id int
}

type User = struct {
	*Model
	Email    string
	Password []byte
}

type Post = struct {
	*Model
	UserId int
	title string
	body string
}
