package db

import (
	"gorm.io/gorm"
	"log"
	"time"
)

const DEFAULT_RATING = 1200

type UserRating struct {
	UserRatingId int
	Rating       int
	UserId       int
	MatchId      int
	IsTombstoned bool
	CreatedAt    time.Time
}

func UpdateUserRating(conn *gorm.DB, userId int, newRating int, matchId int) (success bool) {
	err := conn.Transaction(func(tx *gorm.DB) error {
		tx.Exec("UPDATE users SET current_rating = ? WHERE id = ?", newRating, userId)
		if conn.Error != nil {
			panic(conn.Error)
		}

		createdAt := time.Now()
		tx.Exec(
			"INSERT INTO user_ratings_history (user_id, rating, match_id, is_tombstoned, created_at) values (?, ?, ?, ?, ?)",
			userId,
			newRating,
			matchId,
			false,
			createdAt,
		)
		if tx.Error != nil {
			log.Println(tx.Error)
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

/*
	Tombstones the most recent rating and unwinds the user rating to one rating ago.
*/
func RevertUserRating(conn *gorm.DB, userId int) (success bool) {
	err := conn.Transaction(func(tx *gorm.DB) error {
		ratingHistory := GetUserRatingsHistory(tx, userId, 2)
		if len(ratingHistory) != 2 {
			log.Printf("Cannot tombstone rating for user: %d because they have insufficient ratings history.", userId)
			success = false
			return nil
		}
		currentRating := ratingHistory[0]
		priorRating := ratingHistory[1]
		TombstoneUserRating(tx, currentRating.UserRatingId)
		tx.Exec("UPDATE users SET current_rating = ? WHERE id = ?", priorRating.Rating, userId)
		if conn.Error != nil {
			panic(conn.Error)
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

func GetUserRatingsHistory(conn *gorm.DB, userId int, limit int) (ratings []UserRating) {
	rows, err := conn.Raw(`
		SELECT
			id,
			rating,
			user_id,
			match_id,
			is_tombstoned,
			created_at
		FROM user_ratings_history
		WHERE
			user_id = ? AND
			is_tombstoned = false
		ORDER BY
			id DESC
		LIMIT ?`,
		userId,
		limit).Rows()

	if err != nil {
		panic(err)
	}

	if conn.Error != nil {
		panic(conn.Error)
	}

	for rows.Next() {
		userRating := UserRating{}
		err := rows.Scan(
			&userRating.UserRatingId,
			&userRating.Rating,
			&userRating.UserId,
			&userRating.MatchId,
			&userRating.IsTombstoned,
			&userRating.CreatedAt)

		if err != nil {
			log.Printf("Unable to read rating history row for user id %d %v", userId, err)
		}
		ratings = append(ratings, userRating)
	}
	if err != nil {
		log.Println(err)
		return []UserRating{}
	}
	return ratings
}

func TombstoneUserRating(conn *gorm.DB, ratingId int) {
	conn.Exec("UPDATE user_ratings_history SET is_tombstoned = true WHERE id = ?", ratingId)
	if conn.Error != nil {
		panic(conn.Error)
	}
}
