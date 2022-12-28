package db

import (
	"database/sql"
	"gorm.io/gorm"
)

type User struct {
	UserId          int
	DiscordId       string
	DiscordUserName string
	CurrentRating   int
}

func CreateUser(conn *gorm.DB, user User) {
	conn.Exec("INSERT INTO users (discord_id, discord_username, current_rating) values (?, ?, ?)", user.DiscordId, user.DiscordUserName, user.CurrentRating)
	if conn.Error != nil {
		panic(conn.Error)
	}
}

func GetUser(conn *gorm.DB, discordId string) (foundUser bool, result User) {
	row := conn.Raw("SELECT id, discord_id, discord_username, current_rating FROM users WHERE discord_id = ?", discordId).Row()
	if conn.Error != nil {
		// TODO - How does this work with pooling and concurrency?
		panic(conn.Error)
	}
	err := row.Scan(&result.UserId, &result.DiscordId, &result.DiscordUserName, &result.CurrentRating)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, result
		} else {
			panic(err)
		}
	}
	return true, result
}

func UpdateUserRating(conn *gorm.DB, discordId string, newRating int) {
	conn.Exec("UPDATE users SET current_rating = ? WHERE discord_id = ?", newRating, discordId)
	if conn.Error != nil {
		panic(conn.Error)
	}
}
