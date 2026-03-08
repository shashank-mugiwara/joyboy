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

const (
	serverPort                   = ":8070"
	shutdownTimeout              = 30 * time.Second
	msgServerShutdown            = "shutting down the server"
	msgWorkerInitialized         = "Worker initialized and are Ready..."
	msgWorkersListening          = "Workers are now listening to their worker queue."
	msgSchedulerRunning          = "Running background scheduler"
	msgSchedulerInitiated        = "Initiated background scheduler."
	msgReceivedSignal            = "Received signal: %v\n"
	msgStoppingContainers        = "Stopping all running containers gracefully"
	msgServerShutdownFailed      = "Server shutdown failed: %v\n"
	msgServerShutdownGracefully  = "Server shutdown gracefully"
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
		if err := r.Start(serverPort); err != nil && err != http.ErrServerClosed {
			r.Logger.Fatal(msgServerShutdown)
		}
	}()

	r.Logger.Info(msgWorkerInitialized)
	go worker.RunTasks(w)
	r.Logger.Info(msgWorkersListening)

	r.Logger.Info(msgSchedulerRunning)
	go scheduler.InitBackgroundScheduler()
	r.Logger.Info(msgSchedulerInitiated)

	sig := <-signalCh
	log.Printf(msgReceivedSignal, sig)
	log.Printf(msgStoppingContainers)

	task.StopAllTasks()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown the server gracefully
	if err := r.Shutdown(ctx); err != nil {
		log.Fatalf(msgServerShutdownFailed, err)
	}

	log.Println(msgServerShutdownGracefully)
}
