# Receptor MCP Server

A Model Context Protocol (MCP) server that exposes Ansible Receptor's mesh networking and work execution capabilities to AI applications like Claude Desktop and Claude Code.

## ✅ Phase 1 Complete: MCP Server Implementation

**Status**: Fully functional MCP server with Claude Code integration!

### What's Working

- ✅ **Full MCP Protocol Support** - JSON-RPC 2.0 over stdio
- ✅ **All 7 Receptor Tools** - Available via slash commands in Claude Code
- ✅ **All 4 Resources** - Real-time data access via @ mentions
- ✅ **All 3 Prompts** - Guided workflow assistance
- ✅ **Claude Code Integration** - Native MCP support
- ✅ **Configuration Management** - YAML-based server configuration
- ✅ **Comprehensive Testing** - Unit tests and integration validation

### Current Implementation Status

**Phase 1: Core MCP Server** ✅ **COMPLETE**
- Full MCP server with placeholder implementations
- Perfect Claude Code integration via MCP protocol
- All tools discoverable and executable
- Ready for Phase 2 real Receptor integration

## Quick Start

### Claude Code Integration

1. **Add the MCP Server**:
```bash
claude mcp add receptor /path/to/receptor-mcp/bin/receptor-mcp-server --debug
```

2. **Verify Integration**:
```bash
claude mcp list
# Should show: receptor: ... - ✓ Connected
```

3. **Use Receptor Tools**:
   - `/mcp__receptor__list_nodes` - List mesh nodes
   - `/mcp__receptor__submit_work` - Submit work to nodes
   - `/mcp__receptor__get_mesh_status` - Get mesh health
   - And 4 more tools...

4. **Access Resources**:
   - `@receptor:receptor://mesh/topology` - Mesh topology data
   - `@receptor:receptor://nodes/status` - Node status info
   - `@receptor:receptor://work/queue` - Work queue status
   - `@receptor:receptor://work/history` - Historical data

### Building and Testing

```bash
# Build the MCP server
go build -o bin/receptor-mcp-server ./cmd/receptor-mcp-server/

# Run tests
go test ./pkg/mcp/

# Test MCP handshake
./test-mcp-handshake.sh
```

### Claude Desktop Integration

Copy contents of `claude_desktop_config.json` to your Claude Desktop config:

```json
{
  "mcpServers": {
    "receptor": {
      "command": "/path/to/receptor-mcp/bin/receptor-mcp-server",
      "args": ["--config", "/path/to/receptor-mcp.yaml", "--debug"]
    }
  }
}
```

## Project Structure

```
receptor-mcp/
├── cmd/
│   └── receptor-mcp-server/   # Main MCP server application
├── pkg/
│   └── mcp/                   # MCP protocol implementation
├── configs/                   # Receptor configuration templates  
│   ├── dev/                   # Development environments
│   ├── prod/                  # Production environments
│   └── work-types/            # AI-optimized work definitions
├── deploy/                    # Deployment files
│   ├── docker-compose.*.yaml # Docker Compose configurations
│   └── setup.sh              # Automated setup script
├── tools/
│   └── receptor-config-gen/   # Configuration generator CLI
├── bin/                       # Built binaries (gitignored)
├── receptor-mcp.yaml          # MCP server configuration
├── claude_desktop_config.json # Claude Desktop integration template
└── test-mcp-handshake.sh     # MCP integration test script
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
# Run all tests
./test-all.sh

# Unit tests only
go test ./pkg/mcp/

# Test MCP handshake
./test-mcp-handshake.sh
```

## Example Workflows in Claude Code

### Basic Mesh Operations
```
/mcp__receptor__list_nodes
/mcp__receptor__get_mesh_status
@receptor:receptor://mesh/topology
```

### Work Submission and Monitoring
```
/mcp__receptor__submit_work node_id=worker-01 work_type=ai-script payload="print('Hello Receptor!')"
/mcp__receptor__get_work_status work_id=work_123456
/mcp__receptor__get_work_results work_id=work_123456
```

### Guided Workflows
```
/mcp__receptor__deploy_workflow workflow_type=ai-pipeline
/mcp__receptor__troubleshoot_mesh issue_type=connectivity
/mcp__receptor__optimize_workload
```

## Current Status and Next Steps

### Phase 1: ✅ Complete
- Full MCP server implementation
- Claude Code integration working
- All tools, resources, and prompts functional
- Placeholder responses for all operations

### Phase 2: 🚧 Next Steps  
- Real Receptor integration via Unix sockets
- Live data from actual Receptor instances
- Dynamic resource updates
- Production deployment capabilities

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

## Management Commands

```bash
# MCP server management
claude mcp list                    # List all MCP servers
claude mcp get receptor            # Get receptor server details
claude mcp remove receptor -s local # Remove the server

# Test server directly
./bin/receptor-mcp-server --version
./bin/receptor-mcp-server --help
```

---

**Status**: Phase 1 Complete ✅ - Ready for Phase 2 Implementation 🚀