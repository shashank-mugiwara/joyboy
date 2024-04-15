package utils

import (
	"encoding/json"

	"github.com/shashank-mugiwara/joyboy/task"
)

func UnmarshallTask(data string) task.Task {
	t := task.Task{}
	json.Unmarshal([]byte(data), &t)
	return t
}

func MarshallTask(t task.Task) string {
	b, err := json.Marshal(t)
	if err != nil {
		return "{}"
	}
	return string(b)
}
