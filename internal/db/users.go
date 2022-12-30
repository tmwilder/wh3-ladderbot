package db

import (
	"database/sql"
	"gorm.io/gorm"
	"log"
	"time"
)

type User struct {
	UserId          int
	DiscordId       string
	DiscordUserName string
	CurrentRating   int
}

type UserWithStats struct {
	User   User
	Wins   int
	Losses int
}

func CreateUser(conn *gorm.DB, user User) {
	conn.Exec("INSERT INTO users (discord_id, discord_username, current_rating) values (?, ?, ?)", user.DiscordId, user.DiscordUserName, user.CurrentRating)
	_, user = GetUserByDiscordId(conn, user.DiscordId)
	// Populate initial ratings history.
	UpdateUserRating(conn, user.UserId, user.CurrentRating, -1)
	if conn.Error != nil {
		panic(conn.Error)
	}
}

func GetUserByDiscordId(conn *gorm.DB, discordId string) (foundUser bool, result User) {
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

func GetUserById(conn *gorm.DB, userId int) (foundUser bool, result User) {
	row := conn.Raw("SELECT id, discord_id, discord_username, current_rating FROM users WHERE id = ?", userId).Row()
	if conn.Error != nil {
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

func GetEloLeaderboard(conn *gorm.DB) (result []UserWithStats) {
	rows, err := conn.Raw(`
		SELECT 
			u.id,
			u.discord_username,
			u.discord_id,
			u.current_rating,
			SUM(case when m1.winner ='p1' then 1 else 0 end) + SUM(case when m2.winner ='p2' then 1 else 0 end) as total_wins,
			SUM(case when m2.winner ='p1' then 1 else 0 end) + SUM(case when m1.winner ='p2' then 1 else 0 end) as total_losses
		FROM users u
		LEFT JOIN matches m1 ON
			(u.id = m1.p1_user_id AND m1.match_state = 'completed')
		LEFT JOIN matches m2 ON
			(u.id = m2.p2_user_id AND m2.match_state = 'completed')
		GROUP BY u.id
		ORDER BY current_rating DESC
	`).Rows()
	if err != nil {
		panic(err)
	}

	if conn.Error != nil {
		panic(conn.Error)
	}

	for rows.Next() {
		user := User{}
		userWithStats := UserWithStats{}
		err := rows.Scan(
			&user.UserId,
			&user.DiscordUserName,
			&user.DiscordId,
			&user.CurrentRating,
			&userWithStats.Wins,
			&userWithStats.Losses)

		if err != nil {
			log.Printf("Unable to read history row for user stats %v", err)
			return result
		}

		userWithStats.User = user
		result = append(result, userWithStats)
	}
	return result
}

func GetMonthlyWinLeaderboard(conn *gorm.DB) (result []UserWithStats) {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	rows, err := conn.Raw(`
		SELECT 
			u.id,
			u.discord_username,
			u.discord_id,
			u.current_rating,
			SUM(case when m1.winner ='p1' then 1 else 0 end) + SUM(case when m2.winner ='p2' then 1 else 0 end) as total_wins_this_month,
			SUM(case when m2.winner ='p1' then 1 else 0 end) + SUM(case when m1.winner ='p2' then 1 else 0 end) as total_losses_this_month
		FROM users u
		LEFT JOIN matches m1 ON
			(u.id = m1.p1_user_id AND m1.match_state = 'completed' AND m1.created_at >= ?)
		LEFT JOIN matches m2 ON
			(u.id = m2.p2_user_id AND m2.match_state = 'completed' AND m2.created_at >= ?)
		GROUP BY u.id
		ORDER BY total_wins_this_month DESC
	`, firstOfMonth, firstOfMonth).Rows()
	if err != nil {
		panic(err)
	}

	if conn.Error != nil {
		panic(conn.Error)
	}

	for rows.Next() {
		user := User{}
		userWithStats := UserWithStats{}
		err := rows.Scan(
			&user.UserId,
			&user.DiscordUserName,
			&user.DiscordId,
			&user.CurrentRating,
			&userWithStats.Wins,
			&userWithStats.Losses)

		if err != nil {
			log.Printf("Unable to read history row for user stats %v", err)
			return result
		}

		userWithStats.User = user
		result = append(result, userWithStats)
	}
	return result
}
