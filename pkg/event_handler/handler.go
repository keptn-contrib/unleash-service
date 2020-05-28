package event_handler

import (
	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

type EventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (EventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptn.ActionTriggeredEventType:
		return &ActionTriggeredHandler{
			Logger: logger,
			Event:  event,
		}, nil
	default:
		return nil, nil
	}
}
