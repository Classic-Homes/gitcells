#!/bin/bash

# GitCells Documentation Server
# Launches the documentation site using Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}===================================${NC}"
    echo -e "${BLUE}  GitCells Documentation Server${NC}"
    echo -e "${BLUE}===================================${NC}"
}

# Check if Docker is installed and running
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Check if Docker Compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not available. Please install Docker Compose."
        exit 1
    fi
}

# Main function
main() {
    print_header
    
    print_status "Checking prerequisites..."
    check_docker
    check_docker_compose
    
    print_status "Starting GitCells documentation server..."
    
    # Use docker-compose or docker compose based on what's available
    if command -v docker-compose &> /dev/null; then
        COMPOSE_CMD="docker-compose"
    else
        COMPOSE_CMD="docker compose"
    fi
    
    # Start the documentation server
    $COMPOSE_CMD up -d docs
    
    # Wait a moment for the server to start
    sleep 3
    
    print_status "Documentation server is starting..."
    print_status "📚 Documentation will be available at: http://localhost:8000"
    print_status "🐳 Container name: gitcells-docs"
    print_status ""
    print_status "Commands:"
    print_status "  View logs:    $COMPOSE_CMD logs -f docs"
    print_status "  Stop server:  $COMPOSE_CMD down"
    print_status "  Restart:      $COMPOSE_CMD restart docs"
    print_status ""
    
    # Try to open the browser (optional)
    if command -v open &> /dev/null; then
        print_status "Opening browser..."
        sleep 2
        open "http://localhost:8000" 2>/dev/null || true
    elif command -v xdg-open &> /dev/null; then
        print_status "Opening browser..."
        sleep 2
        xdg-open "http://localhost:8000" 2>/dev/null || true
    else
        print_warning "Please open http://localhost:8000 in your browser"
    fi
    
    print_status "Press Ctrl+C to view logs, or run '$COMPOSE_CMD down' to stop"
    
    # Follow logs
    $COMPOSE_CMD logs -f docs
}

# Handle script arguments
case "${1:-}" in
    "stop"|"down")
        print_status "Stopping documentation server..."
        if command -v docker-compose &> /dev/null; then
            docker-compose down
        else
            docker compose down
        fi
        print_status "Documentation server stopped."
        ;;
    "logs")
        if command -v docker-compose &> /dev/null; then
            docker-compose logs -f docs
        else
            docker compose logs -f docs
        fi
        ;;
    "restart")
        print_status "Restarting documentation server..."
        if command -v docker-compose &> /dev/null; then
            docker-compose restart docs
        else
            docker compose restart docs
        fi
        print_status "Documentation server restarted."
        ;;
    "help"|"-h"|"--help")
        echo "GitCells Documentation Server"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  (no args)  Start the documentation server"
        echo "  stop       Stop the documentation server"
        echo "  restart    Restart the documentation server"
        echo "  logs       View server logs"
        echo "  help       Show this help message"
        ;;
    *)
        main
        ;;
esac