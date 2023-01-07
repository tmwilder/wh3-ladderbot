package interactions

import (
	"discordbot/internal/app/discord/api"
	"discordbot/internal/db"
	"fmt"
	"gorm.io/gorm"
)

func Dequeue(conn *gorm.DB, interaction api.Interaction) (success bool, channelMessage string, shouldCrossPost bool) {
	discordUserId := interaction.Member.User.Id

	foundUser, user := db.GetUserByDiscordId(conn, discordUserId)

	if !foundUser {
		return false, "Unable to find your account in our system. You must queue at least once to register before you can dequeue. If this is a mistake contact the admins to iron it out and we'll help!", false
	}

	foundMatchRequest, _ := db.GetMatchRequest(conn, user.UserId)
	if !foundMatchRequest {
		return false, "You are not currently queued - nothing to do!", false
	}
	cancelledRequest := db.CancelMatchRequest(conn, user.UserId)
	if cancelledRequest {
		return true, fmt.Sprintf("%s dequeued successfully.", user.DiscordUserName), true
	} else {
		return false, "An unidentified technical issue happened while trying to dequeue. Please try again and if the problem persists contact admin and we will hit the TV until it works.", false
	}
}
