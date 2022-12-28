package db

type MatchState string

const (
	Matched   MatchState = "matched"
	Cancelled MatchState = "cancelled"
	Completed MatchState = "completed"
)

type GameMode string

const (
	Bo1 GameMode = "bo1"
	Bo3 GameMode = "bo3"
	All GameMode = "all"
)

type WhoWon string

const (
	P1        WhoWon = "p1"
	P2        WhoWon = "p2"
	Undefined WhoWon = ""
)
