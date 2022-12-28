package db

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestCreateMatchRequest(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})
	user := GetUser(conn, testDiscordId)

	matchRequest := MatchRequest{
		0,
		user.UserId,
		time.Now(),
		time.Now(),
		200,
		GAME_MODE_ALL,
		MATCH_REQUEST_STATE_QUEUED,
	}

	success := CreateMatchRequest(conn, matchRequest)
	if !success {
		panic("Failed to create new match request.")
	}

	foundRequest, persistedRequest := GetMatchRequest(conn, user.UserId)

	if !foundRequest {
		t.Error("Unable to find persisted match request.")
	}

	if persistedRequest.RequestingUserId != user.UserId {
		t.Error("Persisted request value does not equal requesting user.")
	}
}
