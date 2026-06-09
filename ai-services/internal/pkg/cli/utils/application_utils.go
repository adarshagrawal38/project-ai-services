package utils

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	catalogClient "github.com/project-ai-services/ai-services/internal/pkg/catalog/client"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/types"
)

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

func GetAllApps(appClient *catalogClient.ApplicationClient) ([]types.Application, error) {
	// List all applications
	listResponse, err := appClient.ListApplications(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}

	return listResponse.Data, nil
}

func GetAppByName(appClient *catalogClient.ApplicationClient, appName string) (*types.Application, error) {
	listResponse, err := appClient.ListApplications(nil)
	if err != nil {
		return nil, err
	}
	for _, app := range listResponse.Data {
		if app.Name == appName {
			return &app, nil
		}
	}

	return nil, fmt.Errorf("application with name '%s' not found", appName)
}

// ParseErrorResponse attempts to parse the error response from the API.
// It returns the error message if successfully parsed, otherwise returns the raw response body.
func ParseErrorResponse(resp *resty.Response) string {
	var errResp ErrorResponse
	if err := json.Unmarshal(resp.Body(), &errResp); err == nil && errResp.Error != "" {
		return errResp.Error
	}

	return resp.String()
}
