package event_handler

import (
	"encoding/json"
	keptnapi "github.com/keptn/go-utils/pkg/api/models"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/go-utils/pkg/sdk"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
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

func newActionTriggeredEvent(filename string) keptnapi.KeptnContextExtendedCE {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	event := keptnapi.KeptnContextExtendedCE{}
	err = json.Unmarshal(content, &event)
	_ = err
	return event
}

func Test_Receiving_GetActionTriggeredEvent(t *testing.T) {
	ch := make(chan *keptnapi.KeptnContextExtendedCE)

	var returnedStatusCode = 200
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

	fakeKeptn := sdk.NewFakeKeptn("test-unleash-svc")
	fakeKeptn.AddTaskHandler("sh.keptn.event.action.triggered", NewActionTriggeredHandler())

	fakeKeptn.NewEvent(newActionTriggeredEvent("test/events/action_triggered.json"))

	fakeKeptn.AssertNumberOfEventSent(t, 2)

	fakeKeptn.AssertSentEventType(t, 0, keptnv2.GetStartedEventType("action"))
	fakeKeptn.AssertSentEventType(t, 1, keptnv2.GetFinishedEventType("action"))

	fakeKeptn.AssertSentEventStatus(t, 1, keptnv2.StatusSucceeded)
	fakeKeptn.AssertSentEventResult(t, 1, keptnv2.ResultPass)
}
