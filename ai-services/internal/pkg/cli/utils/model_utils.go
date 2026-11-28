package utils

import (
	"context"
	"fmt"
	"strings"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/models"
	catalogClient "github.com/project-ai-services/ai-services/internal/pkg/catalog/client"
)

// modelListPageSize is the maximum page size accepted by GET /api/v1/models.
const modelListPageSize = 100

// FetchModels retrieves either all models matching the filters or a specific model by name.
// If name is empty, every page of results is fetched. Otherwise, the model with that exact name is returned.
func FetchModels(ctx context.Context, modelClient *catalogClient.ModelClient, name string, filters catalogClient.ListModelsParams) ([]models.ModelListItem, error) {
	if name != "" {
		item, err := GetModelByName(ctx, modelClient, name)
		if err != nil {
			return nil, err
		}

		return []models.ModelListItem{*item}, nil
	}

	var items []models.ModelListItem
	for page := 1; ; page++ {
		params := filters
		params.Page = page
		params.PageSize = modelListPageSize

		resp, err := modelClient.ListModels(ctx, &params)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch models: %w", err)
		}

		items = append(items, resp.Data...)

		if !resp.Pagination.HasNext {
			break
		}
	}

	return items, nil
}

// GetModelByName resolves a model name to its list entry. The server search is a substring
// match, so every page of results is scanned for a case-insensitive exact match on the name.
func GetModelByName(ctx context.Context, modelClient *catalogClient.ModelClient, name string) (*models.ModelListItem, error) {
	for page := 1; ; page++ {
		resp, err := modelClient.ListModels(ctx, &catalogClient.ListModelsParams{
			Page:     page,
			PageSize: modelListPageSize,
			Search:   name,
		})
		if err != nil {
			return nil, err
		}

		for i := range resp.Data {
			if strings.EqualFold(resp.Data[i].Name, name) {
				return &resp.Data[i], nil
			}
		}

		if !resp.Pagination.HasNext {
			break
		}
	}

	return nil, fmt.Errorf("model with name '%s' not found", name)
}
