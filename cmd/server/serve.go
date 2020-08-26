package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/sleep2death/goink"
)

type editor struct {
	Value string `json:"value" binding:"required"`
	Uuid  string `json:"uuid"`
}

type user struct {
	id    string
	story *goink.Story
	ctx   *goink.Context
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	// Create a cache with a default expiration time of 30 minutes, and which
	// purges expired items every 60 minutes
	cc := cache.New(30*time.Minute, 60*time.Minute)

	// when something changed in user's editor
	r.POST("/editor/onchange", getChangeHandler(cc))

	// when user select an option in review panel
	r.POST("/editor/onchange", getChooseHandler(cc))

	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err := r.Run(":9090"); err != nil {
		os.Exit(-1)
	}
}

func getChangeHandler(cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var json editor
		// bind json
		if err := c.ShouldBindJSON(&json); err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.JSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		}

		// create uuid if not exist
		var id string
		if json.Uuid == "" {
			id = uuid.NewV4().String()
		} else if _, err := uuid.FromString(json.Uuid); err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		} else {
			// get id from client
			id = json.Uuid
		}

		var store *user

		if u, found := cc.Get(id); found {
			store = u.(*user)
			if store == nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			log.Info("An [Exited] user has updated his(hers) ink")
			// fmt.Println("old user", store.ctx.Current)
		} else { // set new user
			store = &user{id: id, ctx: goink.NewContext()}
			cc.Set(id, store, cache.DefaultExpiration)
			// fmt.Println("new user")
			log.Info("A [New] user has updated his(hers) ink")
		}

		// create story
		story := goink.Default()
		if err := story.Parse(json.Value); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": []error{err}})
			return
		}

		// return multiple errors after post parsing
		if errs := story.PostParsing(); errs != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": errs})
			return
		}

		store.story = story
		store.ctx = goink.NewContext()

		sec, err := story.Resume(store.ctx)

		// TODO: resume error wrap with line number
		if err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.JSON(http.StatusOK, gin.H{"errors": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"section": sec, "uuid": id})
	}
}

func getChooseHandler(cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		var json editor
		// bind json
		if err := c.ShouldBindJSON(&json); err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.JSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		}

		// create uuid if not exist
		var id string
		if json.Uuid == "" {
			msg := (goink.ErrInk{}).Wrap(errors.New("empty user id"))
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		} else if _, err := uuid.FromString(json.Uuid); err != nil {
			msg := (goink.ErrInk{}).Wrap(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": msg})
			return
		}

		// get id from client
		id = json.Uuid
		var store *user

		if u, found := cc.Get(id); found {
			store = u.(*user)
			if store == nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			log.Info("An user has selected his(hers) ink's option")
			// fmt.Println("old user", store.ctx.Current)
		} else { // set new user
			store = &user{id: id, ctx: goink.NewContext()}
			cc.Set(id, store, cache.DefaultExpiration)
			// fmt.Println("new user")
			log.Info("A [New] user has updated his(hers) ink")
		}
	}
}
