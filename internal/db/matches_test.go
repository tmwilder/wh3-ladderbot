package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCreateMatchFromRequests(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername1 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId1 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	testDiscordUsername2 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId2 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	CreateUser(conn, User{0, testDiscordId1, testDiscordUsername1, 750})
	CreateUser(conn, User{0, testDiscordId2, testDiscordUsername2, 800})

	_, user1 := GetUserByDiscordId(conn, testDiscordId1)
	_, user2 := GetUserByDiscordId(conn, testDiscordId2)

	matchRequest := MatchRequest{
		0,
		user1.UserId,
		time.Now(),
		time.Now(),
		100,
		Bo1,
		MatchRequestStateQueued,
	}

	matchRequest2 := MatchRequest{
		0,
		user2.UserId,
		time.Now(),
		time.Now(),
		100,
		All,
		MatchRequestStateQueued,
	}

	CreateMatchRequest(conn, matchRequest)
	CreateMatchRequest(conn, matchRequest2)

	_, r1 := GetMatchRequest(conn, user1.UserId)
	_, r2 := GetMatchRequest(conn, user2.UserId)

	success := CreateMatchFromRequests(conn, r1, r2)

	assert.True(t, success, "Failed to create match from requests.")

	deletedR1, _ := GetMatchRequest(conn, user1.UserId)
	deletedR2, _ := GetMatchRequest(conn, user2.UserId)

	assert.False(t, deletedR1, "Failed to delete first request.")
	assert.False(t, deletedR2, "Failed to delete second request.")

	_, matchForP1 := GetCurrentMatch(conn, user1.UserId)
	_, matchForP2 := GetCurrentMatch(conn, user2.UserId)

	assert.Equal(t, matchForP1.MatchId, matchForP2.MatchId)
	assert.Equal(t, matchForP1.MatchState, Matched)
}
