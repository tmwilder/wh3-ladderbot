package db

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

type Match struct {
	MatchId          int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	MatchState       MatchState
	GameMode         GameMode
	P1UserId         int
	P2UserId         int
	P1MatchRequestId int
	P2MatchRequestId int
	Winner           WhoWon
}

/*
CreateMatchFromRequests

Translate two match requests to completed history records and indicate this in their states and create a Match from
them.
*/
func CreateMatchFromRequests(conn *gorm.DB, matchRequest1 MatchRequest, matchRequest2 MatchRequest) (success bool) {
	err := conn.Transaction(func(tx *gorm.DB) error {
		// Create a match w/MR1 and 2 - need
		persisted := CreateMatch(
			tx,
			Match{
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
				MatchState:       Matched,
				GameMode:         GameMode(matchRequest1.RequestedGameMode),
				P1UserId:         matchRequest1.RequestingUserId,
				P2UserId:         matchRequest2.RequestingUserId,
				P1MatchRequestId: matchRequest1.MatchRequestId,
				P2MatchRequestId: matchRequest2.MatchRequestId,
				Winner:           Undefined,
			})
		if !persisted {
			fmt.Printf("Failed to persist match, aborting create match from requests for match requests %v %v", matchRequest1, matchRequest2)
			success = false
			return nil
		}

		delete1 := completeMatchRequest(tx, matchRequest1)
		if !delete1 {
			success = false
			return nil
		}
		delete2 := completeMatchRequest(tx, matchRequest2)
		if !delete2 {
			success = false
			return nil
		}
		success = true
		return nil
	})
	if err != nil {
		log.Println(err)
		return false
	}
	return success
}

func CreateMatch(conn *gorm.DB, match Match) (success bool) {
	// Create the new match and also its history record as one db txn
	err := conn.Transaction(func(tx *gorm.DB) error {

		// Enforce invariant of only one queued match per user at a time.
		collision1, persistedCollision1 := GetCurrentMatch(tx, match.P1UserId)
		collision2, persistedCollision2 := GetCurrentMatch(tx, match.P2UserId)

		if collision1 {
			log.Printf("Not persisting match %v because user %d already had match %d open.", match, match.P1UserId, persistedCollision1.MatchId)
			success = false
			return nil
		}
		if collision2 {
			log.Printf("Not persisting match %v because user %d already had match %d open.", match, match.P2UserId, persistedCollision2.MatchId)
			success = false
			return nil
		}

		tx.Exec(
			"INSERT INTO matches (created_at, updated_at, match_state, game_mode, p1_user_id, p2_user_id, p1_match_request_id, p2_match_request_id, winner) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			match.CreatedAt,
			match.UpdatedAt,
			match.MatchState,
			match.GameMode,
			match.P1UserId,
			match.P2UserId,
			match.P1MatchRequestId,
			match.P2MatchRequestId,
			match.Winner,
		)
		if tx.Error != nil {
			log.Println(tx.Error)
			success = false
			return nil
		}
		// Get the auto incremented ID.
		foundRequest, persistedMatch := GetCurrentMatch(tx, match.P1UserId)
		if foundRequest == false {
			log.Printf("Unable to find freshly persisted match so we cannot create history record and will bail: %v", match)
			success = false
			return nil
		}
		match.MatchId = persistedMatch.MatchId
		CreateMatchHistory(tx, match)
		success = true
		return nil
	})
	if err != nil {
		return false
	}
	return success
}

/*
	GetCurrentMatch gets the current match if any for the specified user.
*/
func GetCurrentMatch(conn *gorm.DB, userId int) (foundMatch bool, result Match) {
	row := conn.Raw(`
			SELECT 
				id,
				created_at,
				updated_at,
				match_state,
				game_mode,
				p1_user_id,
				p2_user_id,
				p1_match_request_id,
				p2_match_request_id,
				winner
			FROM matches
			WHERE
				(p1_user_id = ? OR p2_user_id = ?) AND
				match_state = ?`,
		userId,
		userId,
		Matched,
	).Row()
	if conn.Error != nil {
		log.Println(conn.Error)
		return false, Match{}
	}
	return parseMatchRow(row)
}

/*
	GetCurrentMatch gets the most recent match if any for the specified user, regardless of whether it is still matched
	so this means fetching cancelled/completed matches as well. We use this for performing user-fixes to data issues.
*/
func GetMostRecentMatch(conn *gorm.DB, userId int) (foundMatch bool, result Match) {
	row := conn.Raw(`
			SELECT 
				id,
				created_at,
				updated_at,
				match_state,
				game_mode,
				p1_user_id,
				p2_user_id,
				p1_match_request_id,
				p2_match_request_id,
				winner
			FROM matches
			WHERE
				(p1_user_id = ? OR p2_user_id = ?) 
			ORDER BY created_at DESC
			LIMIT 1`,
		userId,
		userId,
	).Row()
	if conn.Error != nil {
		log.Println(conn.Error)
		return false, Match{}
	}
	return parseMatchRow(row)
}

func GetMatchById(conn *gorm.DB, matchId int) (foundMatch bool, match Match) {
	row := conn.Raw(`
			SELECT 
				id,
				created_at,
				updated_at,
				match_state,
				game_mode,
				p1_user_id,
				p2_user_id,
				p1_match_request_id,
				p2_match_request_id,
				winner
			FROM matches
			WHERE
				id = ?`, matchId).Row()
	if conn.Error != nil {
		log.Println(conn.Error)
		return false, Match{}
	}
	return parseMatchRow(row)
}

func parseMatchRow(row *sql.Row) (success bool, result Match) {
	err := row.Scan(
		&result.MatchId,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.MatchState,
		&result.GameMode,
		&result.P1UserId,
		&result.P2UserId,
		&result.P1MatchRequestId,
		&result.P2MatchRequestId,
		&result.Winner)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, Match{}
		} else {
			log.Println(err)
			return false, Match{}
		}
	}
	return true, result
}

func UpdateMatch(conn *gorm.DB, matchId int, state MatchState, winner WhoWon) (success bool) {
	now := time.Now()
	conn.Exec("UPDATE matches SET match_state = ?, winner = ?, updated_at = ? WHERE id = ?", state, winner, now, matchId)
	if conn.Error != nil {
		log.Println(conn.Error)
		return false
	}
	_, match := GetMatchById(conn, matchId)
	CreateMatchHistory(conn, match)
	return true
}

func CreateMatchHistory(conn *gorm.DB, match Match) (success bool) {
	conn.Exec(
		"INSERT INTO matches_history (match_id, created_at, updated_at, match_state, game_mode, p1_user_id, p2_user_id, p1_match_request_id, p2_match_request_id, winner) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		match.MatchId,
		match.CreatedAt,
		match.UpdatedAt,
		match.MatchState,
		match.GameMode,
		match.P1UserId,
		match.P2UserId,
		match.P1MatchRequestId,
		match.P2MatchRequestId,
		match.Winner,
	)
	if conn.Error != nil {
		log.Println(conn.Error)
		return false
	}
	return true
}

func GetMatchHistory(conn *gorm.DB, matchId int) []Match {
	return []Match{}
}
