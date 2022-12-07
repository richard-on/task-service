package config

import (
	"fmt"
	"github.com/richard-on/task-service/pkg/logger"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

var LogInfo struct {
	Output        string
	Level         zerolog.Level
	File          string
	ConsoleWriter bool
}

var SentryInfo struct {
	DSN string
	TSR float64
}

var DbInfo struct {
	Name     string
	Host     string
	Port     string
	User     string
	Password string
}
var DbConnString string

var MongoDbName string
var MongoCollection string

var Env string
var GoDotEnv bool
var FiberPrefork bool
var MaxCPU int

func Init(log logger.Logger) {
	var err error

	Env = os.Getenv("ENV")

	GoDotEnv, err = strconv.ParseBool(os.Getenv("GODOTENV"))
	if err != nil {
		log.Infof("GODOTENV init: %v", err)
	}

	FiberPrefork, err = strconv.ParseBool(os.Getenv("FIBER_PREFORK"))
	if err != nil {
		log.Infof("FIBER_PREFORK init: %v", err)
	}

	MaxCPU, err = strconv.Atoi(os.Getenv("MAX_CPU"))
	if err != nil {
		log.Infof("MAX_CPU init: %v", err)
	}

	LogInfo.Output = os.Getenv("LOG_OUTPUT")

	LogInfo.Level, err = zerolog.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Infof("LOG_LEVEL init: %v", err)
	}

	LogInfo.File = os.Getenv("LOG_FILE")

	LogInfo.ConsoleWriter, err = strconv.ParseBool(os.Getenv("LOG_CW"))
	if err != nil {
		log.Infof("LOG_CW init: %v", err)
	}

	SentryInfo.DSN = os.Getenv("SENTRY_DSN")

	SentryInfo.TSR, err = strconv.ParseFloat(os.Getenv("SENTRY_TSR"), 64)
	if err != nil {
		log.Infof("SENTRY_TSR init: %v", err)
	}

	DbInfo.Name = os.Getenv("DB_NAME")
	DbInfo.Host = os.Getenv("DB_HOST")
	DbInfo.Port = os.Getenv("DB_PORT")
	DbInfo.User = os.Getenv("DB_USER")
	DbInfo.Password = os.Getenv("DB_PASSWORD")

	DbConnString = fmt.Sprintf("%s://%s:%s/",
		DbInfo.Name, DbInfo.Host, DbInfo.Port)

	MongoDbName = os.Getenv("MONGO_DB")
	MongoCollection = os.Getenv("MONGO_COLLECTION")
}
