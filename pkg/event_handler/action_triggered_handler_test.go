package event_handler

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	keptnapi "github.com/keptn/go-utils/pkg/api/models"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/keptn/keptn/go-sdk/pkg/sdk"
	"github.com/stretchr/testify/require"
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

func newActionTriggeredEvent(filename string) cloudevents.Event {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	event := keptnapi.KeptnContextExtendedCE{}
	err = json.Unmarshal(content, &event)
	_ = err
	return keptnv2.ToCloudEvent(event)
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
	os.Setenv("EVENTBROKER", ts.URL)

	fakeKeptn := sdk.NewFakeKeptn("test-unleash-svc")
	fakeKeptn.AddTaskHandler("sh.keptn.event.action.triggered", NewActionTriggeredHandler())
	fakeKeptn.Start()

	fakeKeptn.NewEvent(newActionTriggeredEvent("test/events/action_triggered.json"))

	require.Equal(t, 2, len(fakeKeptn.GetEventSender().SentEvents))

	require.Equal(t, keptnv2.GetStartedEventType("action"), fakeKeptn.GetEventSender().SentEvents[0].Type())
	require.Equal(t, keptnv2.GetFinishedEventType("action"), fakeKeptn.GetEventSender().SentEvents[1].Type())

	finishedEvent, _ := keptnv2.ToKeptnEvent(fakeKeptn.GetEventSender().SentEvents[1])
	actionFinishedData := keptnv2.ActionFinishedEventData{}
	finishedEvent.DataAs(&actionFinishedData)
	log.Println("-----------------------------------------------------------")
	require.Equal(t, keptnv2.StatusSucceeded, actionFinishedData.Status)
	require.Equal(t, keptnv2.ResultPass, actionFinishedData.Result)
}
