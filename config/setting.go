package config

import (
	"log"

	"gopkg.in/ini.v1"
)

type Application struct {
	RunType string
	Port    string
}

var ApplicationSetting = &Application{}

type Database struct {
	DbType     string
	DbPort     int
	DbUsername string
	DbPassword string
	DbName     string
	DbTimezone string
	SSLMode    string
}

var DatabaseSetting = &Database{}

var cfg *ini.File

func SetUp(path string) {
	var err error
	var tempCfg *ini.File

	if path != "" {
		tempCfg, err = ini.Load(path)
	} else {
		tempCfg, err = ini.Load("config.ini")
	}

	cfg = tempCfg
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'config.ini': %v", err)
	}

	mapTo("application", ApplicationSetting)
	mapTo("db", DatabaseSetting)
}

func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}
