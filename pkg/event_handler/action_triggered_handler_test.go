package event_handler

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/types"
	"github.com/go-openapi/strfmt"
	keptnapi "github.com/keptn/go-utils/pkg/api/models"
	keptn "github.com/keptn/go-utils/pkg/lib"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_toggleFeature(t *testing.T) {

	var returnedStatusCode int
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(returnedStatusCode)
			w.Write([]byte(`{}`))
		}),
	)
	defer ts.Close()

	type args struct {
		togglename  string
		togglevalue string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		returnStatus    int
		unleashURL      string
		unleashUser     string
		unleashPassword string
	}{
		{
			name:            "Succeed",
			args:            args{},
			wantErr:         false,
			returnStatus:    200,
			unleashURL:      "unleash",
			unleashUser:     "keptn",
			unleashPassword: "keptn",
		},
		{
			name:            "Empty credentials - fail",
			args:            args{},
			wantErr:         true,
			returnStatus:    200,
			unleashURL:      "unleash",
			unleashUser:     "",
			unleashPassword: "",
		},
		{
			name:            "Receive error code - fail",
			args:            args{},
			wantErr:         true,
			returnStatus:    400,
			unleashURL:      "unleash",
			unleashUser:     "keptn",
			unleashPassword: "keptn",
		},
	}
	for _, tt := range tests {
		os.Setenv("UNLEASH_SERVER_URL", ts.URL)
		os.Setenv("UNLEASH_USER", tt.unleashUser)
		os.Setenv("UNLEASH_TOKEN", tt.unleashPassword)

		returnedStatusCode = tt.returnStatus

		t.Run(tt.name, func(t *testing.T) {
			if err := toggleFeature(tt.args.togglename, tt.args.togglevalue); (err != nil) != tt.wantErr {
				t.Errorf("toggleFeature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActionTriggeredHandler_HandleEvent(t *testing.T) {

	ch := make(chan *keptnapi.KeptnContextExtendedCE)

	var returnedStatusCode int
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			if strings.Contains(r.URL.String(), "/admin/features/") {
				w.WriteHeader(returnedStatusCode)
				w.Write([]byte(`{}`))
				return
			}

			defer r.Body.Close()
			body, _ := ioutil.ReadAll(r.Body)
			keptnCE := &keptnapi.KeptnContextExtendedCE{}

			_ = json.Unmarshal(body, keptnCE)

			w.WriteHeader(returnedStatusCode)
			w.Write([]byte(`{}`))
			go func() { ch <- keptnCE }()
		}),
	)
	defer ts.Close()

	os.Setenv("UNLEASH_SERVER_URL", ts.URL)
	os.Setenv("UNLEASH_USER", "keptn")
	os.Setenv("UNLEASH_TOKEN", "keptn")
	os.Setenv("EVENTBROKER", ts.URL)

	type fields struct {
		Logger keptn.LoggerInterface
		Event  cloudevents.Event
	}
	tests := []struct {
		name         string
		fields       fields
		wantErr      bool
		wantEvent    *keptnapi.KeptnContextExtendedCE
		returnStatus int
	}{
		{
			name: "Succeed",
			fields: fields{
				Logger: keptn.NewLogger("", "", ""),
				Event: cloudevents.Event{
					Context: &cloudevents.EventContextV02{
						SpecVersion: "0.2",
						Type:        "sh.keptn.events.tests-finished",
						Source:      types.URLRef{},
						ID:          "",
						Time:        nil,
						SchemaURL:   nil,
						ContentType: stringp("application/json"),
						Extensions:  nil,
					},
					Data: []byte(`{
    "action": {
      "name": "FeatureToggle",
      "action": "toggle-feature",
      "description": "toggle a feature",
      "value": {
        "EnableItemCache": "on"
      }
    },
    "problem": {
      "ImpactedEntity": "carts-primary",
      "PID": "93a5-3fas-a09d-8ckf",
      "ProblemDetails": "Pod name",
      "ProblemID": "762",
      "ProblemTitle": "cpu_usage_sockshop_carts",
      "State": "OPEN"
    },
    "project": "sockshop",
    "stage": "staging",
    "service": "carts",
    "labels": {
      "testid": "12345",
      "buildnr": "build17",
      "runby": "JohnDoe"
    }
  }`),
					DataEncoded: false,
				},
			},
			wantErr: false,
			wantEvent: &keptnapi.KeptnContextExtendedCE{
				Contenttype: "application/json",
				Data: []byte(`{
    "action": {
      "result": "pass",
      "status": "succeeded",
    },
    "problem": {
      "ImpactedEntity": "carts-primary",
      "PID": "93a5-3fas-a09d-8ckf",
      "ProblemDetails": "Pod name",
      "ProblemID": "762",
      "ProblemTitle": "cpu_usage_sockshop_carts",
      "State": "OPEN"
    },
    "project": "sockshop",
    "stage": "staging",
    "service": "carts",
    "labels": {
      "testid": "12345",
      "buildnr": "build17",
      "runby": "JohnDoe"
    }
  }`),
				Extensions:     nil,
				ID:             "",
				Shkeptncontext: "",
				Source:         nil,
				Specversion:    "",
				Time:           strfmt.DateTime{},
				Type:           stringp(keptn.ActionFinishedEventType),
			},
			returnStatus: 200,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			returnedStatusCode = tt.returnStatus
			eh := ActionTriggeredHandler{
				Logger: tt.fields.Logger,
				Event:  tt.fields.Event,
			}
			if err := eh.HandleEvent(); (err != nil) != tt.wantErr {
				t.Errorf("HandleEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			select {
			case msg := <-ch:
				t.Logf("Received event on event broker: %v", msg)

				if *msg.Type != *tt.wantEvent.Type {
					t.Errorf("HandleEvent() sent event type = %v, wantEventType %v", *msg.Type, *tt.wantEvent.Type)
				}
			case <-time.After(5 * time.Second):
				t.Errorf("Message did not make it to the receiver")
			}
		})
	}
}

func stringp(s string) *string {
	return &s
}
