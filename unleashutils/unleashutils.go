package unleashutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strconv"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

// LoginToServer logs in the user and returns a http.client with the login cookie
func LoginToServer(shkeptncontext string, unleashServerURL string, user string) (client *http.Client, body []byte) {
	// setup cookie jar to store login information
	cookiejar, _ := cookiejar.New(nil)
	client = &http.Client{
		Jar: cookiejar,
	}

	requestBody, err := json.Marshal(map[string]string{
		"email": user,
	})
	if err != nil {
		keptnutils.Error(shkeptncontext, err.Error())
	}

	// login to Unleash server
	resp, err := client.Post(unleashServerURL+"/api/admin/login", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when logging in to Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)

	return client, body

}

// ToggleFeatureFlag toggles a given feature flag in the unleash server
func ToggleFeatureFlag(shkeptncontext string, client *http.Client, unleashServerURL string, featureToggleName string, body []byte) {
	keptnutils.Debug(shkeptncontext, "Trying to toogle feature "+featureToggleName+" in Unleash server: "+unleashServerURL)
	resp, err := client.Post(unleashServerURL+"/api/admin/features/"+featureToggleName+"/toggle", "application/json", bytes.NewBuffer(body))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when toggling feature in Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
}

// SetFeatureFlag sets a given flag to a given state (on or off)
func SetFeatureFlag(shkeptncontext string, client *http.Client, unleashServerURL string, featureToggleName string, enabled bool, body []byte) {
	keptnutils.Debug(shkeptncontext, "Trying to set feature for "+featureToggleName+" to: "+strconv.FormatBool(enabled)+" in Unleash server: "+unleashServerURL)
	featureState := "off"
	if enabled {
		featureState = "on"
	}
	resp, err := client.Post(unleashServerURL+"/api/admin/features/"+featureToggleName+"/toggle/"+featureState, "application/json", bytes.NewBuffer(body))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when toggling feature in Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
}
