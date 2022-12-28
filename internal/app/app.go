package app

import (
	"discordbot/internal/app/discord"
	"github.com/gin-gonic/gin"
)

func App() {
	r := gin.Default()
	r.POST("/commands", installSlashCommandsHandler)
	r.POST("/interactions", discord.InteractionsHandler)

	err := r.Run()
	if err != nil {
		panic("Could not start web server: " + err.Error())
	}
}
