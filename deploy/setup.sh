#!/bin/bash

# Receptor MCP Server Setup Script
# Automates the setup of Receptor mesh with MCP server integration

set -euo pipefail

# Default values
ENVIRONMENT="dev"
NODES=1
CERT_DIR="./certs"
CONFIG_DIR="./configs"
DEPLOY_DIR="./deploy"
FORCE=false
VERBOSE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [[ "$VERBOSE" == "true" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Help function
show_help() {
    cat << EOF
Receptor MCP Server Setup Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Environment type: dev, mesh, prod (default: dev)
    -n, --nodes NUM         Number of nodes to deploy (default: 1)
    -c, --cert-dir DIR      Certificate directory (default: ./certs)
    -f, --force             Force overwrite existing configurations
    -v, --verbose           Enable verbose output
    -h, --help              Show this help message

ENVIRONMENTS:
    dev     - Single node development setup
    mesh    - Multi-node development mesh
    prod    - Production setup with security

EXAMPLES:
    $0 -e dev                           # Single node development
    $0 -e mesh -n 3                     # 3-node development mesh
    $0 -e prod -n 5 -c /etc/certs       # 5-node production setup

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -n|--nodes)
                NODES="$2"
                shift 2
                ;;
            -c|--cert-dir)
                CERT_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate environment
validate_environment() {
    case $ENVIRONMENT in
        dev|mesh|prod)
            log_info "Environment: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT"
            log_error "Valid environments: dev, mesh, prod"
            exit 1
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check for required tools
    local tools=("docker" "docker-compose")
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            log_error "$tool is required but not installed"
            exit 1
        fi
    done
    
    # Check for Go (needed for config generator)
    if ! command -v "go" &> /dev/null; then
        log_warn "Go is not installed. Configuration generator will not be available."
    fi
    
    log_info "Prerequisites check passed"
}

# Generate TLS certificates for production
generate_certificates() {
    if [[ "$ENVIRONMENT" != "prod" ]]; then
        return 0
    fi
    
    log_info "Generating TLS certificates for production environment..."
    
    if [[ -d "$CERT_DIR" && "$FORCE" != "true" ]]; then
        log_warn "Certificate directory already exists. Use --force to regenerate."
        return 0
    fi
    
    mkdir -p "$CERT_DIR"
    
    # Generate CA private key
    openssl genrsa -out "$CERT_DIR/ca.key" 4096
    
    # Generate CA certificate
    openssl req -new -x509 -key "$CERT_DIR/ca.key" -sha256 -subj "/C=US/ST=CA/O=Receptor/CN=Receptor-CA" -days 3650 -out "$CERT_DIR/ca.crt"
    
    # Generate server private key
    openssl genrsa -out "$CERT_DIR/server.key" 4096
    
    # Generate server certificate request
    openssl req -subj "/C=US/ST=CA/O=Receptor/CN=prod-controller" -sha256 -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr"
    
    # Generate server certificate
    openssl x509 -req -in "$CERT_DIR/server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -out "$CERT_DIR/server.crt" -days 365 -sha256
    
    # Generate client private key
    openssl genrsa -out "$CERT_DIR/client.key" 4096
    
    # Generate client certificate request
    openssl req -subj "/C=US/ST=CA/O=Receptor/CN=client" -sha256 -new -key "$CERT_DIR/client.key" -out "$CERT_DIR/client.csr"
    
    # Generate client certificate
    openssl x509 -req -in "$CERT_DIR/client.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -out "$CERT_DIR/client.crt" -days 365 -sha256
    
    # Generate MCP server certificates
    openssl genrsa -out "$CERT_DIR/mcp-server.key" 4096
    openssl req -subj "/C=US/ST=CA/O=Receptor/CN=mcp-server" -sha256 -new -key "$CERT_DIR/mcp-server.key" -out "$CERT_DIR/mcp-server.csr"
    openssl x509 -req -in "$CERT_DIR/mcp-server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -out "$CERT_DIR/mcp-server.crt" -days 365 -sha256
    
    # Clean up CSR files
    rm "$CERT_DIR"/*.csr
    
    # Set proper permissions
    chmod 600 "$CERT_DIR"/*.key
    chmod 644 "$CERT_DIR"/*.crt
    
    log_info "TLS certificates generated successfully"
}

# Generate work signing keys for production
generate_work_keys() {
    if [[ "$ENVIRONMENT" != "prod" ]]; then
        return 0
    fi
    
    log_info "Generating work signing keys..."
    
    local key_dir="./keys"
    mkdir -p "$key_dir"
    
    if [[ -f "$key_dir/work-signing.key" && "$FORCE" != "true" ]]; then
        log_warn "Work signing keys already exist. Use --force to regenerate."
        return 0
    fi
    
    # Generate RSA key pair for work signing
    openssl genrsa -out "$key_dir/work-signing.key" 2048
    openssl rsa -in "$key_dir/work-signing.key" -pubout -out "$key_dir/work-signing.pub"
    
    # Set proper permissions
    chmod 600 "$key_dir/work-signing.key"
    chmod 644 "$key_dir/work-signing.pub"
    
    log_info "Work signing keys generated successfully"
}

# Build configuration generator
build_config_generator() {
    log_info "Building configuration generator..."
    
    if [[ ! -d "tools/receptor-config-gen" ]]; then
        log_warn "Configuration generator source not found. Skipping build."
        return 0
    fi
    
    cd tools/receptor-config-gen
    go build -o ../../bin/receptor-config-gen .
    cd ../..
    
    log_info "Configuration generator built successfully"
}

# Generate configurations
generate_configurations() {
    log_info "Generating Receptor configurations..."
    
    mkdir -p generated-configs
    
    case $ENVIRONMENT in
        dev)
            if [[ -f "bin/receptor-config-gen" ]]; then
                ./bin/receptor-config-gen -template dev-single-node -output generated-configs/dev-single-node.yaml
            else
                cp configs/dev/dev-single-node.yaml generated-configs/
            fi
            ;;
        mesh)
            if [[ -f "bin/receptor-config-gen" ]]; then
                ./bin/receptor-config-gen -template dev-mesh-controller -output generated-configs/controller.yaml
                ./bin/receptor-config-gen -template dev-mesh-worker -param WorkerID=worker1 -param ControllerAddress=controller -output generated-configs/worker1.yaml
                ./bin/receptor-config-gen -template dev-mesh-edge -param EdgeID=edge1 -param ControllerAddress=controller -output generated-configs/edge1.yaml
            else
                cp configs/dev/dev-mesh-*.yaml generated-configs/
            fi
            ;;
        prod)
            if [[ -f "bin/receptor-config-gen" ]]; then
                ./bin/receptor-config-gen -template prod-controller -output generated-configs/prod-controller.yaml
                ./bin/receptor-config-gen -template prod-worker -param ControllerAddress=receptor-controller -nodes "$NODES" -output-dir generated-configs/workers/
            else
                cp configs/prod/prod-*.yaml generated-configs/
            fi
            ;;
    esac
    
    log_info "Configurations generated successfully"
}

# Start services
start_services() {
    log_info "Starting Receptor MCP services..."
    
    local compose_file="$DEPLOY_DIR/docker-compose.$ENVIRONMENT.yaml"
    
    if [[ ! -f "$compose_file" ]]; then
        log_error "Docker Compose file not found: $compose_file"
        exit 1
    fi
    
    # Stop existing services if running
    docker-compose -f "$compose_file" down 2>/dev/null || true
    
    # Start services
    docker-compose -f "$compose_file" up -d
    
    log_info "Services started successfully"
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check service health
    if docker-compose -f "$compose_file" ps | grep -q "unhealthy"; then
        log_warn "Some services may not be healthy. Check with: docker-compose -f $compose_file ps"
    else
        log_info "All services are running"
    fi
}

# Show connection information
show_connection_info() {
    log_info "Receptor MCP Server Setup Complete!"
    echo
    echo "Connection Information:"
    echo "======================"
    
    case $ENVIRONMENT in
        dev)
            echo "MCP Server Access: Connect to localhost:8888"
            echo "Receptor Control Socket: docker exec -it receptor-dev receptorctl status"
            ;;
        mesh)
            echo "MCP Server Access: Connect to localhost:8888"
            echo "Controller: docker exec -it receptor-controller receptorctl status"
            echo "Worker1: docker exec -it receptor-worker1 receptorctl status"
            echo "Edge1: docker exec -it receptor-edge1 receptorctl status"
            ;;
        prod)
            echo "MCP Server Access: Connect to localhost:8888 (with TLS client cert)"
            echo "Controller: docker exec -it receptor-prod-controller receptorctl status"
            echo "Workers: docker-compose -f deploy/docker-compose.prod.yaml ps"
            ;;
    esac
    
    echo
    echo "Useful Commands:"
    echo "==============="
    echo "Check status: docker-compose -f deploy/docker-compose.$ENVIRONMENT.yaml ps"
    echo "View logs: docker-compose -f deploy/docker-compose.$ENVIRONMENT.yaml logs -f"
    echo "Stop services: docker-compose -f deploy/docker-compose.$ENVIRONMENT.yaml down"
    echo
}

# Main function
main() {
    log_info "Starting Receptor MCP Server setup..."
    
    parse_args "$@"
    validate_environment
    check_prerequisites
    
    # Create necessary directories
    mkdir -p bin generated-configs
    
    # Environment-specific setup
    case $ENVIRONMENT in
        prod)
            generate_certificates
            generate_work_keys
            ;;
    esac
    
    build_config_generator
    generate_configurations
    start_services
    show_connection_info
    
    log_info "Setup completed successfully!"
}

# Run main function with all arguments
main "$@"