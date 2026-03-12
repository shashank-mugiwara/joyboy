package main

import (
	"context"
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
	"github.com/shashank-mugiwara/joyboy/pkg/logging"
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

	// Initialize structured logger
	logger := logging.GetDefaultLogger()

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
			logger.Fatal("shutting down the server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	logger.Info("Worker initialized and are Ready", nil)
	go worker.RunTasks(w)
	logger.Info("Workers are now listening to their worker queue", nil)

	logger.Info("Running background scheduler", nil)
	go scheduler.InitBackgroundScheduler()
	logger.Info("Initiated background scheduler", nil)

	sig := <-signalCh
	logger.Info("Received signal", map[string]interface{}{
		"signal": sig.String(),
	})
	logger.Info("Stopping all running containers gracefully", nil)

	task.StopAllTasks()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := r.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Server shutdown gracefully", nil)
}
