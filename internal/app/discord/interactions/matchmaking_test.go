package interactions

import (
	"discordbot/internal/db"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFindBestMatchOne(t *testing.T) {
	bestMatch := findBestPairing(
		db.MatchRequest{},
		800,
		[]db.CandidatePairing{
			{
				OpponentMatchRequest: db.MatchRequest{CreatedAt: time.Now()},
				OpponentRating:       800,
			},
		},
	)
	assert.NotNil(t, bestMatch)
}

func TestFindBestMatchEarlyOption(t *testing.T) {
	now := time.Now()

	twentyMinutesAgo := now.Add(-time.Duration(20) * time.Minute)

	badRatingGoodTime := db.CandidatePairing{OpponentMatchRequest: db.MatchRequest{MatchRequestId: 1, CreatedAt: twentyMinutesAgo}, OpponentRating: 1200}
	goodRatingBadTime := db.CandidatePairing{OpponentMatchRequest: db.MatchRequest{MatchRequestId: 2, CreatedAt: now}, OpponentRating: 800}
	badRatingBadTime := db.CandidatePairing{OpponentMatchRequest: db.MatchRequest{MatchRequestId: 3, CreatedAt: now}, OpponentRating: 1200}

	bestMatch := findBestPairing(
		db.MatchRequest{},
		800,
		[]db.CandidatePairing{badRatingGoodTime, goodRatingBadTime, badRatingBadTime},
	)
	assert.Equal(t, bestMatch.MatchRequestId, 1, "Expected best match to be a poor rating match that had been in queue for 20m but it was a different one.")
}

func TestAssignMaps(t *testing.T) {
	conn := db.GetGorm(db.GetTestMysSQLConnStr())
	db.InsertMapSet(conn, []string{"a", "b", "c", "d", "e", "f"}, db.Bo3)
	db.InsertMapSet(conn, []string{"a", "b", "c", "d", "e", "f"}, db.Bo1)

	maps := assignMaps(conn, db.Bo3)
	assert.Len(t, maps, 3)

	maps2 := assignMaps(conn, db.Bo1)
	assert.Len(t, maps2, 1)
}
