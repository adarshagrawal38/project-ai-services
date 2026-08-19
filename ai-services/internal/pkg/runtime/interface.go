package runtime

import (
	"context"
	"io"

	"github.com/project-ai-services/ai-services/internal/pkg/models"
	"github.com/project-ai-services/ai-services/internal/pkg/runtime/types"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Runtime interface {
	// Image operations
	ListImages() ([]types.Image, error)
	PullImage(ctx context.Context, image string) error

	// Pod operations
	ListPods(filters map[string][]string) ([]types.Pod, error)
	CreatePod(ctx context.Context, body io.Reader, opts map[string]string) ([]types.Pod, error)
	DeletePod(id string, force *bool) error
	StopPod(id string) error
	StartPod(id string) error
	InspectPod(nameOrId string) (*types.Pod, error)
	PodExists(nameOrID string) (bool, error)
	PodLogs(nameOrID string) error
	GetPodResources(nameOrID string) (*types.PodResources, error)
	GetNamespace() (string, error)

	// Secret operations
	ListSecrets(filters map[string][]string) ([]string, error)
	DeleteSecret(name string) error
	SecretExists(nameOrID string) (bool, error)
	UpdateSecret(name, deploymentName string, data map[string][]byte) error

	// Volume operations
	DeleteVolume(name string) error
	VolumeExists(nameOrID string) (bool, error)

	// Container operations
	// ListContainers(filters map[string][]string) ([]types.Container, error)
	InspectContainer(nameOrId string) (*types.Container, error)
	ContainerExists(nameOrID string) (bool, error)
	ContainerLogs(containerNameOrID string) error
	ExecInContainerWithCmd(podName, containerName string, command []string) (string, error)

	// Network operations
	ListRoutes(labelSelector string) ([]types.Route, error)

	// ListCRD populates crd list
	// resources in the namespace that carry every label key in filters["label"].
	ListCRD(list *unstructured.UnstructuredList, filters map[string][]string) ([]types.CRDResource, error)

	// Namespace operations
	DeleteNamespace(name string) error

	// PVC operations
	DeletePVCs(appLabel string) error

	// System information
	GetSystemInfo() (*models.SystemInfo, error)

	// RunEphemeralContainer runs a one-shot container (image + cmd + mounts),
	// waits for it to exit, and returns its exit code.
	// For local Podman this runs via the Podman socket directly.
	// For a remote agent this is dispatched over gRPC so the container
	// runs on the worker LPAR, not the control plane.
	RunEphemeralContainer(image string, cmd []string, mounts []types.BindMount) (int32, error)

	// Proxy operations – Caddy management on the node.
	// RegisterProxyRoute registers a route with the local Caddy instance.
	RegisterProxyRoute(ctx context.Context, route types.ProxyRoute) error
	// UnregisterProxyRoute removes a route from the local Caddy instance.
	UnregisterProxyRoute(routeID string) error
	// GetProxyRoute retrieves a route by ID from the local Caddy instance.
	GetProxyRoute(routeID string) (*types.ProxyRoute, error)
	// ProxyHealthCheck verifies the local Caddy instance is reachable.
	ProxyHealthCheck() error

	// HTTPProxy tunnels an HTTP request through the gRPC stream to a worker
	// pod endpoint and returns the response.
	// method is the HTTP verb (GET, POST, …), targetURL is the full URL of the
	// pod endpoint on the worker (e.g. "http://pod-name:8080/health"),
	// headers are optional extra request headers, and body is the request body
	// (may be nil).  Returns the HTTP status code, response headers, and body.
	HTTPProxy(ctx context.Context, method, targetURL string, headers map[string]string, body []byte) (*types.HTTPProxyResponse, error)

	// Runtime type identification
	Type() types.RuntimeType
}
