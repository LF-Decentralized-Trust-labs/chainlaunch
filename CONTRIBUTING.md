# Contributing to ChainLaunch

Thank you for your interest in contributing to ChainLaunch! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Contributing to ChainLaunch](#contributing-to-chainlaunch)
	- [Table of Contents](#table-of-contents)
	- [Getting Started](#getting-started)
	- [Development Setup](#development-setup)
		- [Prerequisites](#prerequisites)
		- [Local Development](#local-development)
	- [Project Structure](#project-structure)
	- [Code Style](#code-style)
	- [Making Changes](#making-changes)
	- [Adding New Nodes](#adding-new-nodes)
	- [Pull Request Process](#pull-request-process)
	- [Code Review](#code-review)
	- [Questions?](#questions)

## Getting Started

1. Fork the repository
2. Clone your fork:
  ```bash
  git clone https://github.com/YOUR_USERNAME/chainlaunch.git
  cd chainlaunch
  ```
3. Add the original repository as upstream:
  ```bash
  git remote add upstream https://github.com/original/chainlaunch.git
  ```

## Development Setup

### Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- Make (for using Makefile commands)
- SQLC

### Local Development

1. Install dependencies:
  ```bash
  go mod download
  ```

2. Set up the development environment:
  ```bash
  make dev
  ```

3. Run the application:
  ```bash
  make run
  ```

## Project Structure

The project follows a modular architecture with the following structure:

```
.
├── cmd/                    # Application entry points
├── config/                # Configuration files
├── data/                  # Data storage
├── docs/                  # Documentation
├── internal/              # Private application code
│   └── protoutil/        # Protocol utilities
├── pkg/                   # Public libraries
│   ├── api/              # API definitions and interfaces
│   ├── auth/             # Authentication and authorization
│   ├── backups/          # Backup functionality
│   ├── binaries/         # Binary management
│   ├── certutils/        # Certificate utilities
│   ├── common/           # Common utilities
│   ├── config/           # Configuration management
│   ├── db/               # Database operations
│   ├── errors/           # Error handling
│   ├── fabric/           # Fabric network integration
│   ├── http/             # HTTP utilities
│   ├── keymanagement/    # Key management
│   ├── log/              # Logging utilities
│   ├── logger/           # Logger implementation
│   ├── monitoring/       # Monitoring and metrics
│   ├── networks/         # Network management
│   ├── nodes/            # Node management
│   ├── notifications/    # Notification system
│   ├── plugin/           # Plugin system
│   ├── settings/         # Application settings
│   └── version/          # Version management
├── plugins/              # Plugin implementations
├── web/                  # Web interface
└── sqlc.yaml            # SQLC configuration
```

## Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for code formatting
- Run `golangci-lint` before submitting PRs
- Write meaningful commit messages following [Conventional Commits](https://www.conventionalcommits.org/)

## Making Changes

1. Create a new branch:
  ```bash
  git checkout -b feature/your-feature-name
  ```

2. Make your changes following the modular structure:
   - For network-related changes: Work in `pkg/networks/`
   - For API changes: Work in `pkg/api/`
   - For database operations: Work in `pkg/db/`
   - For authentication: Work in `pkg/auth/`
   - For node management: Work in `pkg/nodes/`
   - For key management: Work in `pkg/keymanagement/`
   - For new features: Create a new package in `pkg/` if needed
   - For protocol utilities: Work in `internal/protoutil/`
   - For web interface changes: Work in `web/`
   - For plugin development: Work in `plugins/`

3. Follow these guidelines for each module:
   - Keep related functionality within the same package
   - Use interfaces for better modularity
   - Add proper documentation and comments
   - Follow Go best practices for package organization
   - Use appropriate error handling from `pkg/errors/`

4. Add tests for your changes:
   - Unit tests in the same package
   - Integration tests if needed
   - Update existing tests if modifying existing functionality

5. Update documentation:
   - Add or update package documentation
   - Update relevant README files
   - Add examples if applicable

## Adding New Nodes

When adding support for a new node type to the project, follow these steps:

1. Node Implementation:
   - Create a new package in `pkg/nodes/` for your node type
   - Implement the required interfaces from `pkg/nodes/interfaces.go`
   - Add node-specific configuration in `pkg/config/`
   - Implement node lifecycle management (start, stop, status)

2. Key Files to Modify:
   ```
   pkg/nodes/
   ├── interfaces.go      # Define node interfaces
   ├── factory.go        # Node creation factory
   └── your_node_type/   # Your node implementation
       ├── node.go       # Main node implementation
       ├── config.go     # Node configuration
       └── types.go      # Node-specific types
   ```

3. Required Components:
   - Node configuration struct
   - Node implementation struct
   - Node status monitoring
   - Error handling
   - Logging integration
   - Metrics collection

4. Integration Points:
   - Add node type to the node factory
   - Update configuration validation
   - Add node-specific API endpoints
   - Implement node health checks
   - Add node metrics collection

5. Testing Requirements:
   - Unit tests for node implementation
   - Integration tests with the node type
   - Configuration validation tests
   - Error handling tests
   - Performance benchmarks

6. Documentation:
   - Add node type documentation
   - Update configuration documentation
   - Add usage examples
   - Document any special requirements

7. Security Considerations:
   - Implement proper authentication
   - Handle sensitive data securely
   - Follow security best practices
   - Add security documentation

## Pull Request Process

1. Update the README.md with details of changes if needed
2. Update the documentation if you're changing functionality
3. The PR will be merged once you have the sign-off of at least one other developer
4. Make sure all CI checks pass

## Code Review

- All submissions require review
- Any PR needs to be reviewed by at least one maintainer
- PRs should be small and focused on a single feature or fix
- Respond to review comments promptly

## Questions?

Feel free to open an issue for any questions or concerns you might have.

Thank you for contributing to ChainLaunch!
