package migrate

import (
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/task"
)

func AutoMigrate() error {
	err := database.GetDb().AutoMigrate(&task.Task{})
	return err
}
