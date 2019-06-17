package dtutils

import (
	"bytes"
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

// Event describes Dynatrace problem events
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

// CustomProperties describes the custom properties attached to a Dynatrace event
type CustomProperties struct {
	RemediationProvider string
	RemediationAction   string
	RemediationURL      string
	Approver            string
}

// Comment describes a comment (to be) attached to a Dynatrace problem
type Comment struct {
	Comment string `json:"comment"`
	User    string `json:"user"`
	Context string `json:"context"`
}

// ClusterInternal specifies if script is run inside the cluster or not
var ClusterInternal = true

// GetEventsFromEntity gets events from a Dynatrace entity
func GetEventsFromEntity(shkeptncontext, entityID string, startTime int) Events {
	dthost, dtapitoken, err := getDynatraceCredentials()

	if err != nil {
		keptnutils.Error(shkeptncontext, "Error when getting Dynatrace credentials: "+err.Error())
	}

	url := "https://" + dthost + "/api/v1/events?from=" + strconv.Itoa(startTime) + "&entityId=" + entityID + "&Api-Token=" + dtapitoken
	resp, err := http.Get(url)
	if err != nil {
		keptnutils.Error(shkeptncontext, "Error when getting Dynatrace events for entity: "+err.Error())
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

// PostComment posts a comment to a Dynatrace problem
func PostComment(shkeptncontext string, problemID string, commentText string) {
	dthost, dtapitoken, err := getDynatraceCredentials()
	fmt.Println("host, token: ", dthost, dtapitoken)
	if err != nil {
		keptnutils.Error(shkeptncontext, err.Error())
	}
	comment := Comment{commentText, "keptn@keptn.sh", "keptn"}
	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(comment)
	url := "https://" + dthost + "/api/v1/problem/details/" + problemID + "/comments?Api-Token=" + dtapitoken
	fmt.Println(url)
	res, err := http.Post(url, "application/json", payload)

	if err != nil {
		keptnutils.Error(shkeptncontext, "Error when posting comment to Dynatrace: "+err.Error())
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		keptnutils.Error(shkeptncontext, "Reponse error when posting comment to Dynatrace: "+err.Error())
		return
	}
	keptnutils.Debug(shkeptncontext, "Response from posting a comment: "+string(body))
}

func getDynatraceCredentials() (string, string, error) {

	api, err := keptnutils.GetKubeAPI(ClusterInternal)
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
