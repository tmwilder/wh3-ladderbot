package db

import (
	"database/sql"
	"gorm.io/gorm"
	"log"
	"time"
)

const MatchRequestStateQueued = "queued"
const MatchRequestStateCancelled = "cancelled"
const MatchRequestStateCompleted = "completed"

const GameModeBo1 = "bo1"
const GameModeBo3 = "bo3"
const GameModeAll = "all"

const MaxSecondsInQueue = 30 * 60

type MatchRequest struct {
	MatchRequestId    int
	RequestingUserId  int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	RequestRange      int
	RequestedGameMode string
	MatchRequestState string
}

type MatchRequestHistory struct {
}

func CreateMatchRequest(conn *gorm.DB, request MatchRequest) (success bool) {
	// Create the new match request and also its history record as one db txn
	err := conn.Transaction(func(tx *gorm.DB) error {
		tx.Exec(
			"INSERT INTO match_requests (requesting_user_id, created_at, updated_at, request_range, requested_game_mode, match_request_state) values (?, ?, ?, ?, ?, ?)",
			request.RequestingUserId,
			request.CreatedAt,
			request.UpdatedAt,
			request.RequestRange,
			request.RequestedGameMode,
			request.MatchRequestState,
		)
		if tx.Error != nil {
			log.Println(tx.Error)
			success = false
			return nil
		}
		// Get the auto incremented ID.
		foundRequest, persistedRequest := GetMatchRequest(tx, request.RequestingUserId)
		if foundRequest == false {
			log.Printf("Unable to persist new match request: %v", request)
		}
		request.MatchRequestId = persistedRequest.MatchRequestId
		createMatchRequestHistory(tx, request)
		return nil
	})
	if err != nil {
		return false
	}
	return true
}

func GetMatchRequest(conn *gorm.DB, userId int) (foundRequest bool, matchRequest MatchRequest) {
	row := conn.Raw("SELECT id, requesting_user_id, created_at, updated_at, request_range, requested_game_mode, match_request_state FROM match_requests WHERE requesting_user_id = ?", userId).Row()
	if conn.Error != nil {
		panic(conn.Error)
	}

	if row == nil {
		return false, MatchRequest{}
	} else {
		err := row.Scan(
			&matchRequest.MatchRequestId,
			&matchRequest.RequestingUserId,
			&matchRequest.CreatedAt,
			&matchRequest.UpdatedAt,
			&matchRequest.RequestRange,
			&matchRequest.RequestedGameMode,
			&matchRequest.MatchRequestState)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, MatchRequest{}
			} else {
				panic(err)
			}
		}
		return true, matchRequest
	}
}

type CandidatePairing struct {
	OpponentMatchRequest    MatchRequest
	OpponentRating          int
	OpponentDiscordUsername string
}

/*
	FindPairing Find all legal matches for the current match request as defined by being within rating range and
	matching the requested game mode.

	We'll then in-code determine an optimal one so that we can write more testable/detailed pairing routines than the DB
	query makes easy.
*/
func FindCandidatePairings(conn *gorm.DB, request MatchRequest) (response []CandidatePairing) {
	// TODO - decide if we should disqualify playing same person twice in a row.
	rows, err := conn.Raw(
		`SELECT
				mr.id,
				requesting_user_id,
				created_at,
				updated_at,
				request_range,
				requested_game_mode,
				match_request_state,
				opponent.current_rating,
				opponent.discord_username
			FROM match_requests mr
			INNER JOIN users opponent
				ON mr.requesting_user_id = opponent.id
			INNER JOIN users requester
				ON requester.id = @requester_user_id
			WHERE 
				requesting_user_id != @requester_user_id AND
				(requested_game_mode = @requested_game_mode OR requested_game_mode = @game_mode_all OR @requested_game_mode = @game_mode_all) AND
				ABS(requester.current_rating - opponent.current_rating) <= @request_rating_range AND
				ABS(requester.current_rating - opponent.current_rating) <= mr.request_range  
			ORDER BY mr.created_at ASC
			LIMIT 100`,
		sql.Named("requester_user_id", request.RequestingUserId),
		sql.Named("requested_game_mode", request.RequestedGameMode),
		sql.Named("game_mode_all", GameModeAll),
		sql.Named("request_rating_range", request.RequestRange),
	).Rows()

	if err != nil {
		panic(err)
	}

	if conn.Error != nil {
		panic(conn.Error)
	}

	for rows.Next() {
		matchRequest := MatchRequest{}
		candidatePairing := CandidatePairing{}
		err := rows.Scan(
			&matchRequest.MatchRequestId,
			&matchRequest.RequestingUserId,
			&matchRequest.CreatedAt,
			&matchRequest.UpdatedAt,
			&matchRequest.RequestRange,
			&matchRequest.RequestedGameMode,
			&matchRequest.MatchRequestState,
			&candidatePairing.OpponentRating,
			&candidatePairing.OpponentDiscordUsername)
		candidatePairing.OpponentMatchRequest = matchRequest
		if err != nil {
			log.Printf("Unable to read history row for matchRequest %d: %v", request.MatchRequestId, err)
		}
		response = append(response, candidatePairing)
	}
	return response
}

/*
CancelMatchRequest remove the match request from queue.
*/
func CancelMatchRequest(conn *gorm.DB, userId int) (success bool) {
	// Create the new match and also its history record as one db txn
	err := conn.Transaction(func(tx *gorm.DB) error {
		foundRequest, matchRequest := GetMatchRequest(tx, userId)
		if foundRequest == false {
			log.Printf("Unable to find match request for player: %d", userId)
			success = false
			return nil
		}

		matchRequest.MatchRequestState = MatchRequestStateCancelled
		matchRequest.UpdatedAt = time.Now()
		createMatchRequestHistory(tx, matchRequest)
		deleteMatchRequest(tx, matchRequest.MatchRequestId)
		success = true
		return nil
	})
	if err != nil {
		return false
	}
	return success
}

// create completed request history record and delete the match request as one txn.
func completeMatchRequest(conn *gorm.DB, request MatchRequest) (success bool) {
	err := conn.Transaction(func(tx *gorm.DB) error {
		request.MatchRequestState = MatchRequestStateCompleted
		request.UpdatedAt = time.Now()
		createMatchRequestHistory(tx, request)
		deleteMatchRequest(tx, request.MatchRequestId)
		success = true
		return nil
	})
	if err != nil {
		return false
	}
	return success
}

func GetMatchRequestHistory(conn *gorm.DB, matchRequestId int) (matchRequests []MatchRequest) {
	rows, err := conn.Raw(`
		SELECT
			match_request_id,
			requesting_user_id,
			created_at,
			updated_at,
			request_range,
			requested_game_mode,
			match_request_state
		FROM match_requests_history
		WHERE
			match_request_id = ?
		ORDER BY
			updated_at ASC`,
		matchRequestId).Rows()

	if err != nil {
		panic(err)
	}

	if conn.Error != nil {
		panic(conn.Error)
	}

	for rows.Next() {
		matchRequest := MatchRequest{}
		err := rows.Scan(
			&matchRequest.MatchRequestId,
			&matchRequest.RequestingUserId,
			&matchRequest.CreatedAt,
			&matchRequest.UpdatedAt,
			&matchRequest.RequestRange,
			&matchRequest.RequestedGameMode,
			&matchRequest.MatchRequestState)

		if err != nil {
			log.Printf("Unable to read history row for matchRequest %d: %v", matchRequestId, err)
		}
		matchRequests = append(matchRequests, matchRequest)
	}
	if err != nil {
		panic(err)
	}
	return matchRequests
}

func updateMatchRequestState(conn *gorm.DB, matchRequestId int, newState string) (success bool) {
	return true
}

func deleteMatchRequest(conn *gorm.DB, matchRequestId int) (success bool) {
	conn.Exec("DELETE FROM match_requests WHERE id = ?", matchRequestId)
	if conn.Error != nil {
		log.Println(conn.Error)
		return false
	}
	return true
}

func createMatchRequestHistory(conn *gorm.DB, request MatchRequest) (success bool) {
	conn.Exec(
		"INSERT INTO match_requests_history (match_request_id, requesting_user_id, created_at, updated_at, request_range, requested_game_mode, match_request_state) values (?, ?, ?, ?, ?, ?, ?)",
		request.MatchRequestId,
		request.RequestingUserId,
		request.CreatedAt,
		request.UpdatedAt,
		request.RequestRange,
		request.RequestedGameMode,
		request.MatchRequestState)

	if conn.Error != nil {
		log.Println(conn.Error)
		return false
	}
	return true
}
