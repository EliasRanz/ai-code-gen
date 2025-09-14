#!/bin/bash

# AI Code Generator - Developer Environment Setup Script
# This script sets up a complete development environment for new developers
# Supports: macOS, Linux, Windows (WSL), and native Windows (via PowerShell)

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Global variables
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="AI Code Generator"
REQUIRED_GO_VERSION="1.21"
REQUIRED_NODE_VERSION="18"
DOCKER_COMPOSE_VERSION="2.0"

# Progress tracking
STEPS_TOTAL=12
STEP_CURRENT=0

# Function to print colored output
print_step() {
    STEP_CURRENT=$((STEP_CURRENT + 1))
    echo -e "${BLUE}[${STEP_CURRENT}/${STEPS_TOTAL}]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

# Function to detect OS
detect_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if [ -f /proc/version ] && grep -qi microsoft /proc/version; then
            echo "wsl"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Homebrew (macOS/Linux)
install_homebrew() {
    if ! command_exists brew; then
        print_info "Installing Homebrew..."
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        if [[ "$(detect_os)" == "linux" ]]; then
            echo 'eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"' >> ~/.bashrc
            eval "$(/home/linuxbrew/.linuxbrew/bin/brew shellenv)"
        fi
    else
        print_success "Homebrew is already installed"
    fi
}

# Function to install Chocolatey (Windows)
install_chocolatey() {
    if ! command_exists choco; then
        print_info "Installing Chocolatey..."
        powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))"
    else
        print_success "Chocolatey is already installed"
    fi
}

# Function to install Go
install_go() {
    if ! command_exists go; then
        print_info "Installing Go ${REQUIRED_GO_VERSION}..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install go@${REQUIRED_GO_VERSION}
                ;;
            "linux")
                wget -q https://go.dev/dl/go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
                export PATH=$PATH:/usr/local/go/bin
                rm go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                ;;
            "wsl")
                wget -q https://go.dev/dl/go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
                export PATH=$PATH:/usr/local/go/bin
                rm go${REQUIRED_GO_VERSION}.linux-amd64.tar.gz
                ;;
            "windows")
                print_info "Please install Go manually from https://go.dev/dl/"
                print_info "Or run: choco install golang"
                exit 1
                ;;
        esac
    fi

    # Verify Go installation
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]]; then
        print_warning "Go version $GO_VERSION is installed, but $REQUIRED_GO_VERSION is recommended"
    else
        print_success "Go $GO_VERSION is installed"
    fi
}

# Function to install Node.js
install_nodejs() {
    if ! command_exists node; then
        print_info "Installing Node.js ${REQUIRED_NODE_VERSION}..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install node@${REQUIRED_NODE_VERSION}
                ;;
            "linux"|"wsl")
                curl -fsSL https://deb.nodesource.com/setup_${REQUIRED_NODE_VERSION}.x | sudo -E bash -
                sudo apt-get install -y nodejs
                ;;
            "windows")
                print_info "Please install Node.js manually from https://nodejs.org/"
                print_info "Or run: choco install nodejs"
                exit 1
                ;;
        esac
    fi

    # Verify Node.js installation
    NODE_VERSION=$(node --version | sed 's/v//')
    print_success "Node.js $NODE_VERSION is installed"

    # Install npm if not included
    if ! command_exists npm; then
        print_info "Installing npm..."
        curl -L https://www.npmjs.com/install.sh | sh
    fi
}

# Function to install Docker
install_docker() {
    if ! command_exists docker; then
        print_info "Installing Docker..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install --cask docker
                print_info "Please start Docker Desktop manually"
                ;;
            "linux")
                curl -fsSL https://get.docker.com -o get-docker.sh
                sudo sh get-docker.sh
                sudo usermod -aG docker $USER
                rm get-docker.sh
                ;;
            "wsl")
                print_info "For WSL, install Docker Desktop on Windows and enable WSL integration"
                print_info "See: https://docs.docker.com/desktop/wsl/"
                ;;
            "windows")
                print_info "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
                print_info "Or run: choco install docker-desktop"
                exit 1
                ;;
        esac
    else
        print_success "Docker is already installed"
    fi
}

# Function to install kubectl
install_kubectl() {
    if ! command_exists kubectl; then
        print_info "Installing kubectl..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install kubectl
                ;;
            "linux"|"wsl")
                curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
                sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
                rm kubectl
                ;;
            "windows")
                print_info "Please install kubectl manually or run: choco install kubernetes-cli"
                exit 1
                ;;
        esac
    fi
    print_success "kubectl $(kubectl version --client --short) is installed"
}

# Function to install Helm
install_helm() {
    if ! command_exists helm; then
        print_info "Installing Helm..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install helm
                ;;
            "linux"|"wsl")
                curl https://get.helm.sh/helm-v3.12.0-linux-amd64.tar.gz -o helm.tar.gz
                tar -zxvf helm.tar.gz
                sudo mv linux-amd64/helm /usr/local/bin/helm
                rm -rf linux-amd64 helm.tar.gz
                ;;
            "windows")
                print_info "Please install Helm manually or run: choco install kubernetes-helm"
                exit 1
                ;;
        esac
    fi
    print_success "Helm $(helm version --short) is installed"
}

# Function to install kind
install_kind() {
    if ! command_exists kind; then
        print_info "Installing kind..."
        OS_TYPE=$(detect_os)

        case $OS_TYPE in
            "macos")
                brew install kind
                ;;
            "linux"|"wsl")
                curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
                chmod +x ./kind
                sudo mv ./kind /usr/local/bin/kind
                ;;
            "windows")
                print_info "Please install kind manually or run: choco install kind"
                exit 1
                ;;
        esac
    fi
    print_success "kind $(kind version) is installed"
}

# Function to setup project dependencies
setup_project() {
    print_info "Setting up project dependencies..."

    # Install Go dependencies
    print_info "Installing Go dependencies..."
    go mod download

    # Install Node.js dependencies for frontend
    if [ -d "web" ]; then
        print_info "Installing frontend dependencies..."
        cd web
        npm install
        cd ..
    fi

    print_success "Project dependencies installed"
}

# Function to create environment file
create_env_file() {
    if [ ! -f ".env" ]; then
        print_info "Creating .env file..."
        cp .env.example .env 2>/dev/null || cat > .env << EOF
# Database Configuration
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ai_ui_generator

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6380
REDIS_PASSWORD=password

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# AI Service Configuration
OPENAI_API_KEY=your-openai-api-key

# Application Configuration
APP_ENV=development
APP_PORT=8080
FRONTEND_URL=http://localhost:3000
EOF
        print_success ".env file created"
    else
        print_success ".env file already exists"
    fi
}

# Function to setup databases
setup_databases() {
    print_info "Setting up databases..."

    # Start databases with Docker Compose
    if command_exists docker && command_exists docker-compose; then
        print_info "Starting databases with Docker Compose..."
        docker-compose up -d postgres redis
        print_success "Databases started"
    else
        print_warning "Docker/Docker Compose not available. Please start databases manually."
    fi
}

# Function to run database migrations
run_migrations() {
    print_info "Running database migrations..."
    # Add migration commands here when ready
    print_info "Migrations completed (placeholder)"
}

# Function to build services
build_services() {
    print_info "Building services..."
    make build
    print_success "Services built successfully"
}

# Function to start development environment
start_dev_environment() {
    print_info "Starting development environment..."

    echo ""
    echo -e "${PURPLE}🚀 ${PROJECT_NAME} Development Environment Setup Complete!${NC}"
    echo ""
    echo -e "${WHITE}Available commands:${NC}"
    echo -e "  ${CYAN}make dev${NC}          - Start databases and services"
    echo -e "  ${CYAN}make up${NC}           - Start all services"
    echo -e "  ${CYAN}make build${NC}        - Build all services"
    echo -e "  ${CYAN}make test${NC}         - Run tests"
    echo -e "  ${CYAN}make k8s-dev${NC}     - Start Kubernetes development environment"
    echo ""
    echo -e "${WHITE}Access your application:${NC}"
    echo -e "  ${CYAN}Frontend:${NC}        http://localhost:3000"
    echo -e "  ${CYAN}API Gateway:${NC}     http://localhost:8080"
    echo -e "  ${CYAN}Adminer (DB):${NC}    http://localhost:8090"
    echo ""
    echo -e "${WHITE}Next steps:${NC}"
    echo -e "  1. Edit ${CYAN}.env${NC} file with your configuration"
    echo -e "  2. Run ${CYAN}make dev${NC} to start the databases"
    echo -e "  3. Run ${CYAN}make up${NC} to start all services"
    echo -e "  4. Visit http://localhost:3000 to see the application"
    echo ""
}

# Main setup function
main() {
    echo ""
    echo -e "${PURPLE}🚀 Welcome to ${PROJECT_NAME} Development Setup!${NC}"
    echo -e "${WHITE}This script will set up your complete development environment.${NC}"
    echo ""

    # Detect OS
    OS_TYPE=$(detect_os)
    print_info "Detected OS: $OS_TYPE"

    # Confirm setup
    read -p "Continue with setup? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Setup cancelled."
        exit 0
    fi

    echo ""

    # Step 1: Install package managers
    print_step "Installing package managers..."
    case $OS_TYPE in
        "macos"|"linux")
            install_homebrew
            ;;
        "windows")
            install_chocolatey
            ;;
        "wsl")
            install_homebrew
            ;;
    esac

    # Step 2: Install Go
    print_step "Installing Go..."
    install_go

    # Step 3: Install Node.js
    print_step "Installing Node.js..."
    install_nodejs

    # Step 4: Install Docker
    print_step "Installing Docker..."
    install_docker

    # Step 5: Install Kubernetes tools
    print_step "Installing Kubernetes tools..."
    install_kubectl
    install_helm
    install_kind

    # Step 6: Setup project dependencies
    print_step "Setting up project dependencies..."
    setup_project

    # Step 7: Create environment file
    print_step "Creating environment configuration..."
    create_env_file

    # Step 8: Setup databases
    print_step "Setting up databases..."
    setup_databases

    # Step 9: Run migrations
    print_step "Running database migrations..."
    run_migrations

    # Step 10: Build services
    print_step "Building services..."
    build_services

    # Step 11: Final setup
    print_step "Finalizing setup..."
    print_success "Development environment setup complete!"

    # Step 12: Display next steps
    start_dev_environment
}

# Run main function
main "$@"