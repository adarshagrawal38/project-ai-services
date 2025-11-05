package vars

import "regexp"

var (
	SpyreCardAnnotationRegex = regexp.MustCompile(`^ai-services\.io\/([A-Za-z0-9][-A-Za-z0-9_.]*)--sypre-cards$`)
	ToolImage                = "icr.io/ai-services-private/tools:latest"
	ModelDirectory           = "/var/lib/ai-services/models"
)
