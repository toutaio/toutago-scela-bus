# Contributing to Scéla

Thank you for your interest in contributing to Scéla! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker
- Check if the issue already exists
- Provide detailed information:
  - Go version
  - Operating system
  - Steps to reproduce
  - Expected vs actual behavior

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write or update tests
5. Ensure tests pass (`go test ./...`)
6. Ensure code is formatted (`go fmt ./...`)
7. Run linter (`golangci-lint run`)
8. Commit with conventional commit format
9. Push to your fork
10. Open a Pull Request

### Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement
- `refactor`: Code restructuring
- `test`: Test additions or modifications
- `docs`: Documentation changes
- `chore`: Build, CI, or tooling changes

**Examples:**
```
feat(bus): add support for priority queues
fix(subscriber): handle concurrent unsubscribe correctly
perf(routing): optimize pattern matching with trie
docs(readme): add quick start guide
test(middleware): add pipeline execution tests
```

## Development Setup

### Prerequisites

- Go 1.22.x or higher
- Git
- golangci-lint (for linting)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/toutago-scela-bus
cd toutago-scela-bus

# Install dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linter
golangci-lint run
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Keep functions focused and testable
- Write clear, descriptive names
- Add comments for non-obvious logic
- Follow SOLID principles

## Testing

- Write tests for all new features
- Maintain 80%+ code coverage
- Include table-driven tests where appropriate
- Test edge cases and error conditions
- Use race detector (`go test -race`)

## Documentation

- Update README.md for user-facing changes
- Add GoDoc comments for all exported symbols
- Include usage examples
- Document breaking changes in CHANGELOG.md

## Questions?

Feel free to open an issue for questions or discussions.
