# AI Code Generator - Windows Developer Setup Script
# This script sets up a complete development environment for new developers on Windows
# Run this script as Administrator in PowerShell

param(
    [switch]$SkipConfirmation,
    [switch]$Force
)

# Configuration
$GoVersion = "1.21.0"
$NodeVersion = "18"
$ProjectName = "AI Code Generator"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    Purple = "Magenta"
    Cyan = "Cyan"
    White = "White"
}

function Write-ColoredOutput {
    param(
        [string]$Message,
        [string]$Color = "White",
        [switch]$NoNewline
    )

    $OriginalColor = $Host.UI.RawUI.ForegroundColor
    $Host.UI.RawUI.ForegroundColor = $Color

    if ($NoNewline) {
        Write-Host $Message -NoNewline
    } else {
        Write-Host $Message
    }

    $Host.UI.RawUI.ForegroundColor = $OriginalColor
}

function Write-Step {
    param([string]$Message)
    Write-ColoredOutput "[STEP] $Message" "Blue"
}

function Write-Success {
    param([string]$Message)
    Write-ColoredOutput "✅ $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColoredOutput "⚠️  $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColoredOutput "❌ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColoredOutput "ℹ️  $Message" "Cyan"
}

function Test-Command {
    param([string]$Command)
    try {
        Get-Command $Command -ErrorAction Stop
        return $true
    } catch {
        return $false
    }
}

function Install-Chocolatey {
    if (!(Test-Command "choco")) {
        Write-Info "Installing Chocolatey..."
        Set-ExecutionPolicy Bypass -Scope Process -Force
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
        Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))
        refreshenv
    } else {
        Write-Success "Chocolatey is already installed"
    }
}

function Install-Go {
    if (!(Test-Command "go")) {
        Write-Info "Installing Go $GoVersion..."

        $GoUrl = "https://go.dev/dl/go$GoVersion.windows-amd64.msi"
        $InstallerPath = "$env:TEMP\go-installer.msi"

        Write-Info "Downloading Go installer..."
        Invoke-WebRequest -Uri $GoUrl -OutFile $InstallerPath

        Write-Info "Installing Go..."
        Start-Process msiexec.exe -Wait -ArgumentList "/i $InstallerPath /quiet /norestart"

        # Add Go to PATH
        $GoPath = "C:\Program Files\Go\bin"
        if (Test-Path $GoPath) {
            $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
            if ($CurrentPath -notlike "*$GoPath*") {
                [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$GoPath", "Machine")
            }
        }

        Remove-Item $InstallerPath -ErrorAction SilentlyContinue
        refreshenv
    }

    # Verify Go installation
    try {
        $GoVersionOutput = & go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            $InstalledVersion = $GoVersionOutput -replace "go version go", "" -replace " windows/amd64", ""
            Write-Success "Go $InstalledVersion is installed"
        } else {
            Write-Warning "Go installation may have issues"
        }
    } catch {
        Write-Error "Go installation failed"
    }
}

function Install-NodeJS {
    if (!(Test-Command "node")) {
        Write-Info "Installing Node.js $NodeVersion..."

        # Install via Chocolatey
        choco install nodejs --version=$NodeVersion -y
        refreshenv
    }

    # Verify Node.js installation
    try {
        $NodeVersionOutput = & node --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Node.js $NodeVersionOutput is installed"
        }
    } catch {
        Write-Error "Node.js installation verification failed"
    }
}

function Install-Docker {
    if (!(Test-Command "docker")) {
        Write-Info "Installing Docker Desktop..."

        $DockerUrl = "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"
        $InstallerPath = "$env:TEMP\docker-installer.exe"

        Write-Info "Downloading Docker Desktop installer..."
        Invoke-WebRequest -Uri $DockerUrl -OutFile $InstallerPath

        Write-Info "Installing Docker Desktop..."
        Write-Warning "Please complete the Docker Desktop installation manually"
        Write-Info "Installer downloaded to: $InstallerPath"
        Start-Process $InstallerPath

        Write-Info "After installation, start Docker Desktop and ensure it's running"
        $null = Read-Host "Press Enter after Docker Desktop is installed and running"
    } else {
        Write-Success "Docker is already installed"
    }
}

function Install-KubernetesTools {
    Write-Info "Installing Kubernetes tools..."

    # Install kubectl
    if (!(Test-Command "kubectl")) {
        Write-Info "Installing kubectl..."
        choco install kubernetes-cli -y
        refreshenv
    } else {
        Write-Success "kubectl is already installed"
    }

    # Install Helm
    if (!(Test-Command "helm")) {
        Write-Info "Installing Helm..."
        choco install kubernetes-helm -y
        refreshenv
    } else {
        Write-Success "Helm is already installed"
    }

    # Install kind
    if (!(Test-Command "kind")) {
        Write-Info "Installing kind..."
        choco install kind -y
        refreshenv
    } else {
        Write-Success "kind is already installed"
    }
}

function Setup-ProjectDependencies {
    Write-Info "Setting up project dependencies..."

    Push-Location $ScriptDir

    # Install Go dependencies
    Write-Info "Installing Go dependencies..."
    & go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to install Go dependencies"
        exit 1
    }

    # Install Node.js dependencies for frontend
    if (Test-Path "web") {
        Write-Info "Installing frontend dependencies..."
        Push-Location "web"
        & npm install
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "Failed to install frontend dependencies"
        }
        Pop-Location
    }

    Pop-Location
    Write-Success "Project dependencies installed"
}

function Create-EnvFile {
    $EnvFile = Join-Path $ScriptDir ".env"

    if (!(Test-Path $EnvFile)) {
        Write-Info "Creating .env file..."

        $EnvContent = @"
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
"@

        $EnvContent | Out-File -FilePath $EnvFile -Encoding UTF8
        Write-Success ".env file created"
    } else {
        Write-Success ".env file already exists"
    }
}

function Setup-Databases {
    Write-Info "Setting up databases..."

    Push-Location $ScriptDir

    if ((Test-Command "docker") -and (Test-Command "docker-compose")) {
        Write-Info "Starting databases with Docker Compose..."
        & docker-compose up -d postgres redis
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Databases started"
        } else {
            Write-Warning "Failed to start databases with Docker Compose"
        }
    } else {
        Write-Warning "Docker/Docker Compose not available. Please start databases manually."
    }

    Pop-Location
}

function Build-Services {
    Write-Info "Building services..."

    Push-Location $ScriptDir

    # Check if Makefile exists and has build target
    if (Test-Path "Makefile") {
        Write-Info "Building services with Make..."
        & make build
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Services built successfully"
        } else {
            Write-Warning "Failed to build services with Make"
        }
    } else {
        Write-Info "Building Go services manually..."
        if (Test-Path "cmd") {
            Get-ChildItem "cmd" -Directory | ForEach-Object {
                $ServiceName = $_.Name
                Write-Info "Building $ServiceName..."
                & go build -o "bin\$ServiceName.exe" ".\cmd\$ServiceName"
            }
        }
        Write-Success "Services built manually"
    }

    Pop-Location
}

function Show-NextSteps {
    Write-Host ""
    Write-ColoredOutput "🚀 $ProjectName Development Environment Setup Complete!" "Magenta"
    Write-Host ""
    Write-ColoredOutput "Available commands:" "White"
    Write-ColoredOutput "  make dev          - Start databases and services" "Cyan"
    Write-ColoredOutput "  make up           - Start all services" "Cyan"
    Write-ColoredOutput "  make build        - Build all services" "Cyan"
    Write-ColoredOutput "  make test         - Run tests" "Cyan"
    Write-ColoredOutput "  make k8s-dev     - Start Kubernetes development environment" "Cyan"
    Write-Host ""
    Write-ColoredOutput "Access your application:" "White"
    Write-ColoredOutput "  Frontend:        http://localhost:3000" "Cyan"
    Write-ColoredOutput "  API Gateway:     http://localhost:8080" "Cyan"
    Write-ColoredOutput "  Adminer (DB):    http://localhost:8090" "Cyan"
    Write-Host ""
    Write-ColoredOutput "Next steps:" "White"
    Write-ColoredOutput "  1. Edit .env file with your configuration" "Cyan"
    Write-ColoredOutput "  2. Run 'make dev' to start the databases" "Cyan"
    Write-ColoredOutput "  3. Run 'make up' to start all services" "Cyan"
    Write-ColoredOutput "  4. Visit http://localhost:3000 to see the application" "Cyan"
    Write-Host ""
}

# Main setup function
function Start-Setup {
    Clear-Host
    Write-Host ""
    Write-ColoredOutput "🚀 Welcome to $ProjectName Windows Development Setup!" "Magenta"
    Write-ColoredOutput "This script will set up your complete development environment." "White"
    Write-Host ""

    # Check if running as administrator
    $IsAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if (!$IsAdmin -and !$Force) {
        Write-Error "Please run this script as Administrator"
        Write-Info "Right-click PowerShell and select 'Run as Administrator'"
        exit 1
    }

    # Confirm setup
    if (!$SkipConfirmation) {
        $Confirmation = Read-Host "Continue with setup? (y/N)"
        if ($Confirmation -notmatch "^[Yy]$") {
            Write-Info "Setup cancelled."
            exit 0
        }
    }

    Write-Host ""

    # Step 1: Install Chocolatey
    Write-Step "Installing package managers..."
    Install-Chocolatey

    # Step 2: Install Go
    Write-Step "Installing Go..."
    Install-Go

    # Step 3: Install Node.js
    Write-Step "Installing Node.js..."
    Install-NodeJS

    # Step 4: Install Docker
    Write-Step "Installing Docker..."
    Install-Docker

    # Step 5: Install Kubernetes tools
    Write-Step "Installing Kubernetes tools..."
    Install-KubernetesTools

    # Step 6: Setup project dependencies
    Write-Step "Setting up project dependencies..."
    Setup-ProjectDependencies

    # Step 7: Create environment file
    Write-Step "Creating environment configuration..."
    Create-EnvFile

    # Step 8: Setup databases
    Write-Step "Setting up databases..."
    Setup-Databases

    # Step 9: Build services
    Write-Step "Building services..."
    Build-Services

    # Step 10: Final setup
    Write-Step "Finalizing setup..."
    Write-Success "Development environment setup complete!"

    # Show next steps
    Show-NextSteps
}

# Run main function
Start-Setup