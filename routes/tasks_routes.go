package routes

import (
	"tt-board-api/controllers"

	"github.com/gin-gonic/gin"
)

func TasksRoutes(router *gin.Engine) {
	task := router.Group("/tasks")
	{
		task.GET("/", controllers.GetTasks)
		task.PUT("/progress/update/:id", controllers.UpdateTaskProgressById)
		task.GET("/types", controllers.GetTaskTypes)
		task.GET("/:id", controllers.GetTaskById)
		task.PUT("/:id", controllers.UpdateTaskById)
		task.DELETE("/:id", controllers.DeleteTaskById)
		task.POST("/add", controllers.PostAddTask)
		task.GET("/boards", controllers.GetTaskBoards)
	}

}
