package mcp

// MCP Protocol Version
const MCPVersion = "2024-11-05"

// JSON-RPC 2.0 Message Types
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type JSONRPCNotification struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCP Initialize Request/Response
type InitializeRequest struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

type InitializeResponse struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

type ServerCapabilities struct {
	Logging   *LoggingCapability   `json:"logging,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapability struct{}

type LoggingCapability struct{}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// MCP Tool Types
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolsListRequest struct{}

type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
}

type ToolsCallRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type ToolsCallResponse struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

// MCP Resource Types
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

type ResourcesListRequest struct{}

type ResourcesListResponse struct {
	Resources []Resource `json:"resources"`
}

type ResourcesReadRequest struct {
	URI string `json:"uri"`
}

type ResourcesReadResponse struct {
	Contents []ResourceContent `json:"contents"`
}

type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// MCP Prompt Types
type Prompt struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Arguments   []PromptArgument         `json:"arguments,omitempty"`
}

type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

type PromptsListRequest struct{}

type PromptsListResponse struct {
	Prompts []Prompt `json:"prompts"`
}

type PromptsGetRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type PromptsGetResponse struct {
	Description string    `json:"description,omitempty"`
	Messages    []Message `json:"messages"`
}

// Common Content Types
type Content struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

// Logging Types
type LoggingLevel string

const (
	LogLevelDebug LoggingLevel = "debug"
	LogLevelInfo  LoggingLevel = "info"
	LogLevelNotice LoggingLevel = "notice"
	LogLevelWarning LoggingLevel = "warning"
	LogLevelError LoggingLevel = "error"
	LogLevelCritical LoggingLevel = "critical"
	LogLevelAlert LoggingLevel = "alert"
	LogLevelEmergency LoggingLevel = "emergency"
)

type LoggingMessageNotification struct {
	Level  LoggingLevel `json:"level"`
	Data   interface{}  `json:"data,omitempty"`
	Logger string       `json:"logger,omitempty"`
}

// Standard JSON-RPC Error Codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

// MCP-specific Error Codes
const (
	InvalidRequestError = -32000
	InternalServerError = -32001
)