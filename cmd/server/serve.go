package main

import (
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sleep2death/goink"
)

type editor struct {
	Value string `json:"value" binding:"required"`
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	// on something changed in user's editor
	r.POST("/editor/onchange", func(c *gin.Context) {
		var json editor
		if err := c.ShouldBindJSON(&json); err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.JSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		}

		story := goink.Default()

		if err := story.Parse(json.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"errors": []error{err}})
			return
		}

		// may return multiple errors
		if errs := story.PostParsing(); errs != nil {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{"errors": errs})
			return
		}

		ctx := goink.NewContext()
		sec, err := story.Resume(ctx)

		if err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.JSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": sec})
	})

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err := r.Run(":9090"); err != nil {
		os.Exit(-1)
	}
}
