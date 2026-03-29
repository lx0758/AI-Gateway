## 1. Project Structure Restructure

- [x] 1.1 Create `server/` directory
- [x] 1.2 Move `cmd/` to `server/cmd/`
- [x] 1.3 Move `internal/` to `server/internal/`
- [x] 1.4 Move `configs/` to `server/configs/`
- [x] 1.5 Update `server/go.mod` module name to `ai-model-proxy`
- [x] 1.6 Update all import paths in Go files
- [x] 1.7 Move `go.mod` and `go.sum` to `server/`
- [x] 1.8 Verify Go build succeeds

## 2. Environment Variable Configuration

- [x] 2.1 Remove YAML config loading from `internal/config/config.go`
- [x] 2.2 Implement environment variable loading with defaults
- [x] 2.3 Add configuration logging at startup (mask sensitive values)
- [x] 2.4 Update `cmd/server/main.go` to use new config system
- [x] 2.5 Remove `configs/config.yaml` dependency
- [x] 2.6 Verify server starts without config file

## 3. API Path Adjustment

- [x] 3.1 Update proxy routes from `/v1/*` to `/openai/v1/*`
- [x] 3.2 Update `internal/middleware/static.go` path detection
- [x] 3.3 Update frontend `vite.config.ts` proxy configuration
- [x] 3.4 Update frontend API client base URL (if needed)
- [x] 3.5 Update README API documentation

## 4. Dependency Updates

- [x] 4.1 Add `github.com/gin-contrib/cors` v1.7.6 to go.mod
- [x] 4.2 Replace custom CORS middleware with gin-contrib/cors
- [x] 4.3 Update `github.com/gin-contrib/sessions` to latest
- [x] 4.4 Run `go mod tidy` to clean dependencies
- [x] 4.5 Remove custom `internal/middleware/cors.go`

## 5. Documentation Updates

- [x] 5.1 Update README with new project structure
- [x] 5.2 Update README with environment variable configuration
- [x] 5.3 Update README with new API paths
- [x] 5.4 Add migration guide for v1 users
- [x] 5.5 Update example curl commands in README

## 6. Verification

- [x] 6.1 Verify Go build succeeds
- [x] 6.2 Verify frontend build succeeds
- [x] 6.3 Verify server starts with default configuration
- [x] 6.4 Verify server starts with environment variables
- [x] 6.5 Verify `/openai/v1/models` endpoint works
- [x] 6.6 Verify `/openai/v1/chat/completions` endpoint works
- [x] 6.7 Verify `/api/v1/auth/login` endpoint works
- [x] 6.8 Verify web UI loads correctly

## 7. Build System (NEW)

- [x] 7.1 Add root `Makefile` with build/dev/clean targets
- [x] 7.2 Add `server/Makefile` with version injection
- [x] 7.3 Add `web/Makefile` with output to `server/res/web/`
- [x] 7.4 Add `server/res/res.go` for embed and version
- [x] 7.5 Add `.vscode/launch.json` debug configurations
- [x] 7.6 Update `cmd/server/main.go` to use embedded static files
- [x] 7.7 Verify build with `make build`
- [x] 7.8 Verify single binary deployment works

## 8. Bug Fixes

- [x] 8.1 Fix i18n locale switch not working (sync i18n.locale on setLocale)
- [x] 8.2 Fix zh.ts Chinese translations (was all English)
- [x] 8.3 Fix Models page hardcoded English labels (alias, weight, model)
- [x] 8.4 Fix Login page locale switch not working
- [x] 8.5 Fix form label-width alignment (use `label-width="auto"`)
- [x] 8.6 Fix Usage page hardcoded English labels (time, model)
- [x] 8.7 Fix Provider Detail page hardcoded English labels
- [x] 8.8 Fix layout: main area padding and height for full-screen content
- [x] 8.9 Add datetime format utility and apply to last_sync_at and usage logs
- [x] 8.10 Change Usage page date picker to datetimerange with i18n labels
- [x] 8.11 Add sidebar version display (format: v2026.3.30)
- [x] 8.12 Fix dark mode not working (add Element Plus dark theme CSS)
- [x] 8.13 Improve Dashboard charts: add loading state, empty state, better styling
- [x] 8.14 Fix Dashboard empty state showing both chart and empty
- [x] 8.15 Add i18n for sidebar title (app.title, app.shortTitle)
- [x] 8.16 Fix model alias not replaced with actual model ID in proxy requests
- [x] 8.17 Fix duplicate `/v1` in API endpoint URL construction
- [x] 8.18 Remove API key encryption (store as plaintext for simplicity)
