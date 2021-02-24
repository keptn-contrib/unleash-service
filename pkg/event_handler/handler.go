package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
)

type EventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (EventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptnv2.GetTriggeredEventType(keptnv2.ActionTaskName):
		return &ActionTriggeredHandler{
			Logger: logger,
			Event:  event,
		}, nil
	default:
		return nil, nil
	}
}
