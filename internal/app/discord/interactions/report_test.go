package interactions

import (
	"discordbot/internal/app/discord/commands"
	"discordbot/internal/db"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"math/rand"
	"testing"
	"time"
)

func setUpTestMatch(conn *gorm.DB) (user1 db.User, user2 db.User, match db.Match) {
	testDiscordUsername1 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId1 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	testDiscordUsername2 := fmt.Sprintf("coolsk8r1990%d", rand.Intn(1000000))
	testDiscordId2 := fmt.Sprintf("somediscordId%d", rand.Intn(1000000))

	db.CreateUser(conn, db.User{DiscordId: testDiscordId1, DiscordUserName: testDiscordUsername1, CurrentRating: db.DEFAULT_RATING})
	db.CreateUser(conn, db.User{DiscordId: testDiscordId2, DiscordUserName: testDiscordUsername2, CurrentRating: db.DEFAULT_RATING})

	_, user1 = db.GetUserByDiscordId(conn, testDiscordId1)
	_, user2 = db.GetUserByDiscordId(conn, testDiscordId2)

	matchRequest := db.MatchRequest{
		RequestingUserId:  user1.UserId,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RequestRange:      100,
		RequestedGameMode: db.Bo1,
		MatchRequestState: db.MatchRequestStateQueued,
	}

	matchRequest2 := db.MatchRequest{
		RequestingUserId:  user2.UserId,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RequestRange:      100,
		RequestedGameMode: db.Bo1,
		MatchRequestState: db.MatchRequestStateQueued,
	}

	db.CreateMatchRequest(conn, matchRequest)
	db.CreateMatchRequest(conn, matchRequest2)

	_, r1 := db.GetMatchRequest(conn, user1.UserId)
	_, r2 := db.GetMatchRequest(conn, user2.UserId)

	db.CreateMatchFromRequests(conn, r1, r2)

	_, match = db.GetMostRecentMatch(conn, user1.UserId)

	return user1, user2, match
}

func TestReportWin(t *testing.T) {
	conn := db.GetGorm(db.GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())
	user1, user2, match := setUpTestMatch(conn)

	Report(conn, Interaction{
		Member: DiscordMemberInfo{DiscordUser{Id: user1.DiscordId, Username: user1.DiscordUserName}},
		Data: InteractionData{
			Name: commands.Report,
			Options: []OptionData{
				{
					Type:  3,
					Name:  "Win",
					Value: int(commands.Win),
				},
			},
		}})

	_, updatedMatch := db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User := db.GetUserById(conn, user1.UserId)
	_, updatedP2User := db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Completed, updatedMatch.MatchState)

	assert.Equal(t, 1216, updatedP1User.CurrentRating)
	assert.Equal(t, 1184, updatedP2User.CurrentRating)
}

func TestReportLoss(t *testing.T) {
	conn := db.GetGorm(db.GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())
	user1, user2, match := setUpTestMatch(conn)

	Report(conn, Interaction{
		Member: DiscordMemberInfo{DiscordUser{Id: user1.DiscordId, Username: user1.DiscordUserName}},
		Data: InteractionData{
			Name: commands.Report,
			Options: []OptionData{
				{
					Type:  3,
					Name:  "Win",
					Value: int(commands.Loss),
				},
			},
		}})

	_, updatedMatch := db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User := db.GetUserById(conn, user1.UserId)
	_, updatedP2User := db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Completed, updatedMatch.MatchState)

	assert.Equal(t, 1184, updatedP1User.CurrentRating)
	assert.Equal(t, 1216, updatedP2User.CurrentRating)
}

func TestReportCancel(t *testing.T) {
	conn := db.GetGorm(db.GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())
	user1, user2, match := setUpTestMatch(conn)

	Report(conn, Interaction{
		Member: DiscordMemberInfo{DiscordUser{Id: user1.DiscordId, Username: user1.DiscordUserName}},
		Data: InteractionData{
			Name: commands.Report,
			Options: []OptionData{
				{
					Type:  3,
					Name:  "Win",
					Value: int(commands.Cancel),
				},
			},
		}})

	_, updatedMatch := db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User := db.GetUserById(conn, user1.UserId)
	_, updatedP2User := db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Cancelled, updatedMatch.MatchState)

	assert.Equal(t, 1200, updatedP1User.CurrentRating)
	assert.Equal(t, 1200, updatedP2User.CurrentRating)
}

func TestReportCircularClusterOfNonsense(t *testing.T) {
	conn := db.GetGorm(db.GetTestMysSQLConnStr())
	rand.Seed(time.Now().UnixNano())
	user1, user2, match := setUpTestMatch(conn)

	interaction := Interaction{
		Member: DiscordMemberInfo{DiscordUser{Id: user1.DiscordId, Username: user1.DiscordUserName}},
		Data: InteractionData{
			Name: commands.Report,
			Options: []OptionData{
				{
					Type:  3,
					Name:  "Win",
					Value: int(commands.Win),
				},
			},
		}}

	// Report a win
	Report(conn, interaction)

	_, updatedMatch := db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User := db.GetUserById(conn, user1.UserId)
	_, updatedP2User := db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Completed, updatedMatch.MatchState)

	assert.Equal(t, 1216, updatedP1User.CurrentRating)
	assert.Equal(t, 1184, updatedP2User.CurrentRating)

	// Cancel the win
	interaction.Data.Options[0].Value = int(commands.Cancel)
	Report(conn, interaction)

	// #TODO - why is p2 not getting reset to 1200 - look at history and think about our stack data structure approach.
	_, updatedMatch = db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User = db.GetUserById(conn, user1.UserId)
	_, updatedP2User = db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Cancelled, updatedMatch.MatchState)

	assert.Equal(t, 1200, updatedP1User.CurrentRating)
	assert.Equal(t, 1200, updatedP2User.CurrentRating)

	// Report a loss instead
	interaction.Data.Options[0].Value = int(commands.Loss)
	Report(conn, interaction)

	_, updatedMatch = db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User = db.GetUserById(conn, user1.UserId)
	_, updatedP2User = db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Completed, updatedMatch.MatchState)

	assert.Equal(t, 1184, updatedP1User.CurrentRating)
	assert.Equal(t, 1216, updatedP2User.CurrentRating)

	// Report the same loss again
	interaction.Data.Options[0].Value = int(commands.Loss)
	Report(conn, interaction)

	_, updatedMatch = db.GetMostRecentMatch(conn, user2.UserId)
	_, updatedP1User = db.GetUserById(conn, user1.UserId)
	_, updatedP2User = db.GetUserById(conn, user2.UserId)

	assert.Equal(t, match.MatchId, updatedMatch.MatchId)
	assert.Equal(t, db.Completed, updatedMatch.MatchState)

	assert.Equal(t, 1184, updatedP1User.CurrentRating)
	assert.Equal(t, 1216, updatedP2User.CurrentRating)
}
