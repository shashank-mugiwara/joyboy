package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var gormDb *gorm.DB

func GetDb() *gorm.DB {
	return gormDb
}

func InitDb() {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatalln("Failed to open DB connection. Exiting ...")
	}

	gormDb = db
}
