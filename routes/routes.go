package routes

import (
	"log"
	"time"
	"way-d-interactions/config"
	"way-d-interactions/controllers"
	"way-d-interactions/middleware"
	"way-d-interactions/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		api.POST("/like", controllers.PostLike)
		api.POST("/dislike", controllers.PostDislike)
		api.GET("/matches", controllers.GetMatches)
		api.POST("/message", controllers.PostMessage)
		api.GET("/messages/:match_id", controllers.GetMessages)
		api.POST("/block", controllers.PostBlock)
		api.GET("/blocks", controllers.GetBlocks)
		api.GET("/exclusions", controllers.GetExclusions)
	}

	r.GET("/debug/likes", func(c *gin.Context) {
		db := config.GetDB()
		var likes []models.Like
		db.Find(&likes)
		c.JSON(200, likes)
	})
	r.GET("/debug/matches", func(c *gin.Context) {
		db := config.GetDB()
		var matches []models.Match
		db.Find(&matches)
		c.JSON(200, matches)
	})
	r.GET("/debug/blocks", func(c *gin.Context) {
		db := config.GetDB()
		var blocks []models.Block
		db.Find(&blocks)
		c.JSON(200, blocks)
	})
	r.POST("/debug/clear", func(c *gin.Context) {
		db := config.GetDB()
		db.Exec("DELETE FROM messages")
		db.Exec("DELETE FROM matches")
		db.Exec("DELETE FROM likes")
		db.Exec("DELETE FROM dislikes")
		db.Exec("DELETE FROM blocks")
		c.JSON(200, gin.H{"status": "cleared"})
	})
}

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8083"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(func(c *gin.Context) {
		log.Printf("[INFO] %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Next()
	})

	return r
}
