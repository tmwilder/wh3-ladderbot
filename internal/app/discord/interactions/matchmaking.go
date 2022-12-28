package interactions

import (
	"discordbot/internal/db"
	"math"
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

			if matchRequest.RequestedGameMode == db.GameModeAll && bestMatch.RequestedGameMode == db.GameModeAll {
				// Set game mode to opponent's preference if we specified all.
				bestMatch.RequestedGameMode = db.GameModeBo3
			} else if matchRequest.RequestedGameMode == db.GameModeAll {
				// Nothing to do since game mode already correct.
			} else if bestMatch.RequestedGameMode == db.GameModeAll {
				// Set game mode to request if opponent specified all.
				bestMatch.RequestedGameMode = matchRequest.RequestedGameMode
			}
		}
	}
	return bestMatch
}
