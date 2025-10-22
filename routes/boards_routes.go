package routes

import (
	"tt-board-api/controllers"

	"github.com/gin-gonic/gin"
)

func TaskBoardsRoutes(router *gin.Engine) {
	board := router.Group("/boards")
	{
		board.GET("/", controllers.GetBoards)
		board.GET("/tasks", controllers.GetBoardTasks)
	}
}
