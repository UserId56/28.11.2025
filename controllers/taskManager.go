package controllers

import (
	"context"
	"net/http"
	"sync"

	"25.11.2025/models"
	"25.11.2025/worker"
)

type TaskManager struct {
	TaskQueue chan models.TaskWorker
	taskList  map[int]models.Task
	count     int
	Wg        sync.WaitGroup
	Mu        sync.Mutex
	isRunning bool
}

func NewTaskManager(queueCapacity int) *TaskManager {
	return &TaskManager{
		TaskQueue: make(chan models.TaskWorker, queueCapacity),
		taskList:  make(map[int]models.Task),
	}
}

func (tm *TaskManager) AddTask(req models.TaskRequest, respChan chan models.TaskResponse, ctx context.Context) {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()
	tm.count++
	id := tm.count
	task := models.Task{
		LinksNum: id,
	}
	task.Links = req.Links
	tm.taskList[id] = task
	work := models.TaskWorker{
		Task:            task,
		Ctx:             ctx,
		ResponseChannel: respChan,
	}
	tm.TaskQueue <- work
}

func (tm *TaskManager) GetTask(id int) (models.Task, bool) {
	task, exists := tm.taskList[id]
	return task, exists
}

func (tm *TaskManager) IsWork() bool {
	return tm.isRunning
}

func (tm *TaskManager) Start() {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()
	if tm.isRunning {
		return
	}
	tm.isRunning = true
	numWorkers := 100
	for i := range numWorkers {
		tm.Wg.Add(1)
		go worker.New(&tm.Wg, tm.TaskQueue, i)
	}
}

func (tm *TaskManager) Stop(serv *http.Server) error {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()
	tm.isRunning = false
	close(tm.TaskQueue)
	tm.Wg.Wait()
	return serv.Shutdown(context.Background())
}
