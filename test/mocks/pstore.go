package mocks

import "fmt"
import "github.com/b-camacho/microjournal/internal/db"

type MockPStore struct {
	Trace []string
	UserMock *db.User
	ErrorMock error
	PostMock []*db.Post
	IntMock int
	BoolMock bool
}

func (m *MockPStore) FindUser(arg1 string, arg2 interface{}) (*db.User, error) {
	m.Trace = append(m.Trace, fmt.Sprint("FindUser", arg1, arg2))
	return m.UserMock, m.ErrorMock
}

func (m *MockPStore) CreateUser(arg1 string, arg2 []byte) (*db.User, error) {
	m.Trace = append(m.Trace, fmt.Sprint("CreateUser", arg1, arg2))
	return m.UserMock, m.ErrorMock
}

func (m *MockPStore) DeleteUser(arg1 int) error {
	m.Trace = append(m.Trace, fmt.Sprint("DeleteUser", arg1))
	return m.ErrorMock
}

func (m *MockPStore) FindPosts(arg1 int, arg2 int, arg3 int) ([]*db.Post, int) {
	m.Trace = append(m.Trace, fmt.Sprint("FindPosts", arg1, arg2, arg3))
	return m.PostMock, m.IntMock
}

func (m *MockPStore) CreatePost(arg1 int, arg2 string, arg3 string) error {
	m.Trace = append(m.Trace, fmt.Sprint("CreatePost", arg1, arg2, arg3))
	return m.ErrorMock
}

func (m *MockPStore) DeletePost(arg1 int) error {
	m.Trace = append(m.Trace, fmt.Sprint("DeletePost", arg1))
	return m.ErrorMock
}

func (m *MockPStore) CreateSession(arg1 int) error {
	m.Trace = append(m.Trace, fmt.Sprint("CreateSession", arg1))
	return m.ErrorMock
}

func (m *MockPStore) ExistsSession(arg1 int) bool {
	m.Trace = append(m.Trace, fmt.Sprint("ExistsSession", arg1))
	return m.BoolMock
}

func (m *MockPStore) InvalidateSession(arg1 int) error {
	m.Trace = append(m.Trace, fmt.Sprint("InvalidateSession", arg1))
	return m.ErrorMock
}

