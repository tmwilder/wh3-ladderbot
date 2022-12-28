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
		GAME_MODE_ALL,
		MATCH_REQUEST_STATE_QUEUED,
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

	if history[1].MatchRequestState != MATCH_REQUEST_STATE_CANCELLED {
		t.Errorf("Expected cancelled history state to be %s but was %s", MATCH_REQUEST_STATE_CANCELLED, history[1].MatchRequestState)
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

	CreateUser(conn, User{0, testDiscordId1, testDiscordUsername1, 800})
	CreateUser(conn, User{0, testDiscordId2, testDiscordUsername2, 900})
	CreateUser(conn, User{0, testDiscordId3, testDiscordUsername3, 1000})

	_, user1 := GetUser(conn, testDiscordId1)
	_, user2 := GetUser(conn, testDiscordId2)
	_, user3 := GetUser(conn, testDiscordId3)

	matchRequest := MatchRequest{
		0,
		user1.UserId,
		time.Now(),
		time.Now(),
		99,
		GAME_MODE_BO1,
		MATCH_REQUEST_STATE_QUEUED,
	}

	matchRequest2 := MatchRequest{
		0,
		user2.UserId,
		time.Now(),
		time.Now(),
		100,
		GAME_MODE_ALL,
		MATCH_REQUEST_STATE_QUEUED,
	}

	matchRequest3 := MatchRequest{
		0,
		user3.UserId,
		time.Now(),
		time.Now(),
		200,
		GAME_MODE_BO3,
		MATCH_REQUEST_STATE_QUEUED,
	}

	CreateMatchRequest(conn, matchRequest)
	CreateMatchRequest(conn, matchRequest2)
	CreateMatchRequest(conn, matchRequest3)

	pairings1 := FindCandidatePairings(conn, matchRequest)
	pairings2 := FindCandidatePairings(conn, matchRequest2)
	pairings3 := FindCandidatePairings(conn, matchRequest3)

	// Request 1 has too low a range and should match with nothing
	assert.Equal(t, len(pairings1), 0)

	// Request 2 allows all modes and enough range and should match with both other requests
	assert.Equal(t, len(pairings2), 2)
	assert.Equal(t, pairings2[0].RequestingUserId, user1.UserId)
	assert.Equal(t, pairings2[1].RequestingUserId, user3.UserId)

	// Request 3 has enough range to match with both others but is only for BO3 and so should match with only the 2nd
	assert.Equal(t, len(pairings3), 1)
	assert.Equal(t, pairings3[0].RequestingUserId, user2.UserId)
}
