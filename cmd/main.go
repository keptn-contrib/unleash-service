package main

import (
	"log"
	"os"

	"github.com/keptn-contrib/unleash-service/pkg/event_handler"
	"github.com/keptn/go-utils/pkg/sdk"
	"github.com/sirupsen/logrus"
)

const getActionTriggeredEventType = "sh.keptn.event.action.triggered"
const serviceName = "unleash-service"
const envVarLogLevel = "LOG_LEVEL"

func main() {
	if os.Getenv(envVarLogLevel) != "" {
		logLevel, err := logrus.ParseLevel(os.Getenv(envVarLogLevel))
		if err != nil {
			logrus.WithError(err).Error("could not parse log level provided by 'LOG_LEVEL' env var")
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logLevel)
		}
	}

	log.Fatal(sdk.NewKeptn(
		serviceName,
		sdk.WithTaskHandler(
			getActionTriggeredEventType,
			event_handler.NewActionTriggeredHandler()),
		sdk.WithLogger(logrus.New()),
	).Start())
}
