package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ansible/receptor-mcp/pkg/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	appName    = "receptor-mcp-server"
	appVersion = "1.0.0"
)

var (
	cfgFile string
	debug   bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "MCP server for Ansible Receptor",
	Long: `receptor-mcp-server provides a Model Context Protocol (MCP) interface
to Ansible Receptor mesh networks, enabling AI applications like Claude Desktop
to interact with distributed work execution systems.`,
	Version: appVersion,
	RunE:    runServer,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./receptor-mcp.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// Server configuration flags
	rootCmd.Flags().String("receptor-socket", "/tmp/receptor/receptor.sock", "path to Receptor control socket")
	rootCmd.Flags().StringSlice("receptor-nodes", []string{"localhost"}, "list of Receptor nodes to connect to")
	rootCmd.Flags().Duration("timeout", 30, "default timeout for Receptor operations (seconds)")
	rootCmd.Flags().Bool("tls-verify", true, "verify TLS certificates for Receptor connections")

	// Bind flags to viper
	viper.BindPFlag("receptor.socket", rootCmd.Flags().Lookup("receptor-socket"))
	viper.BindPFlag("receptor.nodes", rootCmd.Flags().Lookup("receptor-nodes"))
	viper.BindPFlag("receptor.timeout", rootCmd.Flags().Lookup("timeout"))
	viper.BindPFlag("receptor.tls_verify", rootCmd.Flags().Lookup("tls-verify"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("receptor-mcp")
	}

	// Environment variables
	viper.SetEnvPrefix("RECEPTOR_MCP")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("receptor.socket", "/tmp/receptor/receptor.sock")
	viper.SetDefault("receptor.nodes", []string{"localhost"})
	viper.SetDefault("receptor.timeout", 30)
	viper.SetDefault("receptor.tls_verify", true)
	viper.SetDefault("debug", false)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintf(os.Stderr, "Received shutdown signal, stopping server...\n")
		cancel()
	}()

	// Create MCP server
	server := mcp.NewServer(appName, appVersion)

	// Configure logging
	if viper.GetBool("debug") {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		fmt.Fprintf(os.Stderr, "Debug logging enabled\n")
	}

	// Register Receptor tools (placeholder implementations for Phase 1)
	registerReceptorTools(server)
	registerReceptorResources(server)
	registerReceptorPrompts(server)

	// Log configuration
	fmt.Fprintf(os.Stderr, "Starting %s v%s\n", appName, appVersion)
	fmt.Fprintf(os.Stderr, "Receptor socket: %s\n", viper.GetString("receptor.socket"))
	fmt.Fprintf(os.Stderr, "Receptor nodes: %v\n", viper.GetStringSlice("receptor.nodes"))
	fmt.Fprintf(os.Stderr, "Ready for MCP communication via stdio\n")

	// Start the MCP server
	return server.Run(ctx)
}

// registerReceptorTools registers the 7 Receptor tools defined in the design
func registerReceptorTools(server *mcp.Server) {
	// Tool 1: submit_work
	server.RegisterTool(mcp.Tool{
		Name:        "submit_work",
		Description: "Submit work to a Receptor node for execution",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"node_id": map[string]interface{}{
					"type":        "string",
					"description": "Target node ID for work execution",
				},
				"work_type": map[string]interface{}{
					"type":        "string",
					"description": "Type of work to execute (e.g., ai-script, compute-task)",
				},
				"payload": map[string]interface{}{
					"type":        "string",
					"description": "Work payload data",
				},
				"params": map[string]interface{}{
					"type":        "object",
					"description": "Additional parameters for work execution",
				},
			},
			"required": []string{"node_id", "work_type", "payload"},
		},
	}, handleSubmitWork)

	// Tool 2: get_work_status
	server.RegisterTool(mcp.Tool{
		Name:        "get_work_status",
		Description: "Get the status of submitted work",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"work_id": map[string]interface{}{
					"type":        "string",
					"description": "Work ID returned from submit_work",
				},
			},
			"required": []string{"work_id"},
		},
	}, handleGetWorkStatus)

	// Tool 3: list_nodes
	server.RegisterTool(mcp.Tool{
		Name:        "list_nodes",
		Description: "List all nodes in the Receptor mesh",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"filter": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter for node selection",
				},
			},
		},
	}, handleListNodes)

	// Tool 4: get_node_info
	server.RegisterTool(mcp.Tool{
		Name:        "get_node_info",
		Description: "Get detailed information about a specific node",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"node_id": map[string]interface{}{
					"type":        "string",
					"description": "Node ID to get information for",
				},
			},
			"required": []string{"node_id"},
		},
	}, handleGetNodeInfo)

	// Tool 5: get_mesh_status
	server.RegisterTool(mcp.Tool{
		Name:        "get_mesh_status",
		Description: "Get overall mesh network status and topology",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
	}, handleGetMeshStatus)

	// Tool 6: cancel_work
	server.RegisterTool(mcp.Tool{
		Name:        "cancel_work",
		Description: "Cancel running or pending work",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"work_id": map[string]interface{}{
					"type":        "string",
					"description": "Work ID to cancel",
				},
			},
			"required": []string{"work_id"},
		},
	}, handleCancelWork)

	// Tool 7: get_work_results
	server.RegisterTool(mcp.Tool{
		Name:        "get_work_results",
		Description: "Retrieve results from completed work",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"work_id": map[string]interface{}{
					"type":        "string",
					"description": "Work ID to get results for",
				},
			},
			"required": []string{"work_id"},
		},
	}, handleGetWorkResults)
}

// registerReceptorResources registers the 4 Receptor resources defined in the design
func registerReceptorResources(server *mcp.Server) {
	// Resource 1: mesh_topology
	server.RegisterResource(mcp.Resource{
		URI:         "receptor://mesh/topology",
		Name:        "Mesh Topology",
		Description: "Real-time mesh network topology information",
		MimeType:    "application/json",
	}, handleMeshTopologyResource)

	// Resource 2: node_status
	server.RegisterResource(mcp.Resource{
		URI:         "receptor://nodes/status",
		Name:        "Node Status",
		Description: "Current status of all nodes in the mesh",
		MimeType:    "application/json",
	}, handleNodeStatusResource)

	// Resource 3: work_queue
	server.RegisterResource(mcp.Resource{
		URI:         "receptor://work/queue",
		Name:        "Work Queue",
		Description: "Active and pending work items",
		MimeType:    "application/json",
	}, handleWorkQueueResource)

	// Resource 4: work_history
	server.RegisterResource(mcp.Resource{
		URI:         "receptor://work/history",
		Name:        "Work History",
		Description: "Historical work execution data",
		MimeType:    "application/json",
	}, handleWorkHistoryResource)
}

// registerReceptorPrompts registers the 3 Receptor prompts defined in the design
func registerReceptorPrompts(server *mcp.Server) {
	// Prompt 1: deploy_workflow
	server.RegisterPrompt(mcp.Prompt{
		Name:        "deploy_workflow",
		Description: "Guide for deploying complex workflows across the mesh",
		Arguments: []mcp.PromptArgument{
			{Name: "workflow_type", Description: "Type of workflow to deploy", Required: true},
			{Name: "target_nodes", Description: "Target nodes for deployment", Required: false},
		},
	}, handleDeployWorkflowPrompt)

	// Prompt 2: troubleshoot_mesh
	server.RegisterPrompt(mcp.Prompt{
		Name:        "troubleshoot_mesh",
		Description: "Mesh network troubleshooting assistant",
		Arguments: []mcp.PromptArgument{
			{Name: "issue_type", Description: "Type of issue being experienced", Required: false},
		},
	}, handleTroubleshootMeshPrompt)

	// Prompt 3: optimize_workload
	server.RegisterPrompt(mcp.Prompt{
		Name:        "optimize_workload",
		Description: "Workload optimization recommendations",
		Arguments: []mcp.PromptArgument{
			{Name: "workload_pattern", Description: "Current workload pattern", Required: false},
			{Name: "performance_goals", Description: "Performance optimization goals", Required: false},
		},
	}, handleOptimizeWorkloadPrompt)
}

// Placeholder tool handlers (Phase 1 - basic responses)
func handleSubmitWork(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"work_id": "work_123456",
		"status":  "submitted",
		"message": "Work submitted successfully (Phase 1 placeholder)",
	}, nil
}

func handleGetWorkStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"work_id": "work_123456",
		"status":  "running",
		"progress": 50,
		"message": "Work is currently running (Phase 1 placeholder)",
	}, nil
}

func handleListNodes(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "controller", "status": "connected", "type": "controller"},
			{"id": "worker-01", "status": "connected", "type": "worker"},
			{"id": "worker-02", "status": "connected", "type": "worker"},
		},
		"message": "Node list (Phase 1 placeholder)",
	}, nil
}

func handleGetNodeInfo(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"node_id": "controller",
		"status":  "connected",
		"capabilities": []string{"work-command", "control-service"},
		"connections": []string{"worker-01", "worker-02"},
		"message": "Node info (Phase 1 placeholder)",
	}, nil
}

func handleGetMeshStatus(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"topology": "mesh",
		"nodes":    3,
		"connections": 2,
		"health": "healthy",
		"message": "Mesh status (Phase 1 placeholder)",
	}, nil
}

func handleCancelWork(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"work_id": "work_123456",
		"status":  "cancelled",
		"message": "Work cancelled successfully (Phase 1 placeholder)",
	}, nil
}

func handleGetWorkResults(ctx context.Context, params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"work_id": "work_123456",
		"status":  "completed",
		"results": "Task completed successfully",
		"logs":    "2024-01-01 12:00:00 Task started\n2024-01-01 12:01:00 Task completed",
		"message": "Work results (Phase 1 placeholder)",
	}, nil
}

// Placeholder resource handlers
func handleMeshTopologyResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	content := mcp.ResourceContent{
		URI:      "receptor://mesh/topology",
		MimeType: "application/json",
		Text:     `{"topology": "mesh", "nodes": ["controller", "worker-01", "worker-02"], "connections": [{"from": "controller", "to": "worker-01"}, {"from": "controller", "to": "worker-02"}]}`,
	}
	return mcp.ResourcesReadResponse{Contents: []mcp.ResourceContent{content}}, nil
}

func handleNodeStatusResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	content := mcp.ResourceContent{
		URI:      "receptor://nodes/status",
		MimeType: "application/json",
		Text:     `{"nodes": [{"id": "controller", "status": "connected", "load": 0.2}, {"id": "worker-01", "status": "connected", "load": 0.5}, {"id": "worker-02", "status": "connected", "load": 0.3}]}`,
	}
	return mcp.ResourcesReadResponse{Contents: []mcp.ResourceContent{content}}, nil
}

func handleWorkQueueResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	content := mcp.ResourceContent{
		URI:      "receptor://work/queue",
		MimeType: "application/json",
		Text:     `{"active": [{"work_id": "work_123456", "status": "running", "node_id": "worker-01"}], "pending": []}`,
	}
	return mcp.ResourcesReadResponse{Contents: []mcp.ResourceContent{content}}, nil
}

func handleWorkHistoryResource(ctx context.Context, params json.RawMessage) (interface{}, error) {
	content := mcp.ResourceContent{
		URI:      "receptor://work/history",
		MimeType: "application/json",
		Text:     `{"completed": [{"work_id": "work_123455", "status": "completed", "node_id": "worker-01", "duration": "30s"}], "failed": []}`,
	}
	return mcp.ResourcesReadResponse{Contents: []mcp.ResourceContent{content}}, nil
}

// Placeholder prompt handlers
func handleDeployWorkflowPrompt(ctx context.Context, params json.RawMessage) (interface{}, error) {
	messages := []mcp.Message{
		{
			Role: "user",
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "I need help deploying a workflow across my Receptor mesh. Here are the steps to consider:\n\n1. **Assess Current Mesh Status**: First, check the health and capacity of your mesh nodes\n2. **Define Workflow Requirements**: Specify the work types, dependencies, and resource requirements\n3. **Plan Node Distribution**: Choose optimal nodes based on capabilities and current load\n4. **Submit Work in Sequence**: Deploy workflow components in the correct order\n5. **Monitor Progress**: Track execution and handle any failures\n\nWhat type of workflow would you like to deploy?",
				},
			},
		},
	}
	return mcp.PromptsGetResponse{
		Description: "Workflow deployment guidance",
		Messages:    messages,
	}, nil
}

func handleTroubleshootMeshPrompt(ctx context.Context, params json.RawMessage) (interface{}, error) {
	messages := []mcp.Message{
		{
			Role: "user",
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Let's troubleshoot your Receptor mesh network. Common issues and solutions:\n\n**Connection Issues:**\n- Check network connectivity between nodes\n- Verify firewall rules and port accessibility\n- Confirm TLS certificates are valid\n\n**Performance Issues:**\n- Monitor node resource usage (CPU, memory)\n- Check work queue backlogs\n- Analyze network latency between nodes\n\n**Work Execution Problems:**\n- Verify work types are properly configured\n- Check node capabilities and permissions\n- Review work execution logs\n\nWhat specific issue are you experiencing?",
				},
			},
		},
	}
	return mcp.PromptsGetResponse{
		Description: "Mesh troubleshooting guidance",
		Messages:    messages,
	}, nil
}

func handleOptimizeWorkloadPrompt(ctx context.Context, params json.RawMessage) (interface{}, error) {
	messages := []mcp.Message{
		{
			Role: "user",
			Content: []mcp.Content{
				{
					Type: "text",
					Text: "Here are strategies to optimize your Receptor workloads:\n\n**Load Balancing:**\n- Distribute work evenly across available nodes\n- Use node capabilities to match work types\n- Monitor and adjust based on node performance\n\n**Resource Optimization:**\n- Configure appropriate work concurrency limits\n- Optimize work payload sizes\n- Use work signing for security without performance impact\n\n**Network Efficiency:**\n- Minimize data transfer between nodes\n- Use local resources when possible\n- Consider edge nodes for geographically distributed work\n\n**Monitoring and Tuning:**\n- Track work execution times and success rates\n- Monitor resource utilization trends\n- Adjust timeout values based on work complexity\n\nWhat aspect of your workload would you like to optimize?",
				},
			},
		},
	}
	return mcp.PromptsGetResponse{
		Description: "Workload optimization recommendations",
		Messages:    messages,
	}, nil
}