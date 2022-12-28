package interactions

import (
	"discordbot/internal/db"
)

func Dequeue(interaction Interaction) (success bool, channelMessage string) {
	discordUserId := interaction.Member.User.Id
	conn := db.GetDbConn()
	foundUser, user := db.GetUser(conn, discordUserId)

	if !foundUser {
		return false, "Unable to find your account in our system. You must queue at least once to register before you can dequeue. If this is a mistake contact the admins to iron it out and we'll help!"
	}

	foundMatchRequest, _ := db.GetMatchRequest(conn, user.UserId)
	if !foundMatchRequest {
		return false, "You have already dequeued - nothing to do!"
	}
	cancelledRequest := db.CancelMatchRequest(conn, user.UserId)
	if cancelledRequest {
		return true, "Dequeued successfully."
	} else {
		return false, "An unidentified technical issue happened while trying to dequeue. Please try again and if the problem persists contact admin and we will hit the TV until it works."
	}
}
