package db

import (
	"fmt"
	"github.com/go-playground/assert/v2"
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
	_, user := GetUser(conn, testDiscordId)

	matchRequest := MatchRequest{
		0,
		user.UserId,
		time.Now(),
		time.Now(),
		200,
		GameModeAll,
		MatchRequestStateQueued,
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

	history := GetMatchRequestHistory(conn, persistedRequest.MatchRequestId)

	if len(history) != 1 {
		t.Errorf("Got history of len %d instead of 1.", len(history))
	}

	if history[0].MatchRequestId != persistedRequest.MatchRequestId {
		t.Errorf("Expected match_request_id in history of %d but got %d", persistedRequest.MatchRequestId, history[0].MatchRequestId)
	}
}

func TestCancelMatchRequest(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})
	_, user := GetUser(conn, testDiscordId)

	matchRequest := MatchRequest{
		0,
		user.UserId,
		time.Now(),
		time.Now(),
		200,
		GameModeAll,
		MatchRequestStateQueued,
	}

	CreateMatchRequest(conn, matchRequest)

	_, persistedRequest := GetMatchRequest(conn, user.UserId)

	didCancelRequest := CancelMatchRequest(conn, user.UserId)

	if !didCancelRequest {
		t.Errorf("Unable to cancel match request.")
	}

	foundCancelledRequest, _ := GetMatchRequest(conn, user.UserId)

	if foundCancelledRequest {
		t.Error("Request still exists despite being cancelled.")
	}

	history := GetMatchRequestHistory(conn, persistedRequest.MatchRequestId)

	if len(history) != 2 {
		t.Errorf("Got history of len %d instead of 2.", len(history))
	}

	if history[1].MatchRequestId != persistedRequest.MatchRequestId {
		t.Errorf("Expected match_request_id in history of %d but got %d", persistedRequest.MatchRequestId, history[0].MatchRequestId)
	}

	if history[1].MatchRequestState != MatchRequestStateCancelled {
		t.Errorf("Expected cancelled history state to be %s but was %s", MatchRequestStateCancelled, history[1].MatchRequestState)
	}
}

func TestFindCandidatePairings(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	// Wipe match requests in test db so we can do global queries.
	// Concurrency unsafe test - I claim hobby project status : D.
	conn.Exec("TRUNCATE TABLE match_requests")

	if conn.Error != nil {
		t.Errorf("%v", conn.Error)
	}

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername1 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId1 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	testDiscordUsername2 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId2 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	testDiscordUsername3 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId3 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	testDiscordUsername4 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId4 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	CreateUser(conn, User{0, testDiscordId1, testDiscordUsername1, 750})
	CreateUser(conn, User{0, testDiscordId2, testDiscordUsername2, 800})
	CreateUser(conn, User{0, testDiscordId3, testDiscordUsername3, 900})
	CreateUser(conn, User{0, testDiscordId4, testDiscordUsername4, 1000})

	_, user1 := GetUser(conn, testDiscordId1)
	_, user2 := GetUser(conn, testDiscordId2)
	_, user3 := GetUser(conn, testDiscordId3)
	_, user4 := GetUser(conn, testDiscordId4)

	matchRequest := MatchRequest{
		0,
		user1.UserId,
		time.Now(),
		time.Now(),
		49,
		GameModeBo1,
		MatchRequestStateQueued,
	}

	matchRequest2 := MatchRequest{
		0,
		user2.UserId,
		time.Now(),
		time.Now(),
		150,
		GameModeBo1,
		MatchRequestStateQueued,
	}

	matchRequest3 := MatchRequest{
		0,
		user3.UserId,
		time.Now(),
		time.Now(),
		300,
		GameModeAll,
		MatchRequestStateQueued,
	}

	matchRequest4 := MatchRequest{
		0,
		user4.UserId,
		time.Now(),
		time.Now(),
		500,
		GameModeBo3,
		MatchRequestStateQueued,
	}

	CreateMatchRequest(conn, matchRequest)
	CreateMatchRequest(conn, matchRequest2)
	CreateMatchRequest(conn, matchRequest3)
	CreateMatchRequest(conn, matchRequest4)

	pairings1 := FindCandidatePairings(conn, matchRequest)
	pairings2 := FindCandidatePairings(conn, matchRequest2)
	pairings3 := FindCandidatePairings(conn, matchRequest3)
	pairings4 := FindCandidatePairings(conn, matchRequest4)

	// Request 1 has too low a range and should match with nothing despite others being open to matching with it.
	assert.Equal(t, len(pairings1), 0)

	// Request 2 allows bo1 and enough range to match with 3, but should not match with 1 because of 1s range or 4 because 4 is outside of 2s range.
	assert.Equal(t, len(pairings2), 1)
	assert.Equal(t, pairings2[0].OpponentMatchRequest.RequestingUserId, user3.UserId)
	assert.Equal(t, pairings2[0].OpponentRating, user3.CurrentRating)
	assert.Equal(t, pairings2[0].OpponentDiscordUsername, user3.DiscordUserName)

	// Request 3 has enough range to match with all others but so should match with 2 and 4 who don't preclude it, but not 1 who does.
	assert.Equal(t, len(pairings3), 2)
	assert.Equal(t, pairings3[0].OpponentMatchRequest.RequestingUserId, user2.UserId)
	assert.Equal(t, pairings3[0].OpponentRating, user2.CurrentRating)
	assert.Equal(t, pairings3[0].OpponentDiscordUsername, user2.DiscordUserName)
	assert.Equal(t, pairings3[1].OpponentMatchRequest.RequestingUserId, user4.UserId)
	assert.Equal(t, pairings3[1].OpponentRating, user4.CurrentRating)
	assert.Equal(t, pairings3[1].OpponentDiscordUsername, user4.DiscordUserName)

	// Request 4 has enough range to match with all others but is only for bo3 so should only match with 3.
	assert.Equal(t, len(pairings4), 1)
	assert.Equal(t, pairings4[0].OpponentMatchRequest.RequestingUserId, user3.UserId)
	assert.Equal(t, pairings4[0].OpponentRating, user3.CurrentRating)
	assert.Equal(t, pairings4[0].OpponentDiscordUsername, user3.DiscordUserName)
}
