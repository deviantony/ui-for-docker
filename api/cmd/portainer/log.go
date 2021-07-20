package main

import (
	"log"

	"github.com/sirupsen/logrus"
)

func configureLogger(json *bool) {
	logger := logrus.New() // logger is to implicitly substitute stdlib's log
	log.SetOutput(logger.Writer())

	if json != nil && *json == true {
		logger.SetFormatter(&logrus.JSONFormatter{})
		// configure explicit logrus usage
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
		// configure explicit logrus usage
		logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	}

	logger.SetLevel(logrus.DebugLevel)
	logrus.SetLevel(logrus.DebugLevel)
}
