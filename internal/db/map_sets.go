package db

import (
	"database/sql"
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"time"
)

type MapSet struct {
	MapSetId  int
	Maps      []string
	CreatedAt time.Time
	GameMode  GameMode
}

func InsertMapSet(conn *gorm.DB, mapSet []string, gameMode GameMode) (success bool) {
	createdAt := time.Now()

	jsonMaps, err := json.Marshal(mapSet)
	if err != nil {
		log.Printf("Unable to serialize maps for persistence: %v", err)
		return false
	}

	conn.Exec(
		"INSERT INTO map_sets (created_at, map_set, game_mode) values (?, ?, ?)",
		createdAt,
		jsonMaps,
		gameMode,
	)
	if conn.Error != nil {
		log.Println(conn.Error)
		success = false
		return false
	}
	return true
}

func GetLatestMapSet(conn *gorm.DB, gameMode GameMode) (foundSet bool, result MapSet) {
	row := conn.Raw("SELECT id, created_at, map_set, game_mode FROM map_sets WHERE game_mode = ? ORDER BY id DESC LIMIT 1", gameMode).Row()
	if conn.Error != nil {
		panic(conn.Error)
	}

	if row == nil {
		return false, MapSet{}
	} else {
		var serializedMaps []byte
		err := row.Scan(
			&result.MapSetId,
			&result.CreatedAt,
			&serializedMaps,
			&result.GameMode)

		var deserializedMaps []string

		err = json.Unmarshal(serializedMaps, &deserializedMaps)

		result.Maps = deserializedMaps

		if err != nil {
			log.Printf("Unable to deserialized maps: %s", err)
			return false, MapSet{}
		}

		if err != nil {
			if err == sql.ErrNoRows {
				return false, MapSet{}
			} else {
				panic(err)
			}
		}
		return true, result
	}
}
