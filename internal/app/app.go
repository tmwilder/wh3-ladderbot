package app

import (
	"github.com/gin-gonic/gin"
)

func App() {
	r := gin.Default()
	r.GET("/view", helloHandler)

	err := r.Run()
	if err != nil {
		panic("Could not start web server: " + err.Error())
	}
}
