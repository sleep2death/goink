package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})

	})

	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.LoadHTMLGlob("templates/*.tmpl")

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "404.tmpl", gin.H{
			"title": "GOINK 0.0.3-alpha",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "GOINK 0.0.3-alpha",
		})
	})

	r.Run(":9090") // listen and serve on 0.0.0.0:9090 (for windows "localhost:8080")
}
