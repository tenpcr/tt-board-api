package main

import (
	"log"
	"os"
	"tt-board-api/config"
	"tt-board-api/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	config.LoadEnv()
	config.ConnectMongo()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	routes.TaskBoardsRoutes(router)
	routes.TasksRoutes(router)

	log.Printf("Server is running on port %s", config.GetEnv("PORT"))
	router.Run(":" + os.Getenv("PORT"))
}
