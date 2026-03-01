# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based USSD (Unstructured Supplementary Service Data) framework built with Fiber. The framework provides a flexible architecture for building USSD applications with support for multiple gateways, session management, and menu systems.

## Architecture

The codebase follows a clean architecture pattern with the following key components:

- **Framework Core** (`pkg/ussd/framework.go`): Main orchestrator that manages gateways, sessions, menus, and routing
- **Gateway System** (`pkg/gateway/`): Abstraction layer for different USSD gateways (currently supports Econet)
- **Session Management** (`pkg/session/`): Handles session persistence with multiple storage backends (Redis, Hazelcast, in-memory)
- **Menu System** (`pkg/menu/`): Manages USSD menu navigation and context
- **Router** (`pkg/router/`): Routes USSD requests to appropriate menu handlers

### Key Patterns

- **Registry Pattern**: Used for gateways (`gateway.Registry`) and menus (`menu.Registry`)
- **Strategy Pattern**: Session storage implementations (`session.Repository` interface)
- **Handler Pattern**: Menu processing with `ProcessHandler` interface

## Development Commands

This is a standard Go project. Use these commands for development:

```bash
# Build the application
go build

# Run the application
go run main.go

# Run tests (if any exist)
go test ./...

# Get dependencies
go mod tidy

# Format code
go fmt ./...

# Lint code (if golangci-lint is available)
golangci-lint run
```

## Configuration

The application uses multiple configuration methods:
- YAML configuration file (`config.yaml`) for app settings and menu navigation
- Environment variables loaded via `.env` file (managed by `internal/config/config.go`)
- Viper for advanced configuration management

Key environment variables:
- `SESSION_PROVIDER`: Choose session storage ("redis", "hazelcast", or defaults to "memory")

## Session Management

The framework supports three session storage backends:
- **Memory**: Default, for development/testing
- **Redis**: For production, set `SESSION_PROVIDER=redis`
- **Hazelcast**: For distributed scenarios, set `SESSION_PROVIDER=hazelcast`

Sessions track user selections, pagination state, and custom attributes.

## Menu System

Menus implement the `Menu` interface with:
- `OnRequest()`: Process incoming USSD requests
- `Process()`: Determine navigation type (Continue, Paginated, Replay)

Navigation types:
- `Continue`: Standard menu flow
- `Paginated`: For large option lists
- `Replay`: For going back in menu hierarchy

## Adding New Features

When extending the framework:
1. **New Gateway**: Implement `gateway.Gateway` interface and register with `Registry`
2. **New Menu**: Implement `menu.Menu` interface and add to menu registry
3. **New Session Storage**: Implement `session.Repository` interface
4. **New Middleware**: Add to `pkg/middleware/` and register with framework