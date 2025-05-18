// Main entrypoint for the interactions service. Sets up Gin, config, routes, and runs the server.

package main

import (
	"log"
	"os"

	"way-d-interactions/config"
	"way-d-interactions/models"
	"way-d-interactions/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	if err := config.DB.AutoMigrate(
		&models.Like{},
		&models.Dislike{},
		&models.Match{},
		&models.Message{},
		&models.Block{},
	); err != nil {
		log.Fatalf("Migration error: %v", err)
	}

	r := gin.Default()
	routes.RegisterRoutes(r)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("Way-d Interactions service running on port %s", port)
	r.Run(":" + port)
}
