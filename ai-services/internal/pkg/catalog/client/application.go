package client

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/httpclient"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/types"
)

// ApplicationClient provides methods for interacting with the applications API.
type ApplicationClient struct {
	serverURL  string
	token      string
	httpClient *httpclient.HTTPClient
}

// NewApplicationClient creates a new ApplicationClient with the given server URL and token.
func NewApplicationClient(serverURL, token string) *ApplicationClient {
	return &ApplicationClient{
		serverURL:  serverURL,
		token:      token,
		httpClient: httpclient.New(serverURL),
	}
}

// ListApplicationsParams holds optional query parameters for listing applications.
type ListApplicationsParams struct {
	// Page is the page number (1-indexed). Default: 1
	Page int
	// PageSize is the number of items per page (max: 100). Default: 20
	PageSize int
	// DeploymentType filters by deployment type: 'architectures' or 'services'
	DeploymentType string
	// CatalogID filters by catalog ID (e.g., 'rag', 'chat', 'digitize', 'summarize')
	CatalogID string
}

// ListApplications retrieves a paginated list of all applications for the authenticated user.
// It supports optional filters via the params argument.
//
// Example:
//
//	client := NewApplicationClient("https://localhost:8080", "your-token")
//	resp, err := client.ListApplications(&ListApplicationsParams{
//	    Page: 1,
//	    PageSize: 20,
//	    DeploymentType: "services",
//	    CatalogID: "rag",
//	})
func (c *ApplicationClient) ListApplications(params *ListApplicationsParams) (*types.ApplicationListResponse, error) {
	query := make(map[string]string)

	if params != nil {
		if params.Page > 0 {
			query["page"] = strconv.Itoa(params.Page)
		}
		if params.PageSize > 0 {
			query["page_size"] = strconv.Itoa(params.PageSize)
		}
		if params.DeploymentType != "" {
			query["deployment_type"] = params.DeploymentType
		}
		if params.CatalogID != "" {
			query["catalog_id"] = params.CatalogID
		}
	}

	var resp types.ApplicationListResponse
	err := c.httpClient.Do(httpclient.Request{
		Method:   http.MethodGet,
		Endpoint: "/api/v1/applications",
		Headers:  map[string]string{"Authorization": "Bearer " + c.token},
		Query:    query,
		Out:      &resp,
	})
	if err != nil {
		return nil, fmt.Errorf("list applications: %w", err)
	}

	return &resp, nil
}

// GetApplicationPS retrieves the process status and runtime information for an application.
// It returns details about pods, containers, and their health status.
func (c *ApplicationClient) GetApplicationPS(id string) (*types.ApplicationPSResponse, error) {
	var psResp types.ApplicationPSResponse
	err := c.httpClient.Do(httpclient.Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/api/v1/applications/%s/ps", id),
		Headers:  map[string]string{"Authorization": "Bearer " + c.token},
		Out:      &psResp,
	})
	if err != nil {
		return nil, fmt.Errorf("get application ps: %w", err)
	}

	return &psResp, nil
}

// Made with Bob
