package event_handler

import (
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/keptn/go-utils/pkg/lib/keptn"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const ActionToggleFeature = "toggle-feature"

type ActionTriggeredHandler struct {
	Logger keptn.LoggerInterface
	Event  cloudevents.Event
}

func (eh ActionTriggeredHandler) HandleEvent() error {
	var shkeptncontext string
	_ = eh.Event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	actionTriggeredEvent := &keptnv2.ActionTriggeredEventData{}

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
	if sendErr := eh.sendEvent(keptnv2.GetStartedEventType(keptnv2.ActionTaskName), eh.getActionStartedEvent(*actionTriggeredEvent)); sendErr != nil {
		eh.Logger.Error(sendErr.Error())
		return errors.New(sendErr.Error())
	}

	values, ok := actionTriggeredEvent.Action.Value.(map[string]interface{})

	if !ok {
		msg := "Could not parse action.value"
		eh.Logger.Error(msg)
		err = eh.sendEvent(keptnv2.GetFinishedEventType(keptnv2.ActionTaskName),
			eh.getActionFinishedEvent(keptnv2.ResultFailed, keptnv2.StatusErrored, *actionTriggeredEvent, msg))
		return errors.New(msg)
	}

	for feature, value := range values {
		if _, ok := value.(string); !ok {
			msg := "Value property of feature toggle remediation action not valid. It must be set: TOGGLENAME:\"on\" or TOGGLENAME:\"off\""
			eh.Logger.Error(msg)
			sendErr := eh.sendEvent(keptnv2.GetFinishedEventType(keptnv2.ActionTaskName),
				eh.getActionFinishedEvent(keptnv2.ResultFailed, keptnv2.StatusErrored, *actionTriggeredEvent, msg))
			if sendErr != nil {
				eh.Logger.Error("could not send action-finished event: " + err.Error())
				return err
			}
			return errors.New("Value property of feature toggle remediation action not valid. It must be set: TOGGLENAME:\"on\" or TOGGLENAME:\"off\"")
		}
		err = toggleFeature(feature, value.(string))
		if err != nil {
			msg := "Could not set feature " + feature + " to value " + value.(string) + ": " + err.Error()
			eh.Logger.Error(msg)
			sendErr := eh.sendEvent(keptnv2.GetFinishedEventType(keptnv2.ActionTaskName),
				eh.getActionFinishedEvent(keptnv2.ResultFailed, keptnv2.StatusErrored, *actionTriggeredEvent, msg))
			if sendErr != nil {
				eh.Logger.Error("could not send action-finished event: " + err.Error())
				return err
			}
			return err
		}
	}

	err = eh.sendEvent(keptnv2.GetFinishedEventType(keptnv2.ActionTaskName),
		eh.getActionFinishedEvent(keptnv2.ResultPass, keptnv2.StatusSucceeded, *actionTriggeredEvent, ""))
	if err != nil {
		eh.Logger.Error("could not send action-finished event: " + err.Error())
		return err
	}
	return nil
}

func (eh ActionTriggeredHandler) getActionFinishedEvent(result keptnv2.ResultType, status keptnv2.StatusType, actionTriggeredEvent keptnv2.ActionTriggeredEventData, message string) keptnv2.ActionFinishedEventData {

	return keptnv2.ActionFinishedEventData{
		EventData: keptnv2.EventData{
			Project: actionTriggeredEvent.Project,
			Stage:   actionTriggeredEvent.Stage,
			Service: actionTriggeredEvent.Service,
			Labels:  actionTriggeredEvent.Labels,
			Status:  status,
			Result:  result,
			Message: message,
		},
		Action: keptnv2.ActionData{},
	}
}

func (eh ActionTriggeredHandler) getActionStartedEvent(actionTriggeredEvent keptnv2.ActionTriggeredEventData) keptnv2.ActionStartedEventData {

	return keptnv2.ActionStartedEventData{
		EventData: keptnv2.EventData{
			Project: actionTriggeredEvent.Project,
			Service: actionTriggeredEvent.Service,
			Stage:   actionTriggeredEvent.Stage,
			Labels:  actionTriggeredEvent.Labels,
		},
	}
}

func (eh ActionTriggeredHandler) sendEvent(eventType string, data interface{}) error {
	keptnHandler, err := keptnv2.NewKeptn(&eh.Event, keptn.KeptnOpts{
		EventBrokerURL: os.Getenv("EVENTBROKER"),
	})
	if err != nil {
		eh.Logger.Error("Could not initialize Keptn handler: " + err.Error())
		return err
	}

	source, _ := url.Parse("unleash-service")

	event := cloudevents.NewEvent()
	event.SetType(eventType)
	event.SetSource(source.String())
	event.SetDataContentType(cloudevents.ApplicationJSON)
	event.SetExtension("shkeptncontext", keptnHandler.KeptnContext)
	event.SetExtension("triggeredid", eh.Event.ID())
	event.SetData(cloudevents.ApplicationJSON, data)

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
