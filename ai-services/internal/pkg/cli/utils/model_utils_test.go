package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/models"
	catalogClient "github.com/project-ai-services/ai-services/internal/pkg/catalog/client"
	catalogConfig "github.com/project-ai-services/ai-services/internal/pkg/catalog/config"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/types"
)

// newModelClientFor returns a ModelClient talking to a fake catalog server that serves the given
// pages of models. Credentials are written to a temp config dir so NewModelClient needs no refresh.
func newModelClientFor(t *testing.T, pages [][]models.ModelListItem) *catalogClient.ModelClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.ModelListResponse{
			Data:       pages[page-1],
			Pagination: types.PaginationMetadata{Page: page, HasNext: page < len(pages)},
		})
	}))
	t.Cleanup(srv.Close)

	// os.UserConfigDir() honours HOME on macOS and XDG_CONFIG_HOME on Linux.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)

	require.NoError(t, catalogConfig.Save(catalogConfig.Credentials{
		ServerURL:         srv.URL,
		AccessToken:       "token",
		RefreshToken:      "refresh",
		AccessTokenExpiry: time.Now().Add(time.Hour),
	}))

	mc, err := catalogClient.NewModelClient(context.Background())
	require.NoError(t, err)

	return mc
}

func item(id, name string) models.ModelListItem {
	return models.ModelListItem{ModelBase: models.ModelBase{ID: id, Name: name}}
}

func TestGetModelByName(t *testing.T) {
	tests := []struct {
		name    string
		pages   [][]models.ModelListItem
		lookup  string
		wantID  string
		wantErr bool
	}{
		{
			name:   "exact match among substring matches",
			pages:  [][]models.ModelListItem{{item("1", "granite-llm-large"), item("2", "granite-llm")}},
			lookup: "granite-llm",
			wantID: "2",
		},
		{
			name:   "case-insensitive",
			pages:  [][]models.ModelListItem{{item("1", "Granite-LLM")}},
			lookup: "granite-llm",
			wantID: "1",
		},
		{
			name:   "match on a later page",
			pages:  [][]models.ModelListItem{{item("1", "granite-a")}, {item("2", "granite")}},
			lookup: "granite",
			wantID: "2",
		},
		{
			name:    "no match",
			pages:   [][]models.ModelListItem{{item("1", "granite-llm-large")}},
			lookup:  "granite-llm",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetModelByName(context.Background(), newModelClientFor(t, tt.pages), tt.lookup)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")

				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}
