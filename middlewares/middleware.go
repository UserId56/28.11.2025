package middlewares

import (
	"25.11.2025/controllers"
	"github.com/gin-gonic/gin"
)

func ServiceWork(tm *controllers.TaskManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !tm.IsWork() {
			c.JSON(503, gin.H{"error": "Сервис не доступен"})
			c.Abort()
			return
		}
		c.Next()
	}
}
