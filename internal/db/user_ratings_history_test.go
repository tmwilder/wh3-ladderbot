package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestUpdateRating(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
	randomMatchId := 1000000

	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})

	_, user := GetUserByDiscordId(conn, testDiscordId)

	testRating := rand.Intn(10000)
	UpdateUserRating(conn, user.UserId, testRating, randomMatchId)

	_, updatedUser := GetUserByDiscordId(conn, testDiscordId)
	if updatedUser.CurrentRating != testRating {
		t.Error("Unable to create test user: " + testDiscordUsername)
	}

	updatedRating := GetUserRatingsHistory(conn, user.UserId, 1)

	assert.Len(t, updatedRating, 1)
	assert.Equal(t, testRating, updatedRating[0].Rating)

	TombstoneUserRating(conn, updatedRating[0].UserRatingId)

	tombstonedRatings := GetUserRatingsHistory(conn, user.UserId, 1)
	assert.Len(t, tombstonedRatings, 1)
}

func TestRevertRating(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
	randomMatchId := 1000000
	randomMatchId2 := 1000001
	randomMatchId3 := 1000002

	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})

	_, user := GetUserByDiscordId(conn, testDiscordId)

	testRating1 := rand.Intn(10000)
	testRating2 := rand.Intn(10000)
	testRating3 := rand.Intn(10000)

	UpdateUserRating(conn, user.UserId, testRating1, randomMatchId)
	UpdateUserRating(conn, user.UserId, testRating2, randomMatchId2)
	UpdateUserRating(conn, user.UserId, testRating3, randomMatchId3)

	updatedRating := GetUserRatingsHistory(conn, user.UserId, 20)

	assert.Len(t, updatedRating, 4)
	assert.Equal(t, testRating3, updatedRating[0].Rating)
	assert.Equal(t, testRating2, updatedRating[1].Rating)
	assert.Equal(t, testRating1, updatedRating[2].Rating)
	_, user = GetUserByDiscordId(conn, testDiscordId)
	assert.Equal(t, testRating3, user.CurrentRating)

	RevertUserRating(conn, user.UserId)

	updatedRating = GetUserRatingsHistory(conn, user.UserId, 20)
	assert.Len(t, updatedRating, 3)
	assert.Equal(t, testRating2, updatedRating[0].Rating)
	assert.Equal(t, testRating1, updatedRating[1].Rating)
	_, user = GetUserByDiscordId(conn, testDiscordId)
	assert.Equal(t, testRating2, user.CurrentRating)

	RevertUserRating(conn, user.UserId)
	updatedRating = GetUserRatingsHistory(conn, user.UserId, 20)
	assert.Len(t, updatedRating, 2)
	assert.Equal(t, testRating1, updatedRating[0].Rating)
	_, user = GetUserByDiscordId(conn, testDiscordId)
	assert.Equal(t, testRating1, user.CurrentRating)

	RevertUserRating(conn, user.UserId)
	updatedRating = GetUserRatingsHistory(conn, user.UserId, 20)
	assert.Len(t, updatedRating, 1)
}
