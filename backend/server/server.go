package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"net/http"
)

func InitServer() {
	r := gin.Default()
	r.GET("/ping", handlePing)
	r.Run(":8080")

}

func handlePing(c *gin.Context) {
	fmt.Println("made it into ping!")
	c.JSON(http.StatusOK, gin.H{
		"ping": "pong",
	})
}
