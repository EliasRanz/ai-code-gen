# Scripts Directory Structure
# ==========================
#
# This directory contains all automation scripts organized by category.
# Each category has a specific purpose and contains related scripts.
#
# ⚠️  **RECOMMENDED**: Use `make <command>` instead of running scripts directly!
#    Make commands handle dependencies, error checking, and provide better feedback.
#
# Directory Structure:
# ├── database/     - Database setup and migration scripts
# ├── development/  - Development workflow and code generation scripts
# ├── setup/        - Environment setup and configuration scripts
# ├── testing/      - Testing and performance scripts
# └── utilities/    - General utility and helper scripts
#
# Usage (Recommended):
#   make <target>                    - Use Makefile targets (recommended)
#
# Usage (Direct execution):
#   ./scripts/<category>/<script>    - Direct script execution
#   ./scripts/run <category> <script> - Wrapper script execution
#
# Quick Reference:
#   make setup      - Complete environment setup
#   make dev        - Start development environment
#   make test       - Run tests
#   make build      - Build services
#   make generate   - Generate code (mocks, protobufs)

## Setup Scripts
### Linux/macOS/WSL Scripts:
- `setup-dev-db.sh` - Set up development databases
- `setup-dev-environment.sh` - Complete development environment setup
- `setup-local-k8s.sh` - Set up local Kubernetes cluster

### Windows Scripts:
- `setup-dev-environment.ps1` - PowerShell script for complete Windows setup
- `start-dev.bat` - Batch script to start development environment
- `validate-setup.bat` - Batch script to validate setup configuration

### Usage Examples:
```bash
# RECOMMENDED: Use Make commands instead
make setup                    # Complete setup
make dev                      # Start development
make test                     # Run tests

# Direct script execution (Linux/macOS/WSL)
./scripts/setup/setup-dev-environment.sh
./scripts/setup/validate-setup.sh

# Direct script execution (Windows)
.\scripts\setup\setup-dev-environment.ps1
.\scripts\setup\start-dev.bat
.\scripts\setup\validate-setup.bat

# Using wrapper script
./scripts/run setup setup-dev-environment.sh    # Linux/macOS/WSL
./scripts/run setup setup-dev-environment.ps1   # Windows
```

# Windows
.\scripts\setup\setup-dev-environment.ps1
.\scripts\setup\start-dev.bat
.\scripts\setup\validate-setup.bat

# Using wrapper script
./scripts/run setup setup-dev-environment.sh    # Linux/macOS/WSL
./scripts/run setup setup-dev-environment.ps1   # Windows
```