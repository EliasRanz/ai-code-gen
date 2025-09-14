@echo off
REM AI Code Generator - Local Development Environment Script (Windows)
REM This script provides a single command to start the entire local development environment

setlocal enabledelayedexpansion

REM Colors for Windows CMD (limited support)
set "GREEN=[SUCCESS]"
set "YELLOW=[WARNING]"
set "RED=[ERROR]"
set "BLUE=[INFO]"

REM Project name
set "PROJECT_NAME=AI Code Generator"
set "COMPOSE_FILE=docker-compose.yml"

REM Function to print status
:print_status
echo %BLUE% %~1
goto :eof

:print_success
echo %GREEN% %~1
goto :eof

:print_warning
echo %YELLOW% %~1
goto :eof

:print_error
echo %RED% %~1
goto :eof

REM Check if Docker is running
:check_docker
docker info >nul 2>&1
if errorlevel 1 (
    call :print_error "Docker is not running. Please start Docker and try again."
    exit /b 1
)
call :print_success "Docker is running"
goto :eof

REM Check if docker-compose file exists
:check_compose_file
if not exist "%COMPOSE_FILE%" (
    call :print_error "docker-compose.yml not found in current directory"
    exit /b 1
)
call :print_success "Found docker-compose.yml"
goto :eof

REM Start services
:start_services
call :print_status "Starting all services..."
docker-compose up -d
if errorlevel 1 (
    call :print_error "Failed to start services"
    exit /b 1
)
call :print_success "Services started successfully"
goto :eof

REM Wait for services to be healthy
:wait_for_services
call :print_status "Waiting for services to be ready..."
REM Note: Windows CMD has limited timeout capabilities, so we'll do a simple wait
timeout /t 10 /nobreak >nul
call :print_success "Services should be ready now"
goto :eof

REM Show service status
:show_status
echo.
call :print_success "=== %PROJECT_NAME% Development Environment ==="
echo.
echo Services Status:
docker-compose ps
echo.
call :print_success "Access URLs:"
echo   Frontend:        http://localhost:3000
echo   API Gateway:     http://localhost:8080
echo   Auth Service:    http://localhost:8081
echo   User Service:    http://localhost:8082
echo   AI Service:      http://localhost:8083
echo   AI Generation:   http://localhost:8084
echo   Database Admin:  http://localhost:8090
echo   vLLM AI Server:  http://localhost:8000
echo.
call :print_success "Database Connections:"
echo   PostgreSQL: localhost:5433 (user: postgres, db: ai_ui_generator)
echo   Redis:      localhost:6380
echo.
call :print_status "Useful Commands:"
echo   View logs:       docker-compose logs -f
echo   Stop services:   docker-compose down
echo   Restart service: docker-compose restart ^<service-name^>
echo   Service status:  docker-compose ps
goto :eof

REM Check if services are already running
:check_running_services
docker-compose ps | findstr "Up" >nul
if not errorlevel 1 (
    call :print_warning "Some services are already running"
    echo Current status:
    docker-compose ps
    echo.
    set /p "choice=Do you want to restart all services? (y/N): "
    if /i "!choice!"=="y" (
        call :print_status "Restarting services..."
        docker-compose down >nul 2>&1
    ) else (
        call :print_status "Using existing services..."
        call :show_status
        exit /b 0
    )
)
goto :eof

REM Main function
:main
echo.
call :print_success "Starting %PROJECT_NAME% Local Development Environment"
echo ==================================================
echo.

REM Pre-flight checks
call :check_docker
call :check_compose_file

REM Check if services are already running
call :check_running_services

REM Start services
call :start_services

REM Wait for services to be ready
call :wait_for_services

REM Show final status
call :show_status

echo.
call :print_success "Development environment is ready!"
call :print_status "Happy coding!"
echo.
goto :eof

REM Run main function
call :main %*