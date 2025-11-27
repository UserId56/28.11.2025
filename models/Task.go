package models

import "context"

type TaskRequest struct {
	Links []string `json:"links"`
}

type Task struct {
	LinksNum int
	TaskRequest
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
