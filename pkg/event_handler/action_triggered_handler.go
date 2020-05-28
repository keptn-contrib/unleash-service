package event_handler

import (
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/google/uuid"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
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

	for feature, value := range actionTriggeredEvent.Action.Values {
		err = toggleFeature(feature, value)
		if err != nil {
			eh.Logger.Error("Could not set feature " + feature + " to value " + value + ": " + err.Error())
			err = eh.sendFinishedEvent(*actionTriggeredEvent, keptn.ActionResultPass, keptn.ActionStatusErrored)
			if err != nil {
				eh.Logger.Error("could not send action-finished event: " + err.Error())
				return err
			}
		}
	}

	err = eh.sendFinishedEvent(*actionTriggeredEvent, keptn.ActionResultPass, keptn.ActionStatusSucceeded)
	if err != nil {
		eh.Logger.Error("could not send action-finished event: " + err.Error())
		return err
	}
	return nil
}

func (eh ActionTriggeredHandler) sendFinishedEvent(actionTriggeredEvent keptn.ActionTriggeredEventData, result keptn.ActionResultType, status keptn.ActionStatusType) error {
	keptnHandler, err := keptn.NewKeptn(&eh.Event, keptn.KeptnOpts{
		EventBrokerURL: os.Getenv("EVENTBROKER"),
	})
	if err != nil {
		eh.Logger.Error("Could not initialize Keptn handler: " + err.Error())
		return err
	}

	source, _ := url.Parse("unleash-service")
	contentType := "application/json"

	actionFinishedEvent := &keptn.ActionFinishedEventData{
		Project: actionTriggeredEvent.Project,
		Service: actionTriggeredEvent.Stage,
		Stage:   actionTriggeredEvent.Stage,
		Action: keptn.ActionResult{
			Result: result,
			Status: status,
		},
		Problem: actionTriggeredEvent.Problem,
		Values:  actionTriggeredEvent.Action.Values,
		Labels:  actionTriggeredEvent.Labels,
	}

	event := cloudevents.Event{
		Context: cloudevents.EventContextV02{
			ID:          uuid.New().String(),
			Time:        &types.Timestamp{Time: time.Now()},
			Type:        keptn.ActionFinishedEventType,
			Source:      types.URLRef{URL: *source},
			ContentType: &contentType,
			Extensions:  map[string]interface{}{"shkeptncontext": keptnHandler.KeptnContext},
		}.AsV02(),
		Data: actionFinishedEvent,
	}

	err = keptnHandler.SendCloudEvent(event)
	if err != nil {
		eh.Logger.Error("Could not send action.finished event: " + err.Error())
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
