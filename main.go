package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/golang-collections/collections/queue"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/migrate"
	taskapi "github.com/shashank-mugiwara/joyboy/pkg/task-api"
	"github.com/shashank-mugiwara/joyboy/router"
	"github.com/shashank-mugiwara/joyboy/worker"
	"gorm.io/gorm"
)

func HandleRoutes(r *echo.Echo, w worker.Worker, db *gorm.DB) {
	taskapi.NewHandler(w, db).InitRoutes(r)
}

func main() {
	os.Setenv("DOCKER_API_VERSION", "1.44")
	r := router.New()
	r.Use(middleware.Recover())
	database.InitDb()
	migrate.AutoMigrate()

	w := worker.Worker{
		Queue: queue.New(),
		DB:    database.GetDb(),
	}

	HandleRoutes(r, w, database.GetDb())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Start server
	go func() {
		if err := r.Start(":8070"); err != nil && err != http.ErrServerClosed {
			r.Logger.Fatal("shutting down the server")
		}
	}()

	r.Logger.Info("Worker initialized and are Ready...")
	go worker.RunTasks(w)
	r.Logger.Info("Workers are now listening to their worker queue.")

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := r.Shutdown(ctx); err != nil {
		r.Logger.Fatal(err)
	}
}
