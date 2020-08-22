package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type editor struct {
	Value string `json:"value" binding:"required"`
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())
	r.POST("/api/onchange", func(c *gin.Context) {
		var json editor
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err := r.Run(":9090"); err != nil {
		os.Exit(-1)
	}
}
