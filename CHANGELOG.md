# Changelog

All notable changes to the Cloud Climbers Slack Bot project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.6.0] - 2026-02-08

### Security
- **CRITICAL**: Fixed 19 security vulnerabilities in Go standard library and dependencies
- Upgraded Go toolchain from 1.22 to 1.24.13 to address CVEs in:
  - `crypto/tls` - TLS handshake encryption level issues and unexpected session resumption
  - `crypto/x509` - Certificate validation vulnerabilities, DNS name constraint bypass
  - `net/http` - Memory exhaustion in cookie parsing, request smuggling, sensitive header leaks
  - `net/url` - Memory exhaustion in query parameter parsing
  - `encoding/asn1` - DER payload parsing memory exhaustion
- Fixed sensitive data leak vulnerability in `github.com/go-viper/mapstructure/v2`
- **Docker**: Updated all base images to latest secure versions
  - Go images: 1.18/1.22 → 1.24.13
  - Python images: 3.9 (EOL) → 3.12
  - Alpine: Updated to 3.21
- **Docker**: Implemented non-root user execution in all containers
- **Docker**: Added CA certificates for HTTPS support in scratch-based images
- Pinned Python dependencies to specific versions to prevent supply chain attacks

### Changed
- **Breaking**: Upgraded major dependencies (requires Go 1.24.13+):
  - `github.com/slack-go/slack`: v0.10.0 → v0.15.0
  - `github.com/spf13/viper`: v1.10.1 → v1.20.0
  - `go.mongodb.org/mongo-driver`: v1.11.2 → v1.17.3
  - `go.uber.org/zap`: v1.19.1 → v1.27.0
  - `github.com/go-viper/mapstructure/v2`: v2.2.1 → v2.4.0
- Updated 15+ indirect dependencies to latest secure versions
- Modernized logging and configuration libraries
- **Docker**: Optimized build process with multi-stage builds
- **Docker**: Added binary stripping (-ldflags="-w -s") to reduce image size
- **CI/CD**: Pinned GitHub Actions to specific Go version (1.24.13)
- **CI/CD**: Added automated vulnerability scanning in CI pipeline
- **CI/CD**: Added Docker Buildx for improved build performance
- **CI/CD**: Separated build/push logic for PRs vs main branch

### Fixed
- Fixed `slack.NewInputBlock()` API compatibility for Slack SDK v0.15.0 (added hint parameter)
- Fixed `GetConversationInfo()` to use new struct-based API (`GetConversationInfoInput`)
- Resolved all compilation warnings and deprecation notices

### Added
- Created `CHANGELOG.md` to track version history going forward
- Added `requirements.txt` files for all Python plugins
- Added `.dockerignore` file to improve build performance and security
- **CI/CD**: Added vulnerability scanning with govulncheck in GitHub Actions
- **CI/CD**: Added PR build validation (build-only without push)

### Notes
- This release contains no functional changes to bot behavior
- All changes are security updates and dependency modernization
- Thoroughly tested with `govulncheck` - zero vulnerabilities reported
- Docker images are significantly smaller and more secure
- Recommended for immediate deployment to production

---

## [2.5.5.4] - 2024-06-04

### Changed
- Updated release configuration in Makefile

## [2.5.5.3 and earlier] - 2024-06-01 to 2024-06-24

The period between v2.5.5.3 and v2.5.5.4 included 40+ commits focused on infrastructure improvements:

### Added
- Nginx load balancer configuration
- Multiple WordPress preview environments support
- Domain configuration for preview environments
- Ingress Nginx integration for WordPress deployments

### Changed
- Enhanced Flux GitOps configuration
- Improved preview environment management
- Infrastructure configuration refinements

### Fixed
- Multiple bug fixes and configuration adjustments

### Notes
- Previous releases focused on Flux-based preview environment creation
- Core bot functionality for Slack-based environment management

---

## Version History Summary

- **v2.6.0** (2026-02-08): Security hardening - Go 1.24.13 upgrade and dependency updates
- **v2.5.5.4** (2024-06-04): Release configuration update
- **v2.5.5.3** (2024-06-03): Infrastructure improvements
- **v2.5.5.2** and **v2.5.5.1**: Previous iterations

For detailed commit history, see: https://github.com/diamonce/cloudclimbers/commits/main

---

## Migration Guide

### Upgrading from v2.5.5.x to v2.6.0

**Prerequisites:**
- Go 1.24.13 or later must be installed on build/deployment systems
- Update CI/CD pipelines to use Go 1.24.13

**Steps:**
1. Pull latest changes from main branch
2. Rebuild all Docker images: `make docker-build`
3. Update Kubernetes deployments with new images
4. No configuration changes required - fully backward compatible
5. Verify deployment with health checks

**Rollback:**
- If issues occur, rollback is straightforward - no database migrations or config changes
- Previous v2.5.5.4 images remain functional
