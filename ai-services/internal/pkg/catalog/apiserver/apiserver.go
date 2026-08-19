// Package apiserver provides the implementation of the API server for the AI Services Catalog.
// It includes the setup of routes, authentication, and server configuration.
//
//	@title						AI Services Catalog API
//	@version					1.0
//	@description				API server for managing AI Services catalog, applications, and authentication
//	@termsOfService				http://swagger.io/terms/
//
//	@contact.name				API Support
//	@contact.url				https://github.com/project-ai-services/ai-services
//	@contact.email				support@example.com
//
//	@license.name				Apache 2.0
//	@license.url				http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@tag.name					Authentication
//	@tag.description			Authentication and authorization endpoints
//
//	@tag.name					Applications
//	@tag.description			Application management endpoints
//
//	@tag.name					Catalog
//	@tag.description			Catalog endpoints for architectures and services
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
package apiserver

import (
	"context"
	"fmt"

	"github.com/project-ai-services/ai-services/internal/pkg/agent/gateway"
	"github.com/project-ai-services/ai-services/internal/pkg/agent/registry"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/repository"
	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/services/auth"
	"github.com/project-ai-services/ai-services/internal/pkg/logger"
	"github.com/project-ai-services/ai-services/internal/pkg/worker/gateway"
	"github.com/project-ai-services/ai-services/internal/pkg/worker/registry"
)

// APIServerOptions defines the configuration options for the API server such as the port to listen
// on and the authentication provider.
type APIServerOptions struct {
	Port               int
	AuthService        auth.Service
	TokenManager       *auth.TokenManager
	Blacklist          repository.TokenBlacklist
<<<<<<< ours
	ApplicationService repository.ApplicationServiceInterface

	// WorkerGatewayPort is the port the gRPC worker gateway listens on.
	// Defaults to 9090 when zero.
	WorkerGatewayPort int
	// WorkerRegistry holds the in-memory state of all connected workers and owns
	// the bootstrap token store.
	WorkerRegistry *registry.Registry
=======
	ApplicationService *repository.ApplicationService
	// AgentGateway is optional. When non-nil the gRPC AgentGateway is started
	// alongside the REST server on AgentGatewayPort.
	AgentGateway     *gateway.Gateway
	AgentGatewayPort int // defaults to 9090 when AgentGateway is set
	// AgentTokenStore and AgentRegistry are passed through to the REST handler
	// so admins can issue bootstrap tokens and list agents via the REST API.
	AgentTokenStore *registry.TokenStore
	AgentRegistry   *registry.Registry
>>>>>>> theirs
}

// APIserver represents the API server instance, holding the configuration and authentication provider.
type APIserver struct {
	port               int
	authService        auth.Service
	tokenManager       *auth.TokenManager
	blacklist          repository.TokenBlacklist
<<<<<<< ours
	applicationService repository.ApplicationServiceInterface

	workerGatewayPort int
	workerRegistry    *registry.Registry
=======
	applicationService *repository.ApplicationService
	agentGateway       *gateway.Gateway
	agentGatewayPort   int
	agentTokenStore    *registry.TokenStore
	agentRegistry      *registry.Registry
>>>>>>> theirs
}

// NewAPIserver creates a new instance of the API server with the provided options, setting default values where necessary.
func NewAPIserver(options APIServerOptions) *APIserver {
	if options.Port == 0 {
		options.Port = 8080
	}
<<<<<<< ours
	if options.WorkerGatewayPort == 0 {
		options.WorkerGatewayPort = 9090
=======
	if options.AgentGateway != nil && options.AgentGatewayPort == 0 {
		options.AgentGatewayPort = 9090
>>>>>>> theirs
	}

	return &APIserver{
		port:               options.Port,
		authService:        options.AuthService,
		tokenManager:       options.TokenManager,
		blacklist:          options.Blacklist,
		applicationService: options.ApplicationService,
<<<<<<< ours
		workerGatewayPort:  options.WorkerGatewayPort,
		workerRegistry:     options.WorkerRegistry,
=======
		agentGateway:       options.AgentGateway,
		agentGatewayPort:   options.AgentGatewayPort,
		agentTokenStore:    options.AgentTokenStore,
		agentRegistry:      options.AgentRegistry,
>>>>>>> theirs
	}
}

// Start initializes the API server and begins listening for incoming requests on the configured port.
<<<<<<< ours
// It sets up the router with authentication middleware and routes, and starts the gRPC worker gateway.
// ctx should be a signal-aware context (e.g. from signal.NotifyContext) so that SIGINT/SIGTERM
// trigger graceful shutdown of the gateway and sweeper.
func (a *APIserver) Start(ctx context.Context) error {
	// Wrap with CancelCause so the gateway can abort the whole process if Serve fails.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// Start the gRPC worker gateway.
	gw := gateway.New(a.workerRegistry)
	gatewayAddr := fmt.Sprintf(":%d", a.workerGatewayPort)
	if err := gw.Start(ctx, cancel, gatewayAddr); err != nil {
		return fmt.Errorf("failed to start worker gateway: %w", err)
	}
	logger.InfofCtx(ctx, "Worker gateway started on %s", gatewayAddr)

	r := CreateRouter(a.authService, a.tokenManager, a.blacklist, a.applicationService, a.workerRegistry)

	if err := r.Run(fmt.Sprintf(":%d", a.port)); err != nil {
		return err
	}

	// If ctx was cancelled by a gateway failure, surface that cause.
	if cause := context.Cause(ctx); cause != nil {
		return cause
	}

	return nil
=======
// If an AgentGateway is configured it is started on AgentGatewayPort before the REST server,
// and the heartbeat watcher is started to mark stale agents DISCONNECTED.
func (a *APIserver) Start() error {
	if a.agentGateway != nil {
		ctx := context.Background()
		addr := fmt.Sprintf(":%d", a.agentGatewayPort)
		if err := a.agentGateway.Start(ctx, addr); err != nil {
			return fmt.Errorf("failed to start AgentGateway: %w", err)
		}
		// Start the heartbeat watcher so agents that miss heartbeats are
		// transitioned from READY → DISCONNECTED automatically.
		a.agentRegistry.StartHeartbeatWatcher(ctx)
	}

	r := CreateRouter(a.authService, a.tokenManager, a.blacklist, a.applicationService, a.agentTokenStore, a.agentRegistry)
	return r.Run(fmt.Sprintf(":%d", a.port))
>>>>>>> theirs
}
