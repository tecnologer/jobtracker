# Changelog

## [0.0.1] - 2026-07-10

Initial release.

### Added

- Go REST API for tracking job applications: jobs, per-job stage logs, contacts, and configurable stages (default/template stages cloned per job).
- SQLite data layer (pure Go, no CGO) with auto-migration and WAL mode.
- Vue 3 SPA frontend served by the Go binary in production, with Vite dev proxy for local development.
- Meetings and top-match flag for jobs (API + UI).
- CSV export endpoint for jobs.
- HTTP basic auth gating all routes.
- `applied_at` stored as a timezone-aware calendar date, rendered as a wall date regardless of viewer timezone.
- Wails desktop target with a release workflow for desktop binaries (Linux/macOS/Windows).
- Docker Compose dev environment with Go hot reload (air), Kubernetes manifest, and Railway deployment config.
- CI lint workflow (golangci-lint + ESLint with the Vue plugin).
- Default stage seeding with tests.

[0.0.1]: https://github.com/tecnologer/jobtracker/releases/tag/v0.0.1
