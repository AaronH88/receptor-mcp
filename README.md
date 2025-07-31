# Receptor MCP Server

A Model Context Protocol (MCP) server that exposes Ansible Receptor's mesh networking and work execution capabilities to AI applications like Claude Desktop and Claude Code.

## 🚧 Current Status: MCP Server Infrastructure

**Status**: Complete MCP server foundation with placeholder implementations

### What's Working

- ✅ **Full MCP Protocol Support** - JSON-RPC 2.0 over stdio implementation
- ✅ **All 7 Receptor Tools** - Complete tool definitions with placeholder responses
- ✅ **All 4 Resources** - Resource endpoints with mock data
- ✅ **All 3 Prompts** - Guided workflow prompts with helpful content
- ✅ **Configuration System** - YAML-based configuration and CLI arguments
- ✅ **Project Infrastructure** - Configuration templates, deployment scripts
- ✅ **Build System** - Go modules, configuration generator tool

### Current Implementation Status

**Phase 1: MCP Server Foundation** ✅ **COMPLETE**
- Functional MCP server that can integrate with Claude Desktop/Code
- All 7 tools, 4 resources, and 3 prompts implemented with placeholders
- Real MCP protocol communication working
- Ready for Phase 2: Real Receptor integration

## Quick Start

### Building and Using the MCP Server

1. **Build the Server**:
```bash
# Build the MCP server binary
go build -o bin/receptor-mcp-server ./cmd/receptor-mcp-server/
```

2. **Test Basic Functionality**:
```bash
# Check version and help
./bin/receptor-mcp-server --version
./bin/receptor-mcp-server --help
```

3. **Configure for Claude Desktop**:
   Edit your Claude Desktop config file to include:
```json
{
  "mcpServers": {
    "receptor": {
      "command": "/full/path/to/receptor-mcp/bin/receptor-mcp-server",
      "args": ["--config", "/full/path/to/receptor-mcp.yaml", "--debug"]
    }
  }
}
```

4. **Available Tools** (with placeholder responses):
   - `submit_work` - Submit work to nodes
   - `get_work_status` - Check work status
   - `list_nodes` - List mesh nodes
   - `get_node_info` - Get node details
   - `get_mesh_status` - Get mesh health
   - `cancel_work` - Cancel work
   - `get_work_results` - Get work results

### Building and Testing

```bash
# Build the MCP server
go build -o bin/receptor-mcp-server ./cmd/receptor-mcp-server/

# Build the configuration generator
go build -o bin/receptor-config-gen ./tools/receptor-config-gen/

# Run unit tests
go test ./pkg/mcp/

# Test basic server functionality
./bin/receptor-mcp-server --version
```

### Claude Desktop Integration

The `claude_desktop_config.json` file contains the template configuration.

## Project Structure

```
receptor-mcp/
├── cmd/
│   └── receptor-mcp-server/   # Main MCP server application
│       └── main.go
├── pkg/
│   └── mcp/                   # MCP protocol implementation
│       ├── server.go          # MCP server implementation
│       ├── server_test.go     # Unit tests
│       └── types.go           # MCP protocol types
├── configs/                   # Receptor configuration templates  
│   ├── dev/                   # Development environments (4 templates)
│   ├── prod/                  # Production environments (3 templates)
│   └── work-types/            # AI-optimized work definitions (4 types)
├── deploy/                    # Deployment infrastructure
│   ├── Dockerfile.mcp-server  # Docker image for MCP server
│   ├── docker-compose.*.yaml # Docker Compose configurations (3 files)
│   └── setup.sh              # Automated setup script
├── tools/
│   └── receptor-config-gen/   # Configuration generator CLI
│       ├── main.go            # Generator implementation
│       ├── go.mod             # Go module for generator
│       └── go.sum             # Dependencies
├── go.mod                     # Main Go module
├── go.sum                     # Dependencies
├── receptor-mcp.yaml          # MCP server configuration
└── claude_desktop_config.json # Claude Desktop integration template
```

## MCP Server Capabilities

### 7 Tools (AI-Callable Functions)

1. **`submit_work`** - Submit work to Receptor nodes
   - Parameters: `node_id`, `work_type`, `payload`, `params`
   
2. **`get_work_status`** - Check work execution status  
   - Parameters: `work_id`
   
3. **`list_nodes`** - List all nodes in the mesh
   - Parameters: `filter` (optional)
   
4. **`get_node_info`** - Get detailed node information
   - Parameters: `node_id`
   
5. **`get_mesh_status`** - Get overall mesh health
   - No parameters required
   
6. **`cancel_work`** - Cancel running work
   - Parameters: `work_id`
   
7. **`get_work_results`** - Retrieve completed work results
   - Parameters: `work_id`

### 4 Resources (Real-time Data Access)

- `receptor://mesh/topology` - Real-time mesh network topology
- `receptor://nodes/status` - Current status of all nodes
- `receptor://work/queue` - Active and pending work items  
- `receptor://work/history` - Historical work execution data

### 3 Prompts (Guided Workflows)

- `deploy_workflow` - Guide for deploying complex workflows
- `troubleshoot_mesh` - Mesh network troubleshooting assistant
- `optimize_workload` - Workload optimization recommendations

## Configuration Templates and Tools

The project includes comprehensive Receptor configuration templates and a Go-based generator:

### Available Templates
- **Development**: `dev-single-node`, `dev-mesh-controller`, `dev-mesh-worker`
- **Production**: `prod-controller`, `prod-worker`, `prod-edge`
- **Work Types**: AI-optimized work definitions for various use cases

### Configuration Generator
```bash
# Build and use the generator
go build -o bin/receptor-config-gen ./tools/receptor-config-gen/
./bin/receptor-config-gen --list
./bin/receptor-config-gen -template dev-single-node -output dev.yaml
```

## Testing

```bash
# Unit tests
go test ./pkg/mcp/

# Build tests
go build ./cmd/receptor-mcp-server/
go build ./tools/receptor-config-gen/

# Configuration validation
python3 -c "import yaml; yaml.safe_load(open('receptor-mcp.yaml', 'r')); print('✅ Config valid')"
```

## Example Usage with Claude Desktop

Once configured in Claude Desktop, you can use these tools:

### Basic Operations (Placeholder Responses)
- **List Nodes**: Get list of mesh nodes with mock data
- **Mesh Status**: Get overall mesh health information  
- **Node Info**: Get detailed information about specific nodes

### Work Management (Placeholder Responses)
- **Submit Work**: Submit work with mock work_id response
- **Work Status**: Check status of submitted work
- **Work Results**: Retrieve completed work results
- **Cancel Work**: Cancel running work

### Resources (Mock Data)
- `receptor://mesh/topology` - Mock mesh topology
- `receptor://nodes/status` - Mock node status data
- `receptor://work/queue` - Mock work queue information
- `receptor://work/history` - Mock historical data

## Current Status and Next Steps

### ✅ Current Status: Foundation Complete
- Complete MCP server with placeholder implementations
- All 7 tools, 4 resources, 3 prompts implemented
- Configuration system and deployment infrastructure
- Ready for integration with actual Receptor instances

### 🚧 Next Steps: Real Receptor Integration
- Replace placeholder responses with actual Receptor API calls
- Implement Unix socket communication with Receptor control service
- Add real-time data from live Receptor mesh networks
- Production deployment and monitoring capabilities

## Development

### Prerequisites
- Go 1.21+
- Claude Code (for MCP integration)
- Docker (for deployment testing)

### Building
```bash
# Build MCP server
go build -o bin/receptor-mcp-server ./cmd/receptor-mcp-server/

# Build configuration generator  
go build -o bin/receptor-config-gen ./tools/receptor-config-gen/
```

## Important Notes

- **This is foundation infrastructure** - The MCP server provides placeholder responses
- **Real Receptor integration not yet implemented** - Tools return mock data
- **Ready for development** - All MCP protocol features working correctly
- **Configuration system complete** - Templates and deployment ready

## Next Development Phase

To complete the Receptor integration:
1. Implement Receptor Unix socket client in `pkg/mcp/server.go`
2. Replace placeholder handlers with real Receptor API calls
3. Add error handling for Receptor connection failures
4. Test with actual Receptor mesh instances

---

**Status**: MCP Foundation Complete ✅ - Ready for Receptor Integration 🚀