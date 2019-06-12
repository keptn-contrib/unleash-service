package unleashutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	keptnutils "github.com/keptn/go-utils/pkg/utils"
)

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

func ToggleFeatureFlag(shkeptncontext string, client *http.Client, unleashServerURL string, featureToggleName string, body []byte) {
	keptnutils.Debug(shkeptncontext, "Trying to set feature to "+featureToggleName)
	resp, err := client.Post(unleashServerURL+"/api/admin/features/"+featureToggleName+"/toggle/on", "application/json", bytes.NewBuffer(body))
	if err != nil {
		keptnutils.Error(shkeptncontext, fmt.Sprintf("Error when toggling feature in to Unleash server: %s", err.Error()))
	}
	defer resp.Body.Close()
	body, _ = ioutil.ReadAll(resp.Body)
}
