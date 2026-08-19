package configure

// ArgParam keys common to both Podman and OpenShift deployments.
const (
	ArgParamAdminPasswordHash = "backend.adminPasswordHash"
	ArgParamDBPassword        = "db.password"
)

// ArgParam keys used only by the OpenShift deployment.
// (no OpenShift-specific params at present; extend here when needed)
// Run executes the configure process for the catalog service.
// It creates runtime-specific options and calls the appropriate runtime implementation.
func Run(runtime types.RuntimeType, baseDir, domainName, sslCertPath, sslKeyPath string, httpsPort, agentGatewayPort int) error {
	ctx := context.Background()
	// Deploy catalog service based on runtime
	switch runtime {
	case types.RuntimeTypePodman:
		opts := catalogUtils.PodmanConfigureOptions{
			BaseDir:          baseDir,
			DomainName:       domainName,
			SSLCertPath:      sslCertPath,
			SSLKeyPath:       sslKeyPath,
			HttpsPort:        httpsPort,
			AgentGatewayPort: agentGatewayPort,
		}

		return catalogPodman.DeployCatalog(ctx, opts)

	case types.RuntimeTypeOpenShift:
		return fmt.Errorf("openshift runtime is not yet supported for catalog configure")

// ArgParam keys used only by the Podman deployment.
const (
	ArgParamRuntime               = "backend.runtime"
	ArgParamPodmanAuthFileContent = "backend.podman.authFileContent"
	ArgParamPodmanURI             = "backend.podman.uri"
	ArgParamCaddyHTTPSPort        = "caddy.httpsPort"
	ArgParamWorkerGatewayPort     = "backend.workerGatewayPort"
)
