package controllers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"25.11.2025/models"
	"25.11.2025/worker"
)

type TaskManager struct {
	TaskQueue   chan models.TaskWorker
	taskList    map[int]models.Task
	сacheStatus map[string]models.LinkStatus
	Cached      bool
	count       int
	Wg          sync.WaitGroup
	Mu          sync.Mutex
	isRunning   bool
}

func NewTaskManager(queueCapacity int, cached bool) *TaskManager {
	fmt.Println("Создание TaskManager с очередью емкостью", queueCapacity, "и кэшированием:", cached)
	return &TaskManager{
		TaskQueue:   make(chan models.TaskWorker, queueCapacity),
		taskList:    make(map[int]models.Task),
		сacheStatus: make(map[string]models.LinkStatus),
		Cached:      cached,
	}
}

func (tm *TaskManager) CacheUrlWrite(data models.LinkStatus) {
	if tm.Cached {
		data.DateCheck = time.Now()
		tm.Mu.Lock()
		tm.сacheStatus[data.Link] = data
		tm.Mu.Unlock()
	}
}

func (tm *TaskManager) CacheUrlGet(link string) (models.LinkStatus, bool) {
	data, exists := tm.сacheStatus[link]
	if !exists {
		return models.LinkStatus{}, false
	}
	if time.Since(data.DateCheck) > 1*time.Minute {
		return models.LinkStatus{}, false
	} else {
		return data, true
	}
}

func (tm *TaskManager) AddTask(req models.TaskRequest, respChan chan models.TaskResponse, ctx context.Context) {
	tm.Mu.Lock()
	tm.count++
	id := tm.count
	task := models.Task{
		LinksNum: id,
	}
	task.Links = req.Links
	tm.taskList[id] = task
	tm.Mu.Unlock()
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
	numWorkers := runtime.GOMAXPROCS(0) * 10
	for i := 0; i < numWorkers; i++ {
		tm.Wg.Add(1)
		go worker.New(&tm.Wg, tm.TaskQueue, i, func(data models.LinkStatus) {
			tm.CacheUrlWrite(data)
		}, func(link string) (models.LinkStatus, bool) {
			return tm.CacheUrlGet(link)
		})
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
