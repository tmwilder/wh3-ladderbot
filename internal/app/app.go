package app

import (
	"discordbot/internal/app/discord"
	"github.com/gin-gonic/gin"
)

func App() {
	g := GetGin()
	err := g.Run()
	if err != nil {
		panic("Could not start web server: " + err.Error())
	}
}

/*
	GetGin
	Configure our gin handlers and return it. We run this separately from app startup so that it can be used both by
	local HTTP server flows and by lambda startup. Used by cmd/lambda and cmd/api.
*/
func GetGin() (g *gin.Engine) {
	g = gin.Default()
	g.POST("/commands", installSlashCommandsHandler)
	g.POST("/maps", setMapsHandler)
	g.POST("/match-requests/expire", expireMatchRequestsHandler)
	g.POST("/migrate", migrationHandler)
	g.POST("/interactions", discord.InteractionsHandler)
	g.POST("/leaderboard", updateLeaderBoardHandler)
	return g
}
