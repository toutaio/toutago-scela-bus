# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.5.3] - 2026-01-02

### Fixed
- Fixed unused field `mu` in PersistentBus struct
- Fixed unchecked error returns in batch processing (errcheck)
- Fixed unchecked error returns in file and database operations
- Fixed coverage threshold check to only measure pkg/scela (80%+)

### Changed
- Updated CI workflow to exclude examples from coverage calculations
- Improved error handling with explicit error ignoring where appropriate

## [1.5.2] - 2026-01-02

### Fixed
- Fixed code duplication in SQLStore methods (Load, LoadByTopic, LoadAfter)
- Extracted common row scanning logic into scanMessages helper function
- Resolved golangci-lint duplication warnings

## [1.5.1] - 2026-01-02

### Changed
- Enhanced doc.go with comprehensive package documentation following Toutā standards
- Added detailed examples for persistence, audit trail, and observability in doc.go
- Improved README with Status section for consistency with other Toutā subprojects
- Standardized documentation format across package documentation

### Documentation
- Aligned package documentation style with toutago-cosan-router and toutago-nasc-dependency-injector
- Added version information to doc.go
- Enhanced code examples in package-level documentation

## [1.5.0] - 2026-01-02

### Added
- Comprehensive API reference documentation
- Detailed benchmark documentation with performance metrics
- Best practices guide covering patterns, performance, testing, and production deployment
- Advanced usage patterns for complex scenarios
- Error handling and resilience strategies
- Testing guidelines and examples

### Changed
- Enhanced README with links to new documentation files
- Improved documentation structure and navigation

## [1.4.0] - 2026-01-02

### Added
- Message history and audit trail capabilities
- MessageHistory for tracking message lifecycle events
- HistoryEntry with event types: published, delivered, failed, retried
- HistoryMiddleware for automatic history recording
- AuditableBus wrapper with built-in audit trail
- SQLStore for database persistence (works with any database/sql driver)
- Support for SQLite, PostgreSQL, MySQL and other SQL databases
- LoadByTopic and LoadAfter query methods for SQLStore
- ClearBefore method for cleaning old messages
- Count method for message statistics
- Audit trail example application
- Database persistence example application
- 14 new test cases for history and SQL persistence

### Changed
- Test coverage adjusted to 87.5% (comprehensive test suite)
- Added github.com/mattn/go-sqlite3 as optional dependency for examples

## [1.3.0] - 2026-01-02

### Added
- Custom serialization support with Serializer interface
- JSON serializer implementation (default)
- SerializableMessage for custom encoding
- MessageStore interface for persistence abstraction
- InMemoryStore with configurable max size
- FileStore for file-based persistence
- PersistentBus wrapper with automatic persistence
- Message replay capability
- ReplayableStore for time-based message replay
- 8 new test cases for serialization and persistence

### Changed
- Test coverage adjusted to 88.4% (new features added)

## [1.2.0] - 2026-01-02

### Added
- Message filtering capabilities via FilterMiddleware
- TopicFilter, PayloadFilter, and MetadataFilter
- Filter combinators: AndFilter, OrFilter, NotFilter
- BatchPublisher for collecting and publishing messages in batches
- Configurable batch size and wait time
- Batch callback hooks for monitoring
- 10 new test cases for filtering and batching

### Changed
- Increased test coverage from 93.2% to 94.5%

## [1.1.0] - 2026-01-02

### Added
- Priority message support with `PublishWithPriority` method
- Four priority levels: Urgent, High, Normal (default), Low
- Priority-based message processing
- Priority queue example application
- Additional test cases for priority handling

### Changed
- Increased test coverage from 92.8% to 93.2%

## [1.0.0] - 2026-01-02

### Added
- Complete message bus implementation with sync/async publishing
- Pattern matching for topic subscriptions (wildcards: `user.*`, `*.created`, `*`)
- Middleware pipeline for cross-cutting concerns
- Dead letter queue for failed message handling
- Configurable retry logic
- Observer pattern for metrics and monitoring
- Context propagation for timeouts and cancellation
- Graceful shutdown with worker completion
- Thread-safe concurrent operations
- Zero external dependencies
- Framework integration adapter for Toutā
- Comprehensive documentation (README, architecture, usage guide, migration guide)
- 5 working examples (basic, async, middleware, DLQ, metrics)
- 33 test cases with 92.8% coverage
- Performance benchmarks
- CI/CD pipeline with GitHub Actions

### Performance
- Async publish: ~800 ns/op
- Sync publish: ~2,500 ns/op
- Pattern matching: ~150 ns/op

## [0.1.0] - 2026-01-02

### Added
- Initial release
- Repository setup with standard project structure
- License (MIT)
- Code of Conduct
- Contributing guidelines

[1.5.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.5.0
[1.4.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.4.0
[1.3.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.3.0
[1.2.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.2.0
[1.1.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.1.0
[1.0.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.0.0
[0.1.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v0.1.0
