package handlers

import (
	"context"
	"time"

	"25.11.2025/controllers"
	"25.11.2025/models"
	"github.com/gin-gonic/gin"
)

func GetStatus(tm *controllers.TaskManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.TaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Не верные данные или формат"})
			return
		}
		respChan := make(chan models.TaskResponse, 1)
		perLink := 5 * time.Second
		margin := 2 * time.Second
		capTimeout := 30 * time.Second
		overall := time.Duration(len(req.Links))*perLink + margin
		if overall > capTimeout {
			overall = capTimeout
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), overall)
		tm.AddTask(req, respChan, ctx)
		defer cancel()
		select {
		case resp := <-respChan:
			if resp.Err != nil {
				c.JSON(500, gin.H{"error": "Ошибка сервере"})
				return
			}
			c.JSON(200, resp)
		case <-ctx.Done():
			c.Status(504)
		}
	}
}
