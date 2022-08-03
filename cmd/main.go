package main

import (
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
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
			event_handler.NewActionTriggeredHandler(),
			actionTriggeredFilter),
		sdk.WithLogger(logrus.New()),
	).Start())
}

func actionTriggeredFilter(keptnHandle sdk.IKeptn, event sdk.KeptnEvent) bool {
	data := &keptnv2.ActionTriggeredEventData{}
	if err := keptnv2.Decode(event.Data, data); err != nil {
		keptnHandle.Logger().Errorf("Could not parse test.triggered event: %s", err.Error())
		return false
	}

	return data.Action.Action == event_handler.ActionToggleFeature
}
