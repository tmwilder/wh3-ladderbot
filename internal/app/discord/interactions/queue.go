package interactions

import (
	"discordbot/internal/db"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

func Queue(conn *gorm.DB, interaction Interaction) (success bool, channelMessage string, shouldCrossPost bool) {
	requestedGameMode := db.Bo1
	ratingRange := 300

	for _, v := range interaction.Data.Options {
		if v.Name == "range" {
			ratingRange = v.Value
		} else if v.Name == "mode" {
			requestedGameMode = db.FromInt(v.Value)
		}
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
		return false, "Found existing queued match request - if you want to change your elo range dequeue and requeue at the new range, otherwise stand by and you will be paired when a matching player joins!", false
	}

	foundActiveMatch, _ := db.GetCurrentMatch(conn, user.UserId)
	if foundActiveMatch {
		return false, "You appear to have a still open match - please report results for that before queuing again.", false
	}

	// Now with assurances of a registered user and no existing entry - try to queue their entry
	newMatchRequest := db.MatchRequest{
		RequestingUserId:  user.UserId,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		RequestRange:      ratingRange,
		RequestedGameMode: requestedGameMode,
		MatchRequestState: db.MatchRequestStateQueued,
	}
	didQueueMatch := db.CreateMatchRequest(conn, newMatchRequest)

	candidatePairings := db.FindCandidatePairings(conn, newMatchRequest)
	if len(candidatePairings) == 0 {
		return didQueueMatch, fmt.Sprintf("%s has successfully joined the matchmaking queue for mode %s with a range of %d elo points and current elo %d.", user.DiscordUserName, requestedGameMode, ratingRange, user.CurrentRating), true
	} else {
		bestPairing := findBestPairing(newMatchRequest, user.CurrentRating, candidatePairings)

		_, currentPersistedMatchRequest := db.GetMatchRequest(conn, user.UserId)

		db.CreateMatchFromRequests(conn, bestPairing, currentPersistedMatchRequest)

		_, opponent := db.GetUserById(conn, bestPairing.RequestingUserId)

		maps := assignMaps(conn, bestPairing.RequestedGameMode)

		mapStr := "[" + strings.Join(maps, ", ") + "]"

		return true, fmt.Sprintf(
				`<@!%s> (P1) joined the queue and was paired against <@!%s> (P2). Please play a %s match and report the results when done. Your randomly assigned map order is: %s.`,
				user.DiscordId,
				opponent.DiscordId,
				bestPairing.RequestedGameMode,
				mapStr),
			true
	}
}
