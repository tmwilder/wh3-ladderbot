package app

import (
	"github.com/gin-gonic/gin"
)

func App() {
	r := gin.Default()
	r.GET("/view", helloHandler)
	r.POST("/commands", installSlashCommandsHandler)
	r.POST("/interactions", interactionsHandler)

	err := r.Run()
	if err != nil {
		panic("Could not start web server: " + err.Error())
	}
}
