package main

import (
	"context"
	"log"
	"os"

	"github.com/EventStore/EventStore-Client-Go/esdb"
	"github.com/joho/godotenv"
	"github.com/mxbikes/mxbikesclient.modManager/client"
	"github.com/mxbikes/mxbikesclient.modManager/projection"
	"github.com/mxbikes/mxbikesclient.modManager/repository"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	userID        = getEnv("USERID")
	esdbUrl       = getEnv("ESDB")
	modFolderPath = getEnv("MOD_FOLDER")
	serviceModUrl = getEnv("SERVICE_MOD_URL")
)

func main() {
	logger := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.DebugLevel,
		Formatter: &prefixed.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceFormatting: true,
		},
	}

	/* Services */
	serviceMod := client.InitModServiceClient(serviceModUrl)

	/* Database */
	repo := repository.NewRepository(modFolderPath)

	settings, err := esdb.ParseConnectionString(esdbUrl)
	if err != nil {
		logger.WithFields(logrus.Fields{"prefix": "EVENTSTOREDB"}).Fatalf("err: {%v}", err)
	}
	eventStoreDB, err := esdb.NewClient(settings)

	projectionHandler := projection.New(*logger, eventStoreDB, repo, serviceMod, userID)
	err = projectionHandler.Subscribe(context.Background())
	if err != nil {
		logger.WithFields(logrus.Fields{"prefix": "POSTGRES_SUBSCRIPTION"}).Fatalf("err: {%v}", err)
	}
}

func getEnv(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv(key)
}
