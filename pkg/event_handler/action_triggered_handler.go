package event_handler

import (
	"errors"
	"fmt"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/keptn/go-sdk/pkg/sdk"
	"net/http"
	"os"
	"strconv"
)

const ActionToggleFeature = "toggle-feature"

type ActionTriggeredHandler struct {
}

// NewActionTriggeredHandler creates a new ActionTriggeredHandler
func NewActionTriggeredHandler() *ActionTriggeredHandler {
	return &ActionTriggeredHandler{}
}

// Execute handles the incoming cloud events
func (g *ActionTriggeredHandler) Execute(k sdk.IKeptn, event sdk.KeptnEvent) (interface{}, *sdk.Error) {
	actionTriggeredEvent := &keptnv2.ActionTriggeredEventData{}

	if err := keptnv2.Decode(event.Data, actionTriggeredEvent); err != nil {
		return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: "feature toggle remediation action not well formed: " + err.Error()}
	}

	if actionTriggeredEvent.Action.Action != ActionToggleFeature {
		k.Logger().Info("Received unknown action: " + actionTriggeredEvent.Action.Action + ". Exiting")
		return nil, nil
	}

	values, ok := actionTriggeredEvent.Action.Value.(map[string]interface{})

	if !ok {
		msg := "could not parse action.value"
		k.Logger().Error(msg)

		return nil, &sdk.Error{Err: errors.New(msg), StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: msg}
	}

	for feature, value := range values {
		if _, ok := value.(string); !ok {
			msg := "value property of feature toggle remediation action not valid. It must be set: TOGGLENAME:\"on\" or TOGGLENAME:\"off\""
			k.Logger().Error(msg)

			return nil, &sdk.Error{Err: errors.New(msg), StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: msg}
		}
		err := toggleFeature(feature, value.(string))
		if err != nil {
			msg := "Could not set feature " + feature + " to value " + value.(string) + ": " + err.Error()
			k.Logger().Error(msg)
			return nil, &sdk.Error{Err: err, StatusType: keptnv2.StatusErrored, ResultType: keptnv2.ResultFailed, Message: msg}
		}
	}

	finishedEventData := g.getActionFinishedEvent(keptnv2.ResultPass, keptnv2.StatusSucceeded, *actionTriggeredEvent, "")

	return finishedEventData, nil
}

func (g *ActionTriggeredHandler) getActionFinishedEvent(result keptnv2.ResultType, status keptnv2.StatusType, actionTriggeredEvent keptnv2.ActionTriggeredEventData, message string) keptnv2.ActionFinishedEventData {

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
