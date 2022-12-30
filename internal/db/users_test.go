package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCreateUser(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())

	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, DEFAULT_RATING})

	_, user := GetUserByDiscordId(conn, testDiscordId)

	if user.DiscordUserName != testDiscordUsername {
		t.Error("Unable to create test user: " + testDiscordUsername)
	}
}

func TestGetUserStatsByRating(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 10; i++ {
		testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
		testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
		CreateUser(conn, User{0, testDiscordId, testDiscordUsername, rand.Intn(1500)})
	}
	// Insert one very high elo user.
	veryHighElo := 10000
	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, veryHighElo})

	usersWithStats := GetEloLeaderboard(conn)
	assert.Greater(t, len(usersWithStats), 10)
	assert.Equal(t, usersWithStats[0].User.CurrentRating, veryHighElo)
}

func TestGetMonthlyWinLeaderboard(t *testing.T) {
	conn := GetGorm(GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())

	testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
	CreateUser(conn, User{0, testDiscordId, testDiscordUsername, rand.Intn(1500)})

	_, thePatsy := GetUserByDiscordId(conn, testDiscordId)

	now := time.Now()
	for i := 0; i < 4; i++ {
		testDiscordUsername := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
		testDiscordId := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))
		CreateUser(conn, User{0, testDiscordId, testDiscordUsername, rand.Intn(1500)})

		_, theLatestWinner := GetUserByDiscordId(conn, testDiscordId)

		for j := 0; j < i+7; j++ {
			CreateMatch(conn, Match{
				CreatedAt:        now,
				UpdatedAt:        now,
				MatchState:       Completed,
				GameMode:         Bo3,
				P1UserId:         thePatsy.UserId,
				P2UserId:         theLatestWinner.UserId,
				P1MatchRequestId: -1,
				P2MatchRequestId: -1,
				Winner:           P2,
			})
		}
	}

	usersWithStats := GetMonthlyWinLeaderboard(conn)
	assert.GreaterOrEqual(t, len(usersWithStats), 10)
	assert.GreaterOrEqual(t, usersWithStats[0].Wins, usersWithStats[1].Wins)
	assert.GreaterOrEqual(t, usersWithStats[1].Wins, usersWithStats[2].Wins)
}
