package database

import (
	"fmt"
	"log"

	"github.com/shashank-mugiwara/joyboy/config"
	"github.com/shashank-mugiwara/joyboy/pkg/presetup"
	"github.com/shashank-mugiwara/joyboy/utils"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var gormDb *gorm.DB

func GetDb() *gorm.DB {
	return gormDb
}

func InitDb() {
	dbType := config.DatabaseSetting.DbType
	dbName := config.DatabaseSetting.DbName
	dbPort := config.DatabaseSetting.DbPort
	dbUsername := config.DatabaseSetting.DbUsername
	dbPassword := config.DatabaseSetting.DbPassword
	dbTimeZone := config.DatabaseSetting.DbTimezone
	dbConnSslEnabled := config.DatabaseSetting.SSLMode

	if utils.IsBlank(dbType) {
		_, ok := presetup.DockerImageToRepoMap[dbType]
		if !ok {
			log.Fatalln("DbType given is not supported. Exiting ...")
		}
	}

	if dbType == "sqlite" {
		handleDatabaseInit(dbType, "", 0, "", "", "", "")
	} else {
		if utils.IsBlank(dbName) {
			log.Fatalln("DbName is missing, cannot configure db without it. Exiting ...")
		}

		if utils.IsBlank(dbUsername) {
			log.Fatalln("DbUsername is missing, cannot configure db without it. Exiting ...")
		}

		if dbPort == 0 {
			log.Fatalln("DbPort is missing, cannot configure db without it. Exiting ...")
		}

		if utils.IsBlank(dbPassword) {
			log.Fatalln("DbPassword is missing, cannot configure db without it. Exiting ...")
		}

		if utils.IsBlank(dbTimeZone) {
			log.Println("Timezone not specified, falling back to default timezone: Asia/Calcutta")
			dbTimeZone = "Asia/Calcutta"
		}

		if dbConnSslEnabled != "enabled" {
			log.Println("SSL not properly specified. Going with disabled ssl by default")
			dbConnSslEnabled = "disabled"
		}

		handleDatabaseInit(dbType, dbName, dbPort, dbUsername, dbPassword, dbTimeZone, dbConnSslEnabled)
	}
}

func handleDatabaseInit(dbType string, dbName string, dbPort int, dbUsername string, dbPassword string, timezone string, sslEnabled string) {

	var db *gorm.DB
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		dbUsername, dbPassword, dbName, dbPort, sslEnabled, timezone)

	switch dbType {
	case "postgres14":
		postgresDb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalln("Failed to open DB connection. Exiting ...")
		}
		db = postgresDb
	default:
		sqlLiteGormDb, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
		if err != nil {
			log.Fatalln("Failed to open DB connection. Exiting ...")
		}
		db = sqlLiteGormDb
	}

	gormDb = db
}
