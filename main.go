package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	"github.com/keptn/unleash-service/dtutils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	// Port on which to listen for cloudevents
	Port int    `envconfig:"RCV_PORT" default:"8080"`
	Path string `envconfig:"RCV_PATH" default:"/"`
}

func main() {
	var env envConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Fatalf("Failed to process env var: %s", err)
	}
	os.Exit(_main(os.Args[1:], env))
}

// ProblemEvent ...
type ProblemEvent struct {
	State            string           `json:"state"`
	ProblemID        string           `json:"problemID"`
	PID              string           `json:"pid"`
	ProblemTitle     string           `json:"problemTitle"`
	ProblemDetails   ProblemDetail    `json:"problemDetails"`
	ImpactedEntities []ImpactedEntity `json:"impactedEntities"`
	ImpactedEntity   string           `json:"impactedEntity"`
}

// ProblemDetail ...
type ProblemDetail struct {
	ID            string
	StartTime     int
	EndTime       int
	DisplayName   string
	ImpactLevel   string
	Status        string
	SeverityLevel string
	CommentCount  int
	//TagsOfAffectedEntitites
	RankedEvents []dtutils.Event
}

// ImpactedEntity ...
type ImpactedEntity struct {
	Type   string
	Name   string
	Entity string
}

// CustomProperties ...
type CustomProperties struct {
	RemediationProvider string
	RemediationAction   string
	RemediationURL      string
	Approver            string
}

const remediationUser = "keptn@keptn.sh"

func gotEvent(ctx context.Context, event cloudevents.Event) error {
	var shkeptncontext string
	event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	keptnutils.Debug(shkeptncontext, fmt.Sprintf("Got Event Context: %+v", event.Context))

	data := &ProblemEvent{}
	if err := event.DataAs(data); err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Got Data Error: %s", err.Error()))
		return err
	}

	// abort if no sh.keptn.events.problem event
	if event.Type() != "sh.keptn.events.problem" {
		const errorMsg = "Received unexpected keptn event"
		keptnutils.Error(shkeptncontext, errorMsg)
		return errors.New(errorMsg)
	}
	var myEvent = dtutils.Event{}

	// check for ranked events
	var generalRootCauseFound = false
	for _, rankedEvent := range data.ProblemDetails.RankedEvents {
		if rankedEvent.IsRootCause {
			generalRootCauseFound = true
			keptnutils.Debug(shkeptncontext, "Root cause entity ID: "+rankedEvent.EntityID)
			events := dtutils.GetEventsFromEntity(shkeptncontext, rankedEvent.EntityID, rankedEvent.StartTime)
			var rootCauseFound = false

			for _, event := range events.Events {
				if event.EventType == "CUSTOM_CONFIGURATION" && event.IsRootCause {
					rootCauseFound = true
					myEvent = event
					keptnutils.Info(shkeptncontext, "Root cause of problem: "+strconv.Itoa(event.EventID))
					fmt.Println("EntityName: " + event.EntityName)

					fmt.Println("RemediationProvider: " + event.CustomProperties.RemediationProvider)
				} else {

					fmt.Println("RemediationProvider: " + event.CustomProperties.RemediationProvider)
				}

			}
			if !rootCauseFound {
				keptnutils.Info(shkeptncontext, "No root cause with CUSTOM_CONFIGURATION found")
				//return nil // errors.New("No root cause with CUSTOM_CONFIGURATION found")
			}
		}
	}
	if !generalRootCauseFound {
		keptnutils.Error(shkeptncontext, "ProblemDetails have no root cause attached.")
	}

	if myEvent.EventType == "CUSTOM_CONFIGURATON" {
		// TODO
	}

	fmt.Println("get remediation provider")

	fmt.Println("RemediationPRovider: " + myEvent.CustomProperties.RemediationProvider)

	// // check for impacted entities
	// for _, v := range data.ImpactedEntities {
	// 	keptnutils.Info(shkeptncontext, "impacted:"+v.Name)
	// }

	//dtEvents := dtutils.GetEventsFromEntity(entityID)

	requestBody, err := json.Marshal(map[string]string{
		"email": remediationUser,
	})
	if err != nil {
		keptnutils.Error(shkeptncontext, err.Error())
	}

	unleashServerURL := os.Getenv("UNLEASH_SERVER_URL")
	if unleashServerURL == "" {
		unleashServerURL = "http://unleash-server-service.default"
	}

	// setup cookie jar to store login information
	cookiejar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookiejar,
	}

	// login to Unleash server
	resp, err := client.Post(unleashServerURL+"/api/admin/login", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when logging in to Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	keptnutils.Info(shkeptncontext, string(body))

	// toggle feature flag to ON
	keptnutils.Debug(shkeptncontext, "Toggling feature flag NAMENAMENAME to ON")
	resp, err = client.Post(unleashServerURL+"/api/admin/features/ServeStaticReviews/toggle/on", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when toggling feature in to Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)

	keptnutils.Info(shkeptncontext, string(body))

	return nil //sendDeploymentFinishedEvent(shkeptncontext, event)
}

// func sendDeploymentFinishedEvent(shkeptncontext string, incomingEvent cloudevents.Event) error {

// 	source, _ := url.Parse("deploy-service")
// 	contentType := "application/json"

// 	event := cloudevents.Event{
// 		Context: cloudevents.EventContextV02{
// 			ID:          uuid.New().String(),
// 			Type:        "sh.keptn.events.deployment-finished",
// 			Source:      types.URLRef{URL: *source},
// 			ContentType: &contentType,
// 			Extensions:  map[string]interface{}{"shkeptncontext": shkeptncontext},
// 		}.AsV02(),
// 		Data: incomingEvent.Data,
// 	}

// 	t, err := cloudeventshttp.New(
// 		cloudeventshttp.WithTarget("http://event-broker.keptn.svc.cluster.local/keptn"),
// 		cloudeventshttp.WithEncoding(cloudeventshttp.StructuredV02),
// 	)
// 	if err != nil {
// 		return errors.New("Failed to create transport:" + err.Error())
// 	}

// 	c, err := client.New(t)
// 	if err != nil {
// 		return errors.New("Failed to create HTTP client:" + err.Error())
// 	}

// 	if _, err := c.Send(context.Background(), event); err != nil {
// 		return errors.New("Failed to send cloudevent:, " + err.Error())
// 	}
// 	return nil
// }

func _main(args []string, env envConfig) int {
	keptnutils.ServiceName = "unleash-service"

	ctx := context.Background()

	t, err := cloudeventshttp.New(
		cloudeventshttp.WithPort(env.Port),
		cloudeventshttp.WithPath(env.Path),
	)

	if err != nil {
		log.Fatalf("failed to create transport, %v", err)
	}
	c, err := client.New(t)
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Printf("will listen on :%d%s\n", env.Port, env.Path)
	log.Fatalf("failed to start receiver: %s", c.StartReceiver(ctx, gotEvent))

	return 0
}
