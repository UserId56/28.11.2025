package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"25.11.2025/controllers"
	"25.11.2025/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	tm := controllers.NewTaskManager(runtime.GOMAXPROCS(0)*10, true)
	tm.Start()
	routes.Init(r, tm)
	server := &http.Server{
		Addr:    ":3000",
		Handler: r,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка запуска сервера: %s\n", err)
		}
	}()
	fmt.Println("Программа запущена. Нажмите Ctrl+C для завершения.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Завершение программы...")
	if err := tm.Stop(server); err != nil {
		fmt.Printf("Ошибка остановки сервера: %s\n", err)
		return
	}
	fmt.Println("Сервер успешно остановлен.")
}
