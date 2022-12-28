package app

import (
	"discordbot/internal/app/config"
	"discordbot/internal/app/discord/commands"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
	We use exclusively global commands because there are no use cases to scope to one guild.
	https://discord.com/developers/docs/interactions/application-commands#authorizing-your-application
*/
func installSlashCommandsHandler(c *gin.Context) {
	commands.InstallGlobalCommands(config.GetDiscordAppConfig())
	c.JSON(http.StatusOK, nil)
}
