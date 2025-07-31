package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Server represents an MCP server instance
type Server struct {
	info         ServerInfo
	capabilities ServerCapabilities
	tools        map[string]Tool
	resources    map[string]Resource
	prompts      map[string]Prompt
	handlers     map[string]Handler
	mu           sync.RWMutex
	initialized  bool
	logger       *log.Logger
}

// Handler represents a method handler function
type Handler func(ctx context.Context, params json.RawMessage) (interface{}, error)

// NewServer creates a new MCP server instance
func NewServer(name, version string) *Server {
	server := &Server{
		info: ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities: ServerCapabilities{
			Logging: &LoggingCapability{},
			Tools:   &ToolsCapability{ListChanged: false},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Prompts: &PromptsCapability{ListChanged: false},
		},
		tools:     make(map[string]Tool),
		resources: make(map[string]Resource),
		prompts:   make(map[string]Prompt),
		handlers:  make(map[string]Handler),
		logger:    log.New(os.Stderr, "[MCP Server] ", log.LstdFlags),
	}

	// Register core MCP handlers
	server.registerCoreHandlers()
	return server
}

// registerCoreHandlers registers the standard MCP protocol handlers
func (s *Server) registerCoreHandlers() {
	s.handlers["initialize"] = s.handleInitialize
	s.handlers["initialized"] = s.handleInitialized
	s.handlers["tools/list"] = s.handleToolsList
	s.handlers["tools/call"] = s.handleToolsCall
	s.handlers["resources/list"] = s.handleResourcesList
	s.handlers["resources/read"] = s.handleResourcesRead
	s.handlers["prompts/list"] = s.handlePromptsList
	s.handlers["prompts/get"] = s.handlePromptsGet
}

// RegisterTool registers a new tool with the server
func (s *Server) RegisterTool(tool Tool, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
	s.handlers["tool_"+tool.Name] = handler
}

// RegisterResource registers a new resource with the server
func (s *Server) RegisterResource(resource Resource, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resource.URI] = resource
	s.handlers["resource_"+resource.URI] = handler
}

// RegisterPrompt registers a new prompt with the server
func (s *Server) RegisterPrompt(prompt Prompt, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts[prompt.Name] = prompt
	s.handlers["prompt_"+prompt.Name] = handler
}

// Run starts the MCP server, reading from stdin and writing to stdout
func (s *Server) Run(ctx context.Context) error {
	s.logger.Printf("Starting MCP server %s v%s", s.info.Name, s.info.Version)
	
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("Server shutting down...")
			return ctx.Err()
		default:
		}

		// Read message line
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Println("Client disconnected")
				return nil
			}
			s.logger.Printf("Error reading input: %v", err)
			continue
		}

		// Process message in goroutine
		go s.processMessage(ctx, line, writer)
	}
}

// processMessage processes a single JSON-RPC message
func (s *Server) processMessage(ctx context.Context, data []byte, writer *bufio.Writer) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.sendError(writer, nil, ParseError, "Parse error", err.Error())
		return
	}

	// Handle notifications (no response expected)
	if req.ID == nil {
		s.handleNotification(ctx, req)
		return
	}

	// Handle regular requests
	s.handleRequest(ctx, req, writer)
}

// handleRequest processes a JSON-RPC request and sends a response
func (s *Server) handleRequest(ctx context.Context, req JSONRPCRequest, writer *bufio.Writer) {
	handler, exists := s.handlers[req.Method]
	if !exists {
		s.sendError(writer, req.ID, MethodNotFound, "Method not found", req.Method)
		return
	}

	var params json.RawMessage
	if req.Params != nil {
		paramBytes, err := json.Marshal(req.Params)
		if err != nil {
			s.sendError(writer, req.ID, InvalidParams, "Invalid params", err.Error())
			return
		}
		params = paramBytes
	}

	result, err := handler(ctx, params)
	if err != nil {
		s.sendError(writer, req.ID, InternalError, "Internal error", err.Error())
		return
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}

	s.sendResponse(writer, response)
}

// handleNotification processes a JSON-RPC notification (no response sent)
func (s *Server) handleNotification(ctx context.Context, req JSONRPCRequest) {
	if handler, exists := s.handlers[req.Method]; exists {
		var params json.RawMessage
		if req.Params != nil {
			paramBytes, _ := json.Marshal(req.Params)
			params = paramBytes
		}
		_, _ = handler(ctx, params)
	}
}

// Core MCP Protocol Handlers

func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req InitializeRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid initialize request: %w", err)
	}

	s.logger.Printf("Initialize request from %s v%s", req.ClientInfo.Name, req.ClientInfo.Version)

	response := InitializeResponse{
		ProtocolVersion: MCPVersion,
		Capabilities:    s.capabilities,
		ServerInfo:      s.info,
	}

	return response, nil
}

func (s *Server) handleInitialized(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.Lock()
	s.initialized = true
	s.mu.Unlock()

	s.logger.Println("Server initialized successfully")
	return nil, nil
}

func (s *Server) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		tools = append(tools, tool)
	}

	return ToolsListResponse{Tools: tools}, nil
}

func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ToolsCallRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid tools/call request: %w", err)
	}

	// Find the tool handler
	handler, exists := s.handlers["tool_"+req.Name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", req.Name)
	}

	// Marshal arguments back to JSON for the handler
	argsBytes, _ := json.Marshal(req.Arguments)
	
	result, err := handler(ctx, argsBytes)
	if err != nil {
		return ToolsCallResponse{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Error executing tool %s: %v", req.Name, err),
			}},
			IsError: true,
		}, nil
	}

	// Convert result to content
	content := []Content{{
		Type: "text",
		Text: fmt.Sprintf("%v", result),
	}}

	return ToolsCallResponse{Content: content}, nil
}

func (s *Server) handleResourcesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}

	return ResourcesListResponse{Resources: resources}, nil
}

func (s *Server) handleResourcesRead(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req ResourcesReadRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid resources/read request: %w", err)
	}

	handler, exists := s.handlers["resource_"+req.URI]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", req.URI)
	}

	result, err := handler(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Server) handlePromptsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}

	return PromptsListResponse{Prompts: prompts}, nil
}

func (s *Server) handlePromptsGet(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var req PromptsGetRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid prompts/get request: %w", err)
	}

	handler, exists := s.handlers["prompt_"+req.Name]
	if !exists {
		return nil, fmt.Errorf("prompt not found: %s", req.Name)
	}

	result, err := handler(ctx, params)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Utility methods

func (s *Server) sendResponse(writer *bufio.Writer, response JSONRPCResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		s.logger.Printf("Error marshaling response: %v", err)
		return
	}

	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}

func (s *Server) sendError(writer *bufio.Writer, id interface{}, code int, message, data string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	s.sendResponse(writer, response)
}

// IsInitialized returns whether the server has been initialized
func (s *Server) IsInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}