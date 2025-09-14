@echo off
REM AI Code Generator - Windows Setup Validation Script
REM This script validates that the development environment is properly configured

setlocal enabledelayedexpansion

echo.
echo 🔍 AI Code Generator - Development Environment Validation
echo ======================================================
echo.

set "ISSUES=0"

:print_success
echo ✅ %~1
goto :eof

:print_error
echo ❌ %~1
set /a ISSUES+=1
goto :eof

:print_warning
echo ⚠️  %~1
goto :eof

:print_info
echo ℹ️  %~1
goto :eof

REM Check if command exists
:check_command
where "%~1" >nul 2>nul
if %errorlevel% equ 0 (
    call :print_success "%~2 found: %~1"
) else (
    call :print_error "%~2 not found: %~1"
)
goto :eof

REM Check Go version
:check_go_version
where go >nul 2>nul
if %errorlevel% equ 0 (
    for /f "tokens=3" %%i in ('go version') do set GOVERSION=%%i
    set GOVERSION=!GOVERSION:go=!
    echo Go version !GOVERSION! found
    REM Basic version check (simplified)
    call :print_success "Go is installed"
) else (
    call :print_error "Go not found"
)
goto :eof

REM Check Node.js version
:check_nodejs_version
where node >nul 2>nul
if %errorlevel% equ 0 (
    for /f %%i in ('node --version') do set NODEVERSION=%%i
    set NODEVERSION=!NODEVERSION:v=!
    echo Node.js version !NODEVERSION! found
    call :print_success "Node.js is installed"
) else (
    call :print_error "Node.js not found"
)
goto :eof

REM Check Docker
:check_docker
where docker >nul 2>nul
if %errorlevel% equ 0 (
    docker info >nul 2>nul
    if %errorlevel% equ 0 (
        call :print_success "Docker is running"
    ) else (
        call :print_warning "Docker is installed but not running"
    )
) else (
    call :print_error "Docker not found"
)
goto :eof

REM Check Docker Compose
:check_docker_compose
where docker-compose >nul 2>nul
if %errorlevel% equ 0 (
    call :print_success "Docker Compose found"
    goto :eof
)
docker compose version >nul 2>nul
if %errorlevel% equ 0 (
    call :print_success "Docker Compose V2 found"
) else (
    call :print_error "Docker Compose not found"
)
goto :eof

REM Check Kubernetes tools
:check_k8s_tools
call :check_command "kubectl" "kubectl"
call :check_command "helm" "Helm"
call :check_command "kind" "kind"
goto :eof

REM Check project files
:check_project_files
call :print_info "Checking project files..."

if exist "go.mod" (
    call :print_success "go.mod found"
) else (
    call :print_error "go.mod not found"
)

if exist "Makefile" (
    call :print_success "Makefile found"
) else (
    call :print_error "Makefile not found"
)

if exist ".env" (
    call :print_success ".env file found"
) else (
    call :print_warning ".env file not found (copy from .env.example)"
)

if exist "cmd" (
    call :print_success "cmd directory found"
) else (
    call :print_error "cmd directory not found"
)

if exist "internal" (
    call :print_success "internal directory found"
) else (
    call :print_error "internal directory not found"
)

if exist "web" (
    call :print_success "web directory found"
) else (
    call :print_error "web directory not found"
)
goto :eof

REM Check Go dependencies
:check_go_deps
call :print_info "Checking Go dependencies..."
where go >nul 2>nul
if %errorlevel% equ 0 (
    go mod verify >nul 2>nul
    if %errorlevel% equ 0 (
        call :print_success "Go dependencies are valid"
    ) else (
        call :print_warning "Go dependencies may need to be downloaded"
    )
) else (
    call :print_error "Go not available to check dependencies"
)
goto :eof

REM Check Node.js dependencies
:check_nodejs_deps
call :print_info "Checking Node.js dependencies..."
if exist "web" (
    where npm >nul 2>nul
    if %errorlevel% equ 0 (
        if exist "web\node_modules" (
            call :print_success "Node.js dependencies installed"
        ) else (
            call :print_warning "Node.js dependencies not installed (run: cd web && npm install)"
        )
    ) else (
        call :print_error "npm not available to check dependencies"
    )
) else (
    call :print_error "web directory not found"
)
goto :eof

REM Check databases
:check_databases
call :print_info "Checking database connectivity..."
where docker >nul 2>nul
if %errorlevel% equ 0 (
    docker ps 2>nul | findstr postgres >nul
    if %errorlevel% equ 0 (
        call :print_success "PostgreSQL container is running"
    ) else (
        call :print_warning "PostgreSQL container not found (run: make dev)"
    )

    docker ps 2>nul | findstr redis >nul
    if %errorlevel% equ 0 (
        call :print_success "Redis container is running"
    ) else (
        call :print_warning "Redis container not found (run: make dev)"
    )
) else (
    call :print_warning "Docker not available to check database containers"
)
goto :eof

REM Check build
:check_build
call :print_info "Checking if services can be built..."
where go >nul 2>nul
if %errorlevel% equ 0 (
    go build -o NUL .\cmd\api-gateway 2>nul
    if %errorlevel% equ 0 (
        call :print_success "API Gateway can be built"
    ) else (
        call :print_warning "API Gateway build failed (may need dependencies)"
    )
) else (
    call :print_error "Go not available to check build"
)
goto :eof

REM Show next steps
:show_next_steps
echo.
call :print_info "Setup validation complete!"
echo.
call :print_info "If you have issues, try these commands:"
echo   • Run setup script: .\setup-dev-environment.ps1
echo   • Start databases: make dev
echo   • Build services: make build
echo   • Start services: make up
echo   • View logs: make logs
echo.
call :print_info "Access your application at:"
echo   • Frontend: http://localhost:3000
echo   • API Gateway: http://localhost:8080
echo   • Adminer (DB): http://localhost:8090
goto :eof

REM Main validation function
:main
echo.
call :print_info "Starting validation checks..."
echo.

REM Check required tools
echo 📋 Checking Required Tools
echo ---------------------------
call :check_go_version
call :check_nodejs_version
call :check_docker
call :check_docker_compose
echo.

REM Check Kubernetes tools
echo ☸️  Checking Kubernetes Tools (Optional)
echo ----------------------------------------
call :check_k8s_tools
echo.

REM Check project structure
echo 📁 Checking Project Structure
echo -------------------------------
call :check_project_files
echo.

REM Check dependencies
echo 📦 Checking Dependencies
echo ---------------------------
call :check_go_deps
call :check_nodejs_deps
echo.

REM Check databases
echo 🗄️  Checking Databases
echo -----------------------
call :check_databases
echo.

REM Check build
echo 🔨 Checking Build
echo -------------------
call :check_build
echo.

REM Summary
echo 📊 Validation Summary
echo ======================
if %ISSUES% equ 0 (
    call :print_success "All critical components are properly configured!"
    call :print_success "Your development environment is ready."
) else (
    call :print_warning "Found %ISSUES% issue(s) that need attention."
    call :print_info "Run .\setup-dev-environment.ps1 to fix missing components."
)

call :show_next_steps

goto :eof

REM Run main function
call :main