package routes

import (
	"25.11.2025/controllers"
	"25.11.2025/handlers"
	"25.11.2025/middlewares"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine, tm *controllers.TaskManager) {
	r.Use(middlewares.ServiceWork(tm))
	r.POST("/status", handlers.GetStatus(tm))
}
