package worker

import (
	"fmt"
	"sync"

	"25.11.2025/internal"
	"25.11.2025/models"
)

func New(wg *sync.WaitGroup, tq chan models.TaskWorker, num int) {
	defer wg.Done()
	fmt.Printf("Worker #%d запущен\n", num)
	for task := range tq {
		fmt.Printf("Worker #%d Выолняю задачу №%d\n", num, task.Task.LinksNum)
		var results models.TaskResponse
		results.LinksNum = task.LinksNum
		results.Links = make(map[string]string)
		for _, link := range task.Links {
			result, err := internal.CheckUrl(link, task.Ctx)
			if err != nil {
				//	Если хоть один URL обработался с ошибкой, возвращаем ошибку
				results.Err = err
				fmt.Printf("Ошибка проверки URL %s: %v\n", link, err)
				break
			}
			results.Links[link] = result
		}
		select {
		case task.ResponseChannel <- results:
			fmt.Println("Задача №", task.Task.LinksNum, "выполнена")
		case <-task.Ctx.Done():
			fmt.Println("Контекст задачи №", task.Task.LinksNum, "отменен до отправки результата")
		}
	}
}
