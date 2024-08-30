package handlers

import (
	"github.com/sirupsen/logrus"
	"os"
)

var log = logrus.New()

func InitLogger() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to log to file, using default stderr: %v", err)
	}
	log.SetOutput(file)
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.InfoLevel)
}
