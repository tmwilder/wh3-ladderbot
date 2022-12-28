package db

import (
	"gorm.io/gorm"
	"log"
	"time"
)

const MATCH_REQUEST_STATE_QUEUED = "queued"
const MATCH_REQUEST_STATE_CANCELLED = "cancelled"
const MATCH_REQUEST_STATE_COMPLETED = "completed"

const GAME_MODE_BO1 = "bo1"
const GAME_MODE_BO3 = "bo3"
const GAME_MODE_ALL = "all"

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
	// Create the new match and also its history record as one db txn
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
			panic(err)
		}
		return true, matchRequest
	}
}

/*
FindPairing Find a legal and optimal match for the current match request.
*/
func FindPairing(conn *gorm.DB, request MatchRequest) (foundPairing bool, pairing MatchRequest) {
	return false, MatchRequest{}
}

/*
CompleteRequests In an idempotent and concurrency safe way - translate two match requests to completed history
records and indicate this in their states.
*/
func CompleteRequests(conn *gorm.DB, matchRequest1 int, matchRequest2 int) (success bool) {
	return true
}

/*
CancelRequest remove the match request from queue.
*/
func CancelRequest(conn *gorm.DB, userId int) (success bool) {
	// Create the new match and also its history record as one db txn
	err := conn.Transaction(func(tx *gorm.DB) error {
		foundRequest, matchRequest := GetMatchRequest(tx, userId)
		if foundRequest == false {
			log.Printf("Unable to find match request for player: %d", userId)
			success = false
			return nil
		}

		matchRequest.MatchRequestState = MATCH_REQUEST_STATE_CANCELLED
		matchRequest.UpdatedAt = time.Now()
		createMatchRequestHistory(tx, matchRequest)
		deleteMatchRequest(tx, matchRequest.MatchRequestId)
		return nil
	})
	if err != nil {
		return false
	}
	return success
}

func GetMatchRequestHistory(conn *gorm.DB, matchRequestId int) (matchRequests []MatchRequest) {
	rows, err := conn.Raw("SELECT match_request_id, requesting_user_id, created_at, updated_at, request_range, requested_game_mode, match_request_state FROM match_requests_history WHERE match_request_id = ? ORDER BY updated_at ASC", matchRequestId).Rows()
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
