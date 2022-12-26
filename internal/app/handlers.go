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
