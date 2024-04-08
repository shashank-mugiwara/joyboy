package taskapi

import "github.com/shashank-mugiwara/joyboy/worker"

type Handler struct {
	worker worker.Worker
}

func NewHandler(w worker.Worker) *Handler {
	return &Handler{
		worker: w,
	}
}
