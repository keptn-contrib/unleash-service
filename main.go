package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	"github.com/keptn/unleash-service/dtutils"
	"github.com/keptn/unleash-service/unleashutils"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/client"
	cloudeventshttp "github.com/cloudevents/sdk-go/pkg/cloudevents/transport/http"

	"github.com/kelseyhightower/envconfig"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// ProblemEvent describes a problem event received via a Dynatrace Problem notification
type ProblemEvent struct {
	State            string           `json:"state"`
	ProblemID        string           `json:"problemID"`
	PID              string           `json:"pid"`
	ProblemTitle     string           `json:"problemTitle"`
	ProblemDetails   ProblemDetail    `json:"problemDetails"`
	ImpactedEntities []ImpactedEntity `json:"impactedEntities"`
	ImpactedEntity   string           `json:"impactedEntity"`
}

// ProblemDetail descibes the details of a Dynatrace problem
type ProblemDetail struct {
	ID            string `json:"id"`
	StartTime     int    `json:"startTime"`
	EndTime       int    `json:"endTime"`
	DisplayName   string `json:"displayName"`
	ImpactLevel   string `json:"impactLevel"`
	Status        string `json:"status"`
	SeverityLevel string `json:"severityLevel"`
	CommentCount  int    `json:"commentCount"`
	//TagsOfAffectedEntitites
	RankedEvents []dtutils.Event `json:"rankedEvents"`
	HasRootCause bool
}

// ImpactedEntity describes the impacted entity of a Dynatrace problem
type ImpactedEntity struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Entity string `json:"entity"`
}

const remediationUser = "keptn@keptn.sh"

func gotEvent(ctx context.Context, event cloudevents.Event) error {
	var shkeptncontext string
	event.Context.ExtensionAs("shkeptncontext", &shkeptncontext)

	keptnutils.Debug(shkeptncontext, fmt.Sprintf("Got Event Context: %+v", event.Context))
	//keptnutils.Debug(shkeptncontext, fmt.Sprintf("Source of Event:"))

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

	if data.State != "OPEN" {
		keptnutils.Error(shkeptncontext, "Problem with ProblemID "+data.PID+" is not OPEN. Aborting processing.")
		return nil
	}

	// post comment to Dynatrace
	dtutils.PostComment(shkeptncontext, data.PID, "starting unleash service, Problem state is: "+data.State)

	fmt.Println(data)

	// assume there is no root cause yet
	var generalRootCauseFound = false

	fmt.Println("has root cause: " + strconv.FormatBool(data.ProblemDetails.HasRootCause))

	// parse events from dynatrace
	for _, rankedEvent := range data.ProblemDetails.RankedEvents {
		if rankedEvent.IsRootCause {
			generalRootCauseFound = true
			keptnutils.Debug(shkeptncontext, "Root cause entity ID: "+rankedEvent.EntityID)
			events := dtutils.GetEventsFromEntity(shkeptncontext, rankedEvent.EntityID, rankedEvent.StartTime)

			for _, event := range events.Events {
				if event.EventType == "CUSTOM_CONFIGURATION" {
					if event.CustomProperties.RemediationProvider == "unleash" {
						myEvent = event
						keptnutils.Debug(shkeptncontext, "CUSTOM_CONFIGURATION with remediation provider 'unleash' event found: "+strconv.Itoa(event.EventID))
						//fmt.Println("EntityName: " + event.EntityName)

						//fmt.Println("RemediationProvider with root cause: " + event.CustomProperties.RemediationProvider)
					}
				}
			}

		}
	}
	if !generalRootCauseFound {
		keptnutils.Error(shkeptncontext, "ProblemDetails have no root cause attached.")
	}

	if myEvent.EventType == "CUSTOM_CONFIGURATON" {
		// TODO
	}

	fmt.Println("RemediationProvider from myEvent: " + myEvent.CustomProperties.RemediationProvider)

	// // check for impacted entities
	// for _, v := range data.ImpactedEntities {
	// 	keptnutils.Info(shkeptncontext, "impacted:"+v.Name)
	// }

	//dtEvents := dtutils.GetEventsFromEntity(entityID)

	unleashServerURL := os.Getenv("UNLEASH_SERVER_URL")
	if unleashServerURL == "" {
		unleashServerURL = "http://unleash-server-service.default"
	}
	keptnutils.Debug(shkeptncontext, "Using Unleash server located at: unleash-server-service.default")

	client, body := unleashutils.LoginToServer(shkeptncontext, unleashServerURL, "keptn@keptn.sh")

	keptnutils.Info(shkeptncontext, string(body))

	// toggle feature flag to ON
	featureToggleName := myEvent.CustomProperties.FeatureToggle
	keptnutils.Debug(shkeptncontext, "Toggling feature flag "+featureToggleName+" to ON")

	unleashutils.SetFeatureFlag(shkeptncontext, client, unleashServerURL, featureToggleName, true, nil)

	keptnutils.Info(shkeptncontext, string(body))

	// post comment to Dynatrace
	dtutils.PostComment(shkeptncontext, data.PID, "finished unleash service, Problem state is: "+data.State)

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
