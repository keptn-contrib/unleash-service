package dtutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetEventsFromEntity ...
func GetEventsFromEntity(shkeptncontext, entityID string, startTime int) Events {
	dthost, dtapitoken, err := getDynatraceCredentials()
	if err != nil {
		keptnutils.Error(shkeptncontext, err.Error())
	}
	fmt.Println("https://" + dthost + "/api/v1/events?from=" + strconv.Itoa(startTime) + "&entityId=" + entityID + "&Api-Token=" + dtapitoken)
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

func getDynatraceCredentials() (string, string, error) {

	api, err := keptnutils.GetKubeAPI(false)
	if err != nil {
		return "", "", err
	}

	getOptions := metav1.GetOptions{}
	secret, err := api.Secrets("keptn").Get("dynatrace", getOptions)
	if err != nil {
		return "", "", err
	}

	return string(secret.Data["DT_TENANT"]), string(secret.Data["DT_API_TOKEN"]), nil
}
