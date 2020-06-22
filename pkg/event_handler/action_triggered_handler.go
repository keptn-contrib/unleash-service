package event_handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	keptn "github.com/keptn/go-utils/pkg/lib"
)

const ActionToggleFeature = "toggle-feature"

type ActionTriggeredHandler struct {
	Logger keptn.LoggerInterface
	Event  cloudevents.Event
}

func (eh ActionTriggeredHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	actionTriggeredEvent := &keptn.ActionTriggeredEventData{}

	err := eh.Event.DataAs(actionTriggeredEvent)
	if err != nil {
		eh.Logger.Error("feature toggle remediation action not well formed: " + err.Error())
		return errors.New("feature toggle remediation action not well formed: " + err.Error())
	}

	if actionTriggeredEvent.Action.Action != ActionToggleFeature {
		eh.Logger.Info("Received unknown action: " + actionTriggeredEvent.Action.Action + ". Exiting")
		return nil
	}

	// Send action.started event
	if sendErr := eh.sendEvent(keptn.ActionStartedEventType, eh.getActionStartedEvent(*actionTriggeredEvent)); sendErr != nil {
		eh.Logger.Error(sendErr.Error())
		return errors.New(sendErr.Error())
	}

	values, ok := actionTriggeredEvent.Action.Value.(map[string]interface{})

	if !ok {
		eh.Logger.Error("Could not parse action.value")
		err = eh.sendEvent(keptn.ActionFinishedEventType,
			eh.getActionFinishedEvent(keptn.ActionResultPass, keptn.ActionStatusErrored, *actionTriggeredEvent))
		return errors.New("Could not parse action.value")
	}

	for feature, value := range values {
		if _, ok := value.(string); !ok {
			eh.Logger.Error("Value property of feature toggle remediation action not valid. It must be set: TOGGLENAME:\"on\" or TOGGLENAME:\"off\"")
			return errors.New("Value property of feature toggle remediation action not valid. It must be set: TOGGLENAME:\"on\" or TOGGLENAME:\"off\"")
		}
		err = toggleFeature(feature, value.(string))
		if err != nil {
			eh.Logger.Error("Could not set feature " + feature + " to value " + value.(string) + ": " + err.Error())
			sendErr := eh.sendEvent(keptn.ActionFinishedEventType,
				eh.getActionFinishedEvent(keptn.ActionResultPass, keptn.ActionStatusErrored, *actionTriggeredEvent))
			if sendErr != nil {
				eh.Logger.Error("could not send action-finished event: " + err.Error())
				return err
			}
			return err
		}
	}

	err = eh.sendEvent(keptn.ActionFinishedEventType,
		eh.getActionFinishedEvent(keptn.ActionResultPass, keptn.ActionStatusSucceeded, *actionTriggeredEvent))
	if err != nil {
		eh.Logger.Error("could not send action-finished event: " + err.Error())
		return err
	}
	return nil
}

func (eh ActionTriggeredHandler) getActionFinishedEvent(result keptn.ActionResultType, status keptn.ActionStatusType,
	actionTriggeredEvent keptn.ActionTriggeredEventData) keptn.ActionFinishedEventData {

	return keptn.ActionFinishedEventData{
		Project: actionTriggeredEvent.Project,
		Service: actionTriggeredEvent.Service,
		Stage:   actionTriggeredEvent.Stage,
		Action: keptn.ActionResult{
			Result: result,
			Status: status,
		},
		Labels: actionTriggeredEvent.Labels,
	}
}

func (eh ActionTriggeredHandler) getActionStartedEvent(actionTriggeredEvent keptn.ActionTriggeredEventData) keptn.ActionStartedEventData {

	return keptn.ActionStartedEventData{
		Project: actionTriggeredEvent.Project,
		Service: actionTriggeredEvent.Service,
		Stage:   actionTriggeredEvent.Stage,
		Labels:  actionTriggeredEvent.Labels,
	}
}

func (eh ActionTriggeredHandler) sendEvent(eventType string, data interface{}) error {
	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{
		EventBrokerURL: os.Getenv("EVENTBROKER"),
	})
	if err != nil {
		eh.Logger.Error("Could not initialize Keptn handler: " + err.Error())
		return err
	}

	source, _ := url.Parse("unleash-service")
	contentType := "application/json"

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        eventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": keptnHandler.KeptnContext, "triggeredid": eh.Event.ID()},
		}.AsV02(),
		Data: data,
	}

	err = keptnHandler.SendCloudEvent(event)
	if err != nil {
		eh.Logger.Error("Could not send " + eventType + " event: " + err.Error())
		return err
	}
	return nil
}

// ToggleFeature sets a value for a feature flag
func toggleFeature(togglename string, togglevalue string) error {

	if os.Getenv("UNLEASH_SERVER_URL") == "" || os.Getenv("UNLEASH_USER") == "" || os.Getenv("UNLEASH_TOKEN") == "" {
		return errors.New("Unleash secret not available. Can not execute remediation action")
	}

	unleashAPIUrl := os.Getenv("UNLEASH_SERVER_URL")
	unleashUser := os.Getenv("UNLEASH_USER")
	unleashToken := os.Getenv("UNLEASH_TOKEN")
	unleashAPIUrlExt := "/admin/features/" + togglename + "/toggle/" + togglevalue

	client := &http.Client{}
	req, err := http.NewRequest("POST", unleashAPIUrl+unleashAPIUrlExt, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(unleashUser, unleashToken)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	fmt.Println("unleash status code: " + strconv.Itoa(resp.StatusCode))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return errors.New("could not update feature toggle")
	}

	return nil
}
