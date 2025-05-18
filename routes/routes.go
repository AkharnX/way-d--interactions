package routes

import (
	"way-d-interactions/config"
	"way-d-interactions/controllers"
	"way-d-interactions/middleware"
	"way-d-interactions/models"

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
