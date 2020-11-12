package event_handler

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnevents "github.com/keptn/go-utils/pkg/lib"
	"github.com/keptn/go-utils/pkg/lib/keptn"
)

type EventHandler interface {
	HandleEvent() error
}

func NewEventHandler(event cloudevents.Event, logger *keptn.Logger) (EventHandler, error) {
	logger.Debug("Received event: " + event.Type())
	switch event.Type() {
	case keptnevents.ActionTriggeredEventType:
		return &ActionTriggeredHandler{
			Logger: logger,
			Event:  event,
		}, nil
	default:
		return nil, nil
	}
}
