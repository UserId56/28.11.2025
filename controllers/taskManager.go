package controllers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"25.11.2025/models"
	"25.11.2025/services"
	"25.11.2025/worker"
)

type TaskManagerConfig struct {
	QueueCapacity int
	Cached        bool
	DataFileName  string
}

type TaskManager struct {
	TaskQueue    chan models.TaskWorker
	taskList     map[int]models.Task
	сacheStatus  map[string]models.LinkStatus
	DataFileName string
	Cached       bool
	count        int
	Wg           sync.WaitGroup
	Mu           sync.Mutex
	isRunning    bool
}

func NewTaskManager(config TaskManagerConfig) *TaskManager {
	//fmt.Println("Создание TaskManager с очередью емкостью", config.QueueCapacity, "и кэшированием:", config.Cached)
	return &TaskManager{
		TaskQueue:    make(chan models.TaskWorker, config.QueueCapacity),
		taskList:     make(map[int]models.Task),
		сacheStatus:  make(map[string]models.LinkStatus),
		DataFileName: config.DataFileName,
		Cached:       config.Cached,
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

func (tm *TaskManager) AddTask(req models.TaskLinksRequest, respChan chan models.TaskResponse, ctx context.Context, isReport bool) {
	var task models.Task
	task.Links = req.Links
	if !isReport {
		tm.Mu.Lock()
		tm.count++
		task.LinksNum = tm.count
		tm.taskList[task.LinksNum] = task
		tm.Mu.Unlock()
	}
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
	if tm.isRunning {
		return
	}
	readData, err := services.ReadDataRequest(tm.DataFileName)
	if err == nil {
		tm.taskList = readData
		tm.count = len(readData)
		fmt.Println("Загружено задач из файла:", tm.DataFileName)
	} else {
		fmt.Println("Не удалось загрузить задачи из файла:", err)
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
	tm.isRunning = false
	close(tm.TaskQueue)
	tm.Wg.Wait()
	err := services.SaveDataRequest(tm.taskList, tm.DataFileName)
	if err != nil {
		fmt.Println("Ошибка сохранения данных в файл:", err)
	} else {
		fmt.Println("Данные успешно сохранены в файл:", tm.DataFileName)
	}
	return serv.Shutdown(context.Background())
}
