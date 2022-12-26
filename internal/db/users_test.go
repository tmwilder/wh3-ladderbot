package db

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	// TODO figure out how to do once-per-suite setup
	rand.Seed(time.Now().UnixNano())

	testEmail := fmt.Sprintf("test%d@test.com", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	createUser(conn, User{testDiscordId, testEmail})

	user := getUser(conn, testEmail)

	if user.Email != testEmail {
		t.Error("Unable to create test user: " + testEmail)
	}
}
