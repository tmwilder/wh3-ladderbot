package interactions

import (
	"discordbot/internal/db"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

func Queue(conn *gorm.DB, interaction Interaction) (success bool, channelMessage string) {
	queueValue := 300
	if len(interaction.Data.Options) > 0 {
		queueValue = interaction.Data.Options[0].Value
	}

	discordUserId := interaction.Member.User.Id
	discordUserName := interaction.Member.User.Username

	// Check to see if the user exists, if not create them.
	// We do this to avoid users ever having a register step - this takes advantage of Discord's Authn
	// and bot token validation flows.
	foundUser, user := db.GetUserByDiscordId(conn, discordUserId)
	if !foundUser {
		db.CreateUser(
			conn,
			db.User{
				DiscordId:       discordUserId,
				DiscordUserName: discordUserName,
				CurrentRating:   db.DEFAULT_RATING,
			})
		_, user = db.GetUserByDiscordId(conn, discordUserId)
	}

	foundEntry, _ := db.GetMatchRequest(conn, user.UserId)
	if foundEntry {
		return false, "Found existing queued match request - if you want to change your elo range dequeue and requeue at the new range, otherwise stand by and you will be paired when a matching player joins!"
	}

	foundActiveMatch, _ := db.GetCurrentMatch(conn, user.UserId)
	if foundActiveMatch {
		return false, "You appear to have a still open match - please report results for that before queuing again."
	}

	// Now with assurances of a registered user and no existing entry - try to queue their entry
	newMatchRequest := db.MatchRequest{
		RequestingUserId:  user.UserId,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RequestRange:      queueValue, // TODO fix client param tooltip
		RequestedGameMode: db.All,     // TODO add as client param
		MatchRequestState: db.MatchRequestStateQueued,
	}
	didQueueMatch := db.CreateMatchRequest(conn, newMatchRequest)

	candidatePairings := db.FindCandidatePairings(conn, newMatchRequest)
	if len(candidatePairings) == 0 {
		return didQueueMatch, fmt.Sprintf("You have successfully joined the matchmaking queue with a range of %d elo points.", queueValue)
	} else {
		bestPairing := findBestPairing(newMatchRequest, user.CurrentRating, candidatePairings)

		_, currentPersistedMatchRequest := db.GetMatchRequest(conn, user.UserId)

		db.CreateMatchFromRequests(conn, bestPairing, currentPersistedMatchRequest)

		_, opponent := db.GetUserById(conn, bestPairing.RequestingUserId)

		maps := assignMaps(conn, bestPairing.RequestedGameMode)

		mapStr := "[" + strings.Join(maps, ", ") + "]"

		// TODO better msg here
		return true, fmt.Sprintf(
			`%s joined the queue and was paired against %s. Please play a %s match and report the results when done. Your randomly assigned map order will be: %s.`,
			user.DiscordUserName,
			opponent.DiscordUserName,
			bestPairing.RequestedGameMode,
			mapStr)
	}
}
