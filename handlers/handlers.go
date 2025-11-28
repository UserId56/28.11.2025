package handlers

import (
	"context"
	"fmt"
	"time"

	"25.11.2025/controllers"
	"25.11.2025/models"
	"25.11.2025/services"
	"github.com/gin-gonic/gin"
)

func GetStatus(tm *controllers.TaskManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.TaskLinksRequest
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
		tm.AddTask(req, respChan, ctx, false)
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

func GetNumLinks(tm *controllers.TaskManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var task models.TaskLinksRequest
		var req models.TaskReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Не верные данные или формат"})
			return
		}
		for _, linkId := range req.LinksNum {
			taskItem, exists := tm.GetTask(linkId)
			if !exists {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Задача с номером %d не найдена", linkId)})
				return
			}
			task.Links = append(task.Links, taskItem.Links...)
		}
		fmt.Printf("%+v", task)
		respChan := make(chan models.TaskResponse, 1)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		tm.AddTask(task, respChan, ctx, true)
		defer cancel()
		select {
		case resp := <-respChan:
			if resp.Err != nil {
				c.JSON(500, gin.H{"error": "Ошибка сервере"})
				return
			}
			c.Header("Content-Type", "application/pdf")
			c.Header("Content-Disposition", "attachment; filename=\"status_report.pdf\"")
			err := services.PDFGeneration(c.Writer, resp)
			if err != nil {
				c.JSON(500, gin.H{"error": "Ошибка при генерации отчета"})
				return
			}
		case <-ctx.Done():
			c.Status(504)
		}
	}
}

func Test(tm *controllers.TaskManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Printf("%+v\n", tm)
	}
}
