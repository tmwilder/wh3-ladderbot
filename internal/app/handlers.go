package app

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type helloResponse struct {
	Message string
}

func helloHandler(c *gin.Context) {
	c.JSON(http.StatusOK, helloResponse{Message: "Hello World"})
}

/*
	We use exclusively global commands because there are no use cases to scope to one guild.
	https://discord.com/developers/docs/interactions/application-commands#authorizing-your-application
*/
func installSlashCommandsHandler(c *gin.Context) {
	installGlobalCommands(GetDiscordAppConfig())
	c.JSON(http.StatusOK, nil)
}
