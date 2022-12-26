package db

import (
	"gorm.io/gorm"
)

type User struct {
	DiscordId string
	Email     string
}

func createUser(conn *gorm.DB, user User) {
	conn.Exec("INSERT INTO users (discord_id, email) values (?, ?)", user.DiscordId, user.Email)
	if conn.Error != nil {
		// TODO - How does this work with pooling and concurrency
		panic(conn.Error)
	}
}

func getUser(conn *gorm.DB, email string) (result User) {
	row := conn.Raw("SELECT email, discord_id FROM users WHERE email = ?", email).Row()
	if conn.Error != nil {
		// TODO - How does this work with pooling and concurrency
		panic(conn.Error)
	}
	err := row.Scan(&result.Email, &result.DiscordId)
	if err != nil {
		panic(err)
	}
	return result
}
