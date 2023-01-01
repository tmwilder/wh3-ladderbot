package db

import (
	"fmt"
	"log"
)

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

func ToInt(gameMode GameMode) int {
	switch gameMode {
	case Bo1:
		return 1
	case Bo3:
		return 2
	case All:
		return 3
	default:
		log.Panicf("Unrecognized game mode: %v", gameMode)
		return -1
	}
}

func FromInt(intMode int) GameMode {
	switch intMode {
	case 1:
		return Bo1
	case 2:
		return Bo3
	case 3:
		return All
	default:
		panic(fmt.Sprintf("Unrecognized game mode int: %d", intMode))
	}
}

type WhoWon string

const (
	P1        WhoWon = "p1"
	P2        WhoWon = "p2"
	Undefined WhoWon = ""
)
