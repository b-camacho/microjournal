package test

import (
	"encoding/hex"
	"fmt"
	"github.com/b-camacho/microjournal/internal/auth"
	"github.com/b-camacho/microjournal/internal/db"
	"github.com/b-camacho/microjournal/test/mocks"
	"golang.org/x/crypto/bcrypt"
	"os"
	"testing"
)

var mockStore mocks.MockPStore
var mockEnv auth.Env
func TestMain(m *testing.M) {
	mockStore = mocks.MockPStore{Trace: []string{}}
	blockkey, _ := hex.DecodeString("30b8a43be0864ee32393f13d2a9fb3a46f868c8b8ae77f5c24e4ceaab4bd5819")
	hashkey, _ := hex.DecodeString("2d511f249902c7c5c9a96a568bdc61b10d0d3f85a24ea3dd20fbc10f9b153fda")
	mockEnv = auth.Init(&mockStore,
		blockkey,
		hashkey,
	)

	os.Exit(m.Run())
}

func TestAuthenticateUser(t *testing.T) {
	var usr = "usr"
	var pwd = "pwd"
	var pwdhash, _ = bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)

	mockStore.UserMock = &db.User{
		Email:"",
		Password: pwdhash,
	}

	res, err := mockEnv.AuthenticateUser(usr, pwd)
	if res != mockStore.UserMock {
		t.Error(res, "!=", mockStore.UserMock)
	}
	if err != nil {
		t.Error("Expected nil, got ", err)
	}

	res, err = mockEnv.AuthenticateUser(usr, "wrongpwd")
	if res != nil {
		t.Error("expected nil got", res)
	}

	if err == nil || err.Error() != "incorrect password" {
		t.Error("expected error got", err)
	}

	mockStore.UserMock = nil
	mockStore.ErrorMock = fmt.Errorf("err")

	res, err = mockEnv.AuthenticateUser(usr, pwd)
	if res != nil {
		t.Error(res, "should be nil for email mismatch")
	}
	if err == nil {
		t.Error(err, "should not be nil for email mismatch")
	}
}

func TestSerializeUser(t *testing.T) {
	user := &db.User{Model: db.Model{
		Id: 1,
	}}
	lenBefore := len(mockStore.Trace)
	mockEnv.SerializeUser(user)
	if len(mockStore.Trace) != lenBefore + 1 {
		t.Error("Trace should be extended by one, instead changed from ",
			lenBefore, " to ", len(mockStore.Trace))
	}
	if len(mockStore.Trace) == 0 || mockStore.Trace[len(mockStore.Trace) - 1] != fmt.Sprint("CreateSession", user.Id) {
		t.Error("most recent trace item should be ", fmt.Sprint("CreateSession", user.Id),
			"but wasn't")
	}
}

func TestDeserializeUser(t *testing.T) {
	testCookie := mockEnv.SerializeUser(&db.User{})
	mockStore.BoolMock = true
	_, err := mockEnv.DeserializeUser(testCookie)
	if err != nil {
		t.Error(err)
	}
	mockStore.BoolMock = false
	_, err = mockEnv.DeserializeUser(testCookie)
	if err == nil {
		t.Error("err was nil, expected \"this session expired\"")
	}
}

