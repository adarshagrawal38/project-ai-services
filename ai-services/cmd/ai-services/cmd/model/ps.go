package model

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/project-ai-services/ai-services/internal/pkg/catalog/apiserver/models"
	catalogClient "github.com/project-ai-services/ai-services/internal/pkg/catalog/client"
	cliUtils "github.com/project-ai-services/ai-services/internal/pkg/cli/utils"
	"github.com/project-ai-services/ai-services/internal/pkg/logger"
	"github.com/project-ai-services/ai-services/internal/pkg/utils"
)

// PS table column indices (shared across normal and wide output).
const (
	psColModelName = 0
	psColWorker    = 1
	psColRuntime   = 2
)

type psOptions struct {
	name           string
	componentType  string
	deploymentType string
	provider       string
	status         string
	output         string
}

func (o psOptions) wide() bool {
	return strings.ToLower(o.output) == "wide"
}

func newPsCmd() *cobra.Command {
	var opts psOptions

	cmd := &cobra.Command{
		Use:   "ps [name]",
		Short: "Lists all or specified model(s) with their runtime status",
		Long: `Retrieves the process and pod status of all models (local and remote) if no name is provided.
Lists the status of a specific model if the name is provided.

Remote models have no pods, so their worker, runtime, namespace and pod columns are shown as '-'.

Arguments:
  [name] : Model name (optional)`,
		Example: `  # List all models
  ai-services model ps

  # List a specific model
  ai-services model ps granite-llm

  # List only local LLM models
  ai-services model ps --type llm --deployment-type local

  # List only running local models with wide output
  ai-services model ps --status Running -o wide`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Once precheck passes, silence usage for any *later* internal errors.
			cmd.SilenceUsage = true

			if len(args) > 0 {
				opts.name = args[0]
			}

			return runPs(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.componentType, "type", "t", "", "Filter by model role: llm, embedding, reranker")
	cmd.Flags().StringVar(&opts.deploymentType, "deployment-type", "", "Filter by kind: local, remote")
	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "", "Filter by provider: vllm-spyre, vllm-cpu, watsonx, hosted_vllm")
	cmd.Flags().StringVar(&opts.status, "status", "", "Filter by status: Running, Deploying, Error (local); connected, offline (remote)")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Output format (e.g., wide)")

	return cmd
}

func runPs(ctx context.Context, opts psOptions) error {
	modelClient, err := catalogClient.NewModelClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create model client: %w", err)
	}

	filters := catalogClient.ListModelsParams{
		ComponentType:  opts.componentType,
		DeploymentType: opts.deploymentType,
		Status:         opts.status,
	}
	if opts.provider != "" {
		filters.ProviderIDs = []string{opts.provider}
	}

	items, err := cliUtils.FetchModels(ctx, modelClient, opts.name, filters)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		logger.Warningln("No Model found")

		return nil
	}

	printer := utils.NewTableWriter()
	defer printer.CloseTableWriter()

	setModelPSTableHeaders(printer, opts.wide())

	for _, item := range items {
		psResp, err := modelClient.GetModelPS(ctx, item.ID)
		if err != nil {
			return fmt.Errorf("failed to fetch model '%s': %w", item.Name, err)
		}

		for _, row := range buildModelPSRows(psResp, opts.wide()) {
			printer.AppendRow(row...)
		}
	}

	return nil
}

// setModelPSTableHeaders sets the table headers and collapse indices based on output format.
func setModelPSTableHeaders(printer *utils.Printer, outputWide bool) {
	if outputWide {
		printer.SetHeaders("MODEL NAME", "WORKER", "RUNTIME", "NAMESPACE", "POD ID", "POD NAME", "STATUS", "CREATED", "CONTAINERS")
	} else {
		printer.SetHeaders("MODEL NAME", "WORKER", "RUNTIME", "NAMESPACE", "POD NAME", "STATUS")
	}

	printer.SetCollapseIndices(psColModelName, psColWorker, psColRuntime)
}

// buildModelPSRows builds the table rows of a model, one per pod. A model without pods
// (remote, or a local model that is not deployed yet) gets a single row with '-' placeholders
// and the model's own status.
func buildModelPSRows(ps *models.ModelPSResponse, outputWide bool) [][]string {
	if len(ps.Pods) == 0 {
		status := cmp.Or(ps.Status, emptyValue)
		worker := cmp.Or(ps.WorkerName, emptyValue)
		runtime := cmp.Or(ps.RuntimeType, emptyValue)
		namespace := cmp.Or(ps.Namespace, emptyValue)

		if outputWide {
			return [][]string{{ps.Name, worker, runtime, namespace, emptyValue, emptyValue, status, emptyValue, emptyValue}}
		}

		return [][]string{{ps.Name, worker, runtime, namespace, emptyValue, status}}
	}

	rows := make([][]string, 0, len(ps.Pods))
	for _, pod := range ps.Pods {
		rows = append(rows, cliUtils.BuildPodRowFromAPI(ps.Name, ps.WorkerName, ps.Namespace, ps.RuntimeType, pod, outputWide))
	}

	return rows
}
