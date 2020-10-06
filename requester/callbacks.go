package requester

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Stats ...
type Stats struct {
	StatusCode int
	Error      error
}

type WorkspaceData struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status" description:"The workspace status"`
}

// CheckInactive Logic to check inativity for the function
func CheckInactive(resp *http.Response, b *Work) Stats {
	data, _ := ioutil.ReadAll(resp.Body)
	var response WorkspaceData
	if err := json.Unmarshal(data, &response); err != nil {
		return Stats{
			StatusCode: resp.StatusCode,
			Error:      nil,
		}
	}

	// Create a request which will do some work
	for {
		status, statusCode, err := checkWorkspaceStatus(response.ID, b)
		if err != nil {
			return Stats{
				StatusCode: statusCode,
				Error:      err,
			}
		}
		if status == "INACTIVE" {
			return Stats{
				StatusCode: resp.StatusCode,
				Error:      nil,
			}
		}
		if status == "TEMPLATE_ERROR" {
			return Stats{
				StatusCode: 500,
				Error:      fmt.Errorf("TEMPLATE_ERROR"),
			}
		}
		time.Sleep(time.Duration(b.CallBackConfig.BackOffTime) * time.Second)
	}
}

// checkWorkspaceStatus ...
func checkWorkspaceStatus(id string, b *Work) (string, int, error) {
	baseUrl := b.ReturnBaseURL() + "/v1/workspaces/" + id
	client := http.Client{}
	req, _ := http.NewRequest("GET", baseUrl, nil)
	req.Header = b.Request.Header
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("Got unexpected status code")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	var response WorkspaceData
	json.Unmarshal(body, &response)
	return response.Status, resp.StatusCode, nil
}
