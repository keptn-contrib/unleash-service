package dtutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

// Events ...
type Events struct {
	TotalEventCount int
	Events          []Event
}

// Event ...
type Event struct {
	EventID          int
	StartTime        int
	EndTime          int
	EntityID         string
	EntityName       string
	SeverityLevel    string
	ImpactLevel      string
	EventType        string
	CustomProperties CustomProperties
	IsRootCause      bool
}

// CustomProperties ...
type CustomProperties struct {
	RemediationProvider string
	RemediationAction   string
	RemediationURL      string
	Approver            string
}

var dthost = ""
var dtapitoken = ""

// GetEventsFromEntity ...
func GetEventsFromEntity(shkeptncontext, entityID string, startTime int) Events {
	resp, err := http.Get("https://" + dthost + "/api/v1/events?from=" + strconv.Itoa(startTime) + "&entityId=" + entityID + "&Api-Token=" + dtapitoken)
	if err != nil {
		keptnutils.Error(shkeptncontext, err.Error())
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	keptnutils.Info(shkeptncontext, string(body))
	data := &Events{}

	if err := json.Unmarshal(body, data); err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Got Data Error: %s", err.Error()))
		//	log.Fatalln(err)
	}
	//fmt.Println("data.TotalEventCount: " + strconv.Itoa(data.TotalEventCount))
	return *data

}
