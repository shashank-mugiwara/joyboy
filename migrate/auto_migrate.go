package migrate

import (
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/task"
)

func AutoMigrate() error {
	err := database.GetDb().AutoMigrate(&task.Task{})
	if err != nil {
		return err
	}

	err = database.GetDb().AutoMigrate(&task.Local{})
	if err != nil {
		return err
	}

	return err
}
