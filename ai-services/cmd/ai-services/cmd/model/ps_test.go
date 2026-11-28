package model

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/models"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/types"
)

func TestBuildModelPSRows_LocalWithPods(t *testing.T) {
	ps := &models.ModelPSResponse{
		Name:        "granite-llm",
		Status:      "Running",
		WorkerName:  "lpar-1",
		RuntimeType: "podman",
		Pods: []types.Pod{
			{PodID: "abcdef1234567890", PodName: "vllm-pod", Status: types.Running, Healthy: true},
			{PodID: "1234567890abcdef", PodName: "vllm-pod-2", Status: types.Exited},
		},
	}

	rows := buildModelPSRows(ps, false)

	assert.Equal(t, [][]string{
		{"granite-llm", "lpar-1", "podman", "", "vllm-pod", "running (healthy)"},
		{"granite-llm", "lpar-1", "podman", "", "vllm-pod-2", "exited"},
	}, rows)
}

func TestBuildModelPSRows_LocalWide(t *testing.T) {
	ps := &models.ModelPSResponse{
		Name:        "granite-llm",
		WorkerName:  "lpar-1",
		RuntimeType: "podman",
		Pods:        []types.Pod{{PodID: "abcdef1234567890", PodName: "vllm-pod", Status: types.Exited}},
	}

	rows := buildModelPSRows(ps, true)

	assert.Len(t, rows, 1)
	assert.Len(t, rows[0], 9)
	assert.Equal(t, "abcdef123456", rows[0][4])
}

func TestBuildModelPSRows_RemoteNoPods(t *testing.T) {
	ps := &models.ModelPSResponse{
		Name:           "prod-watsonx",
		DeploymentType: "remote",
		Status:         "connected",
		Pods:           []types.Pod{},
	}

	assert.Equal(t,
		[][]string{{"prod-watsonx", "-", "-", "-", "-", "connected"}},
		buildModelPSRows(ps, false),
	)

	wide := buildModelPSRows(ps, true)
	assert.Len(t, wide, 1)
	assert.Len(t, wide[0], 9)
	assert.Equal(t, "connected", wide[0][6])
}

func TestPsOptions_Wide(t *testing.T) {
	assert.True(t, psOptions{output: "WIDE"}.wide())
	assert.False(t, psOptions{}.wide())
}
