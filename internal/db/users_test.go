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

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	createUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})

	user := getUser(conn, testDiscordId)

	if user.DiscordUserName != testDiscordUsername {
		t.Error("Unable to create test user: " + testDiscordUsername)
	}
}

func TestUpdateRating(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	createUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})

	testRating := rand.Intn(10000)
	updateRating(conn, testDiscordId, testRating)

	updatedUser := getUser(conn, testDiscordId)
	if updatedUser.CurrentRating != testRating {
		t.Error("Unable to create test user: " + testDiscordUsername)
	}
}
