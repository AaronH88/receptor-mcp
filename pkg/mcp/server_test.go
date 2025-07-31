package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	if server.info.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", server.info.Name)
	}
	
	if server.info.Version != "1.0.0" {
		t.Errorf("Expected server version '1.0.0', got '%s'", server.info.Version)
	}
	
	if server.capabilities.Tools == nil {
		t.Error("Expected tools capability to be initialized")
	}
	
	if server.capabilities.Resources == nil {
		t.Error("Expected resources capability to be initialized")
	}
	
	if server.capabilities.Prompts == nil {
		t.Error("Expected prompts capability to be initialized")
	}
}

func TestRegisterTool(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return map[string]string{"result": "success"}, nil
	}
	
	server.RegisterTool(tool, handler)
	
	// Check tool was registered
	if _, exists := server.tools["test_tool"]; !exists {
		t.Error("Tool was not registered")
	}
	
	// Check handler was registered
	if _, exists := server.handlers["tool_test_tool"]; !exists {
		t.Error("Tool handler was not registered")
	}
}

func TestRegisterResource(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	resource := Resource{
		URI:         "test://resource",
		Name:        "Test Resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return ResourcesReadResponse{
			Contents: []ResourceContent{
				{
					URI:      "test://resource",
					MimeType: "text/plain",
					Text:     "test content",
				},
			},
		}, nil
	}
	
	server.RegisterResource(resource, handler)
	
	// Check resource was registered
	if _, exists := server.resources["test://resource"]; !exists {
		t.Error("Resource was not registered")
	}
	
	// Check handler was registered
	if _, exists := server.handlers["resource_test://resource"]; !exists {
		t.Error("Resource handler was not registered")
	}
}

func TestRegisterPrompt(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	prompt := Prompt{
		Name:        "test_prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{Name: "arg1", Description: "First argument", Required: true},
		},
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return PromptsGetResponse{
			Description: "Test prompt response",
			Messages: []Message{
				{
					Role: "user",
					Content: []Content{
						{Type: "text", Text: "Test prompt content"},
					},
				},
			},
		}, nil
	}
	
	server.RegisterPrompt(prompt, handler)
	
	// Check prompt was registered
	if _, exists := server.prompts["test_prompt"]; !exists {
		t.Error("Prompt was not registered")
	}
	
	// Check handler was registered
	if _, exists := server.handlers["prompt_test_prompt"]; !exists {
		t.Error("Prompt handler was not registered")
	}
}

func TestHandleInitialize(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	req := InitializeRequest{
		ProtocolVersion: MCPVersion,
		Capabilities: ClientCapabilities{
			Roots: &RootsCapability{ListChanged: true},
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}
	
	params, _ := json.Marshal(req)
	result, err := server.handleInitialize(context.Background(), params)
	
	if err != nil {
		t.Fatalf("handleInitialize returned error: %v", err)
	}
	
	response, ok := result.(InitializeResponse)
	if !ok {
		t.Fatal("handleInitialize did not return InitializeResponse")
	}
	
	if response.ProtocolVersion != MCPVersion {
		t.Errorf("Expected protocol version %s, got %s", MCPVersion, response.ProtocolVersion)
	}
	
	if response.ServerInfo.Name != "test-server" {
		t.Errorf("Expected server name 'test-server', got '%s'", response.ServerInfo.Name)
	}
	
	if response.ServerInfo.Version != "1.0.0" {
		t.Errorf("Expected server version '1.0.0', got '%s'", response.ServerInfo.Version)
	}
}

func TestHandleToolsList(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	// Register a test tool
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{"type": "object"},
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "success", nil
	}
	
	server.RegisterTool(tool, handler)
	
	// Test tools/list
	result, err := server.handleToolsList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleToolsList returned error: %v", err)
	}
	
	response, ok := result.(ToolsListResponse)
	if !ok {
		t.Fatal("handleToolsList did not return ToolsListResponse")
	}
	
	if len(response.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(response.Tools))
	}
	
	if response.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", response.Tools[0].Name)
	}
}

func TestHandleToolsCall(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	// Register a test tool
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{"type": "object"},
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return map[string]string{"result": "success"}, nil
	}
	
	server.RegisterTool(tool, handler)
	
	// Test tools/call
	req := ToolsCallRequest{
		Name: "test_tool",
		Arguments: map[string]interface{}{
			"param1": "value1",
		},
	}
	
	params, _ := json.Marshal(req)
	result, err := server.handleToolsCall(context.Background(), params)
	
	if err != nil {
		t.Fatalf("handleToolsCall returned error: %v", err)
	}
	
	response, ok := result.(ToolsCallResponse)
	if !ok {
		t.Fatal("handleToolsCall did not return ToolsCallResponse")
	}
	
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(response.Content))
	}
	
	if !strings.Contains(response.Content[0].Text, "result:success") {
		t.Errorf("Expected content to contain 'result:success', got '%s'", response.Content[0].Text)
	}
}

func TestHandleResourcesList(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	// Register a test resource
	resource := Resource{
		URI:         "test://resource",
		Name:        "Test Resource",
		Description: "A test resource",
		MimeType:    "text/plain",
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "success", nil
	}
	
	server.RegisterResource(resource, handler)
	
	// Test resources/list
	result, err := server.handleResourcesList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleResourcesList returned error: %v", err)
	}
	
	response, ok := result.(ResourcesListResponse)
	if !ok {
		t.Fatal("handleResourcesList did not return ResourcesListResponse")
	}
	
	if len(response.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(response.Resources))
	}
	
	if response.Resources[0].URI != "test://resource" {
		t.Errorf("Expected resource URI 'test://resource', got '%s'", response.Resources[0].URI)
	}
}

func TestHandlePromptsList(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	// Register a test prompt
	prompt := Prompt{
		Name:        "test_prompt",
		Description: "A test prompt",
		Arguments: []PromptArgument{
			{Name: "arg1", Description: "First argument", Required: true},
		},
	}
	
	handler := func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return "success", nil
	}
	
	server.RegisterPrompt(prompt, handler)
	
	// Test prompts/list
	result, err := server.handlePromptsList(context.Background(), nil)
	if err != nil {
		t.Fatalf("handlePromptsList returned error: %v", err)
	}
	
	response, ok := result.(PromptsListResponse)
	if !ok {
		t.Fatal("handlePromptsList did not return PromptsListResponse")
	}
	
	if len(response.Prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(response.Prompts))
	}
	
	if response.Prompts[0].Name != "test_prompt" {
		t.Errorf("Expected prompt name 'test_prompt', got '%s'", response.Prompts[0].Name)
	}
}

func TestIsInitialized(t *testing.T) {
	server := NewServer("test-server", "1.0.0")
	
	// Should not be initialized initially
	if server.IsInitialized() {
		t.Error("Server should not be initialized initially")
	}
	
	// Simulate initialization
	_, err := server.handleInitialized(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleInitialized returned error: %v", err)
	}
	
	// Should be initialized now
	if !server.IsInitialized() {
		t.Error("Server should be initialized after handleInitialized")
	}
}

func TestJSONRPCTypes(t *testing.T) {
	// Test JSONRPCRequest marshaling
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      "test-id",
		Method:  "test/method",
		Params:  map[string]string{"param": "value"},
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal JSONRPCRequest: %v", err)
	}
	
	var unmarshaled JSONRPCRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSONRPCRequest: %v", err)
	}
	
	if unmarshaled.JSONRPC != "2.0" {
		t.Errorf("Expected JSONRPC '2.0', got '%s'", unmarshaled.JSONRPC)
	}
	
	if unmarshaled.Method != "test/method" {
		t.Errorf("Expected method 'test/method', got '%s'", unmarshaled.Method)
	}
}

func TestErrorCodes(t *testing.T) {
	expectedCodes := map[string]int{
		"ParseError":            -32700,
		"InvalidRequest":        -32600,
		"MethodNotFound":        -32601,
		"InvalidParams":         -32602,
		"InternalError":         -32603,
		"InvalidRequestError":   -32000,
		"InternalServerError":   -32001,
	}
	
	actualCodes := map[string]int{
		"ParseError":            ParseError,
		"InvalidRequest":        InvalidRequest,
		"MethodNotFound":        MethodNotFound,
		"InvalidParams":         InvalidParams,
		"InternalError":         InternalError,
		"InvalidRequestError":   InvalidRequestError,
		"InternalServerError":   InternalServerError,
	}
	
	for name, expected := range expectedCodes {
		if actual, exists := actualCodes[name]; !exists {
			t.Errorf("Error code %s not defined", name)
		} else if actual != expected {
			t.Errorf("Expected error code %s to be %d, got %d", name, expected, actual)
		}
	}
}