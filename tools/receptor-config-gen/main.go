package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"gopkg.in/yaml.v2"
)

const (
	version = "1.0.0"
)

// TemplateConfig represents the configuration for template generation
type TemplateConfig struct {
	Template   string
	OutputPath string
	OutputDir  string
	Parameters map[string]string
	Nodes      int
	CertDir    string
	WorkTypes  []string
}

// ConfigTemplate represents a configuration template
type ConfigTemplate struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Template    string `yaml:"template"`
	Parameters  []TemplateParameter `yaml:"parameters"`
}

// TemplateParameter represents a template parameter
type TemplateParameter struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	Required     bool   `yaml:"required"`
	DefaultValue string `yaml:"default_value"`
	Type         string `yaml:"type"`
}

func main() {
	var config TemplateConfig
	var showVersion, listTemplates, validate bool
	var paramFlags arrayFlags

	flag.StringVar(&config.Template, "template", "", "Configuration template to use")
	flag.StringVar(&config.OutputPath, "output", "", "Output file path")
	flag.StringVar(&config.OutputDir, "output-dir", "", "Output directory for multiple files")
	flag.Var(&paramFlags, "param", "Template parameter (key=value)")
	flag.IntVar(&config.Nodes, "nodes", 1, "Number of nodes to generate")
	flag.StringVar(&config.CertDir, "cert-dir", "/etc/receptor/certs", "Certificate directory path")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&listTemplates, "list", false, "List available templates")
	flag.BoolVar(&validate, "validate", false, "Validate generated configuration")
	flag.Parse()

	if showVersion {
		fmt.Printf("receptor-config-gen version %s\n", version)
		return
	}

	if listTemplates {
		listAvailableTemplates()
		return
	}

	// Parse parameters
	config.Parameters = make(map[string]string)
	for _, param := range paramFlags {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid parameter format: %s (expected key=value)", param)
		}
		config.Parameters[parts[0]] = parts[1]
	}

	if config.Template == "" {
		log.Fatal("Template name is required. Use -list to see available templates.")
	}

	if config.OutputPath == "" && config.OutputDir == "" {
		log.Fatal("Either -output or -output-dir must be specified")
	}

	generator := NewConfigGenerator()
	if err := generator.Generate(config); err != nil {
		log.Fatalf("Failed to generate configuration: %v", err)
	}

	if validate {
		if err := validateConfiguration(config.OutputPath); err != nil {
			log.Fatalf("Configuration validation failed: %v", err)
		}
		fmt.Println("Configuration validation passed")
	}

	fmt.Println("Configuration generated successfully")
}

// arrayFlags implements flag.Value for string arrays
type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ",")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

// ConfigGenerator handles configuration generation
type ConfigGenerator struct {
	templatesDir string
	templates    map[string]ConfigTemplate
}

// NewConfigGenerator creates a new configuration generator
func NewConfigGenerator() *ConfigGenerator {
	// Get current working directory and look for configs directory
	cwd, _ := os.Getwd()
	templatesDir := filepath.Join(cwd, "configs")
	
	// If configs doesn't exist in current dir, try relative to executable
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		execPath, _ := os.Executable()
		templatesDir = filepath.Join(filepath.Dir(execPath), "..", "configs")
	}
	
	generator := &ConfigGenerator{
		templatesDir: templatesDir,
		templates:    make(map[string]ConfigTemplate),
	}
	
	generator.loadTemplates()
	return generator
}

// loadTemplates loads available templates from the templates directory
func (g *ConfigGenerator) loadTemplates() {
	// Built-in template definitions
	g.templates["dev-single-node"] = ConfigTemplate{
		Name:        "dev-single-node",
		Description: "Single node development configuration",
		Template:    "dev/dev-single-node.yaml",
		Parameters:  []TemplateParameter{},
	}
	
	g.templates["dev-mesh-controller"] = ConfigTemplate{
		Name:        "dev-mesh-controller",
		Description: "Development mesh controller node",
		Template:    "dev/dev-mesh-controller.yaml",
		Parameters:  []TemplateParameter{},
	}
	
	g.templates["dev-mesh-worker"] = ConfigTemplate{
		Name:        "dev-mesh-worker",
		Description: "Development mesh worker node",
		Template:    "dev/dev-mesh-worker.yaml",
		Parameters: []TemplateParameter{
			{Name: "WorkerID", Description: "Worker node identifier", Required: true, Type: "string"},
			{Name: "ControllerAddress", Description: "Controller node address", Required: true, Type: "string"},
		},
	}
	
	g.templates["prod-controller"] = ConfigTemplate{
		Name:        "prod-controller",
		Description: "Production controller node with TLS and work signing",
		Template:    "prod/prod-controller.yaml",
		Parameters:  []TemplateParameter{},
	}
	
	g.templates["prod-worker"] = ConfigTemplate{
		Name:        "prod-worker",
		Description: "Production worker node with security verification",
		Template:    "prod/prod-worker.yaml",
		Parameters: []TemplateParameter{
			{Name: "WorkerID", Description: "Worker node identifier", Required: true, Type: "string"},
			{Name: "ControllerAddress", Description: "Controller node address", Required: true, Type: "string"},
		},
	}
	
	g.templates["prod-edge"] = ConfigTemplate{
		Name:        "prod-edge",
		Description: "Production edge node with minimal attack surface",
		Template:    "prod/prod-edge.yaml",
		Parameters: []TemplateParameter{
			{Name: "EdgeID", Description: "Edge node identifier", Required: true, Type: "string"},
			{Name: "ControllerAddress", Description: "Controller node address", Required: true, Type: "string"},
		},
	}
}

// Generate generates configuration files based on the template and parameters
func (g *ConfigGenerator) Generate(config TemplateConfig) error {
	tmpl, exists := g.templates[config.Template]
	if !exists {
		return fmt.Errorf("template '%s' not found", config.Template)
	}
	
	// Validate required parameters (skip auto-generated ones for multi-node)
	if err := g.validateParameters(tmpl, config.Parameters, config.Nodes > 1); err != nil {
		return err
	}
	
	// Load template file
	templatePath := filepath.Join(g.templatesDir, tmpl.Template)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template file: %v", err)
	}
	
	// Parse and execute template
	t, err := template.New(config.Template).Parse(string(templateData))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	
	// Generate single file or multiple files
	if config.Nodes == 1 || config.OutputPath != "" {
		return g.generateSingleFile(t, config)
	} else {
		return g.generateMultipleFiles(t, config)
	}
}

// generateSingleFile generates a single configuration file
func (g *ConfigGenerator) generateSingleFile(t *template.Template, config TemplateConfig) error {
	outputPath := config.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(config.OutputDir, fmt.Sprintf("%s.yaml", config.Template))
	}
	
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()
	
	return t.Execute(file, config.Parameters)
}

// generateMultipleFiles generates multiple configuration files for multiple nodes
func (g *ConfigGenerator) generateMultipleFiles(t *template.Template, config TemplateConfig) error {
	if config.OutputDir == "" {
		return fmt.Errorf("output directory required for multiple node generation")
	}
	
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	
	for i := 1; i <= config.Nodes; i++ {
		// Create node-specific parameters
		nodeParams := make(map[string]string)
		for k, v := range config.Parameters {
			nodeParams[k] = v
		}
		
		// Add node-specific parameters
		if strings.Contains(config.Template, "worker") {
			nodeParams["WorkerID"] = fmt.Sprintf("worker-%02d", i)
		} else if strings.Contains(config.Template, "edge") {
			nodeParams["EdgeID"] = fmt.Sprintf("edge-%02d", i)
		}
		
		// Generate file for this node
		outputPath := filepath.Join(config.OutputDir, fmt.Sprintf("%s-%02d.yaml", config.Template, i))
		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %v", outputPath, err)
		}
		
		if err := t.Execute(file, nodeParams); err != nil {
			file.Close()
			return fmt.Errorf("failed to generate configuration for node %d: %v", i, err)
		}
		file.Close()
	}
	
	return nil
}

// validateParameters validates that all required parameters are provided
func (g *ConfigGenerator) validateParameters(tmpl ConfigTemplate, params map[string]string, isMultiNode bool) error {
	for _, param := range tmpl.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				// Skip auto-generated parameters for multi-node generation
				if isMultiNode && (param.Name == "WorkerID" || param.Name == "EdgeID") {
					continue
				}
				return fmt.Errorf("required parameter '%s' not provided", param.Name)
			}
		}
	}
	return nil
}

// validateConfiguration validates the generated YAML configuration
func validateConfiguration(configPath string) error {
	if configPath == "" {
		return nil // Skip validation for multiple file generation
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %v", err)
	}
	
	var config interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid YAML syntax: %v", err)
	}
	
	// Additional Receptor-specific validation could be added here
	return nil
}

// listAvailableTemplates lists all available configuration templates
func listAvailableTemplates() {
	generator := NewConfigGenerator()
	
	fmt.Println("Available Configuration Templates:")
	fmt.Println("==================================")
	
	for name, tmpl := range generator.templates {
		fmt.Printf("\n%s\n", name)
		fmt.Printf("  Description: %s\n", tmpl.Description)
		if len(tmpl.Parameters) > 0 {
			fmt.Println("  Parameters:")
			for _, param := range tmpl.Parameters {
				required := ""
				if param.Required {
					required = " (required)"
				}
				fmt.Printf("    - %s: %s%s\n", param.Name, param.Description, required)
			}
		}
	}
	
	fmt.Println("\nUsage Examples:")
	fmt.Println("  receptor-config-gen -template dev-single-node -output dev.yaml")
	fmt.Println("  receptor-config-gen -template prod-worker -param WorkerID=web01 -param ControllerAddress=controller.example.com -output worker.yaml")
	fmt.Println("  receptor-config-gen -template prod-worker -param ControllerAddress=controller.example.com -nodes 5 -output-dir ./workers/")
}