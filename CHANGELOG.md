# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.0.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v1.0.0
[0.1.0]: https://github.com/toutaio/toutago-scela-bus/releases/tag/v0.1.0
