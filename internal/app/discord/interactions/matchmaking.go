package interactions

import (
	"discordbot/internal/app/discord/api"
	"discordbot/internal/db"
	"fmt"
	"gorm.io/gorm"
	"math"
	"math/rand"
	"time"
)

const RatingDeltaFloor = 600.0

/**
Find the optimal match weighting various factors.

Right now we prefer to match requests that have been in the queue for longer and that are closer to the rating of
the requester.
*/
func findBestPairing(matchRequest db.MatchRequest, requesterRating int, candidates []db.CandidatePairing) (bestMatch db.MatchRequest) {
	bestPriority := -1.0
	for _, candidate := range candidates {
		ratingDelta := requesterRating - candidate.OpponentRating
		if ratingDelta < 0 { // Golang where's my stdlib abs(i int)
			ratingDelta = -ratingDelta
		}

		// Assign a value from 1.0 for the same rating to 0.0 for >= 600 pts apart to be used to prefer closer ratings.
		ratingFraction := 1.0 - (float64(ratingDelta) / RatingDeltaFloor)
		ratingFraction = math.Max(ratingFraction, 0.0)

		// Assign a value from 1.0 for max queue wait time to 0 for just entering queue to be used to FIFO-ish pairings.
		secondsInQueue := time.Now().Sub(candidate.OpponentMatchRequest.CreatedAt).Seconds()
		queueFraction := secondsInQueue / float64(db.MaxSecondsInQueue)

		// Weight rating closeness and queue time on a 30-70 basis to produce priority score.
		priorityScore := .3*ratingFraction + .7*queueFraction

		if priorityScore >= bestPriority {
			bestPriority = priorityScore

			bestMatch = candidate.OpponentMatchRequest

			if matchRequest.RequestedGameMode == db.All && bestMatch.RequestedGameMode == db.All {
				// Set game mode to opponent's preference if we specified all.
				bestMatch.RequestedGameMode = db.Bo3
			} else if matchRequest.RequestedGameMode == db.All {
				// Nothing to do since game mode already correct.
			} else if bestMatch.RequestedGameMode == db.All {
				// Set game mode to request if opponent specified all.
				bestMatch.RequestedGameMode = matchRequest.RequestedGameMode
			}
		}
	}
	return bestMatch
}

func ExpireMatchRequests(conn *gorm.DB, discordApi api.DiscordApi) (success bool) {
	now := time.Now()
	expiredRequests := db.FindExpiredRequests(conn, now)
	messages := []string{}
	for _, v := range expiredRequests {
		success := db.CancelMatchRequest(conn, v.RequestingUserId)
		if success {
			_, user := db.GetUserById(conn, v.RequestingUserId)
			discordApi.RemoveRoleFromGuildMember(LadderQueueRoleName, user.DiscordId)
			messages = append(messages, fmt.Sprintf("Dequeued match request for user <@!%s> because it was 45m stale. Please requeue if you'd like to keep playing!\n", user.DiscordId))
		}
	}
	if len(messages) > 0 {
		// Chunk into 1-2 bulk posts to avoid hitting rate limits.
		collatedMessages := []string{}
		collatedMessage := ""
		for _, message := range messages {
			if len(collatedMessage) <= 1900 {
				collatedMessage += message
			} else {
				collatedMessages = append(collatedMessages, collatedMessage)
				collatedMessage += message
			}
		}
		collatedMessages = append(collatedMessages, collatedMessage)

		for _, v := range collatedMessages {
			api.CrossPostMessageByName(LadderFeedChannel, v)
		}
	}
	return true
}

func assignMaps(conn *gorm.DB, gameMode db.GameMode) (maps []string) {
	howMany := 1
	if gameMode == db.Bo3 {
		howMany = 3
	}
	foundMaps, mapSet := db.GetLatestMapSet(conn, gameMode)

	if !foundMaps {
		panic("Unable to find maps.")
	}

	rand.Seed(time.Now().UnixNano())

	for len(maps) < howMany {
		selection := mapSet.Maps[rand.Intn(len(mapSet.Maps))]
		noCollision := true
		for _, v := range maps {
			if selection == v {
				noCollision = false
				break
			}
		}
		if noCollision == true {
			maps = append(maps, selection)
		}
	}
	return maps
}
