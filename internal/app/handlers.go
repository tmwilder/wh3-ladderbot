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
	appConfig := config.GetAppConfig()
	requestKey, foundKey := c.GetQuery("admin_key")
	if !foundKey {
		c.JSON(http.StatusUnauthorized, "Must supply query param admin key.")
		return
	}
	if requestKey != appConfig.AdminKey {
		c.JSON(http.StatusUnauthorized, "Must supply correct query param admin key.")
		return
	}
	commands.InstallGlobalCommands(config.GetAppConfig())
	c.JSON(http.StatusOK, nil)
}
