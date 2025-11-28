package models

import "context"

type TaskLinksRequest struct {
	Links []string `json:"links" binding:"required,min=1"`
}

type Task struct {
	LinksNum int
	TaskLinksRequest
}

type TaskResponse struct {
	Links    map[string]string `json:"links"`
	Err      error             `json:"error,omitempty"`
	LinksNum int               `json:"links_num"`
}

type TaskWorker struct {
	Task
	Ctx             context.Context
	ResponseChannel chan TaskResponse
}

type TaskReportRequest struct {
	LinksNum []int `json:"links_num" binding:"required,min=1"`
}
