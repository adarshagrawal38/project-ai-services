package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/models"
	"github.com/project-ai-services/ai-services/internal/pkg/utils"
)

const (
	modelsRoute     = "/api/v1/models"
	getModelRoute   = "/api/v1/models/%s"
	getModelPSRoute = "/api/v1/models/%s/ps"
)

// ModelClient provides methods for interacting with the models API.
type ModelClient struct {
	client *Client
}

// NewModelClient creates a new ModelClient using the stored catalog credentials.
func NewModelClient(ctx context.Context) (*ModelClient, error) {
	client, err := New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}

	return &ModelClient{client: client}, nil
}

// listModelsQuery converts the list params into the query parameters of GET /api/v1/models.
// Empty values are omitted.
func listModelsQuery(params *ListModelsParams) map[string][]string {
	query := map[string][]string{}
	if params == nil {
		return query
	}

	if params.Page > 0 {
		query["page"] = []string{strconv.Itoa(params.Page)}
	}
	if params.PageSize > 0 {
		query["page_size"] = []string{strconv.Itoa(params.PageSize)}
	}

	for key, value := range map[string]string{
		"q":               params.Search,
		"component_type":  params.ComponentType,
		"deployment_type": params.DeploymentType,
		"status":          params.Status,
	} {
		if value != "" {
			query[key] = []string{value}
		}
	}

	if len(params.ProviderIDs) > 0 {
		query["provider_id"] = params.ProviderIDs
	}

	return query
}

// ListModels retrieves a page of the caller's models (local and remote).
func (c *ModelClient) ListModels(ctx context.Context, params *ListModelsParams) (*models.ModelListResponse, error) {
	var result models.ModelListResponse
	req := c.client.HTTPClient().R().
		SetContext(ctx).
		SetQueryParamsFromValues(listModelsQuery(params)).
		SetResult(&result)

	resp, err := req.Get(modelsRoute)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}

	if resp.IsError() {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    utils.ParseErrorResponse(resp),
		}
	}

	return &result, nil
}

// GetModel retrieves the full details of a model, local or remote, by ID.
func (c *ModelClient) GetModel(ctx context.Context, id string) (*models.ModelDetail, error) {
	var result models.ModelDetail
	resp, err := c.client.HTTPClient().R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf(getModelRoute, id))
	if err != nil {
		return nil, fmt.Errorf("get model: %w", err)
	}

	if resp.IsError() {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    utils.ParseErrorResponse(resp),
		}
	}

	return &result, nil
}

// GetModelPS retrieves the process and pod status of a model, local or remote, by ID.
func (c *ModelClient) GetModelPS(ctx context.Context, id string) (*models.ModelPSResponse, error) {
	var result models.ModelPSResponse
	resp, err := c.client.HTTPClient().R().
		SetContext(ctx).
		SetResult(&result).
		Get(fmt.Sprintf(getModelPSRoute, id))
	if err != nil {
		return nil, fmt.Errorf("get model ps: %w", err)
	}

	if resp.IsError() {
		return nil, &HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    utils.ParseErrorResponse(resp),
		}
	}

	return &result, nil
}
