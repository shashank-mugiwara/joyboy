package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shashank-mugiwara/joyboy/config"
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/dkrclient"
	"github.com/shashank-mugiwara/joyboy/migrate"
	taskapi "github.com/shashank-mugiwara/joyboy/pkg/task-api"
	"github.com/shashank-mugiwara/joyboy/router"
	"github.com/shashank-mugiwara/joyboy/scheduler"
	"github.com/shashank-mugiwara/joyboy/task"
	"github.com/shashank-mugiwara/joyboy/worker"
	"gorm.io/gorm"
)

func HandleRoutes(r *echo.Echo, w worker.Worker, db *gorm.DB) {
	taskapi.NewHandler(w, db).InitRoutes(r)
}

func main() {
	r := router.New()
	r.Use(middleware.Recover())

	// Read Configs
	config.SetUp("")

	database.InitDb()
	migrate.AutoMigrate()
	dkrclient.InitPlainDockerClient()

	w := worker.Worker{
		Queue: queue.New(),
		DB:    database.GetDb(),
	}

	HandleRoutes(r, w, database.GetDb())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	go func() {
		if err := r.Start(":8070"); err != nil && err != http.ErrServerClosed {
			r.Logger.Fatal("shutting down the server")
		}
	}()

	r.Logger.Info("Worker initialized and are Ready...")
	go worker.RunTasks(w)
	r.Logger.Info("Workers are now listening to their worker queue.")

	r.Logger.Info("Running background scheduler")
	go scheduler.InitBackgroundScheduler()
	r.Logger.Info("Initiated background scheduler.")

	sig := <-signalCh
	log.Printf("Received signal: %v\n", sig)
	log.Printf("Stopping all running containers gracefully")

	task.StopAllTasks()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := r.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v\n", err)
	}

	log.Println("Server shutdown gracefully")
}
