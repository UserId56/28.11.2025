package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"25.11.2025/internal"
	"25.11.2025/models"
)

func New(wg *sync.WaitGroup, tq chan models.TaskWorker, num int, writeCache func(status models.LinkStatus), getCache func(link string) (models.LinkStatus, bool)) {
	defer wg.Done()
	fmt.Printf("Worker #%d запущен\n", num)
	for task := range tq {
		startTime := time.Now()
		fmt.Printf("Worker #%d Выолняю задачу №%d\n", num, task.Task.LinksNum)
		var results models.TaskResponse
		results.LinksNum = task.LinksNum
		results.Links = make(map[string]string)
		var wgCh sync.WaitGroup
		CheckCtx, cancelCheckCtx := context.WithCancel(task.Ctx)
		resChan := make(chan models.LinkStatus, len(task.Links))
		for _, link := range task.Links {
			if cachedStatus, exists := getCache(link); exists {
				fmt.Printf("Worker #%d: Используется кэш для %s\n", num, link)
				results.Links[link] = cachedStatus.Status
				continue
			}
			wgCh.Add(1)
			go func(linkUrl string) {
				defer wgCh.Done()
				result, err := internal.CheckUrl(linkUrl, CheckCtx)
				if err != nil {
					resChan <- models.LinkStatus{Link: linkUrl, Status: "error", Error: err}
					return
				}
				resChan <- models.LinkStatus{Link: linkUrl, Status: result, Error: nil}
			}(link)
		}
		go func() {
			wgCh.Wait()
			close(resChan)
		}()
		for res := range resChan {
			if res.Error != nil && results.Err == nil {
				//results.Err = res.Error
				cancelCheckCtx()
			}
			results.Links[res.Link] = res.Status
			writeCache(res)
		}
		cancelCheckCtx()
		select {
		case task.ResponseChannel <- results:
			workTime := time.Since(startTime)
			fmt.Println("Задача №", task.Task.LinksNum, "выполнена за", workTime, "секунд")
		case <-task.Ctx.Done():
			fmt.Println("Контекст задачи №", task.Task.LinksNum, "отменен до отправки результата")
		}
	}
}
