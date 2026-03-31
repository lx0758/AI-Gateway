## 1. Project Setup

- [x] 1.1 Initialize Go module with `go mod init github.com/user/ai-model-proxy`
- [x] 1.2 Add core dependencies (gin, gorm, sqlite driver, gin-sessions, bcrypt)
- [x] 1.3 Create project directory structure (cmd, internal, pkg, configs)
- [x] 1.4 Setup configuration loading (YAML config file)
- [x] 1.5 Initialize Vue 3 + TypeScript project with Vite
- [x] 1.6 Add frontend dependencies (element-plus, echarts, vue-echarts, axios, vue-i18n, pinia)

## 2. Database Layer

- [x] 2.1 Define GORM models (User, Provider, ProviderModel, ModelMapping, APIKey, UsageLog)
- [x] 2.2 Setup SQLite connection and auto-migration
- [x] 2.3 Implement database initialization with default admin user
- [x] 2.4 Add model hooks for password hashing (User) and API key encryption (Provider)

## 3. Authentication System

- [x] 3.1 Implement session store (in-memory or Redis)
- [x] 3.2 Create auth middleware for protected routes
- [x] 3.3 Implement login endpoint (`POST /api/v1/auth/login`)
- [x] 3.4 Implement logout endpoint (`POST /api/v1/auth/logout`)
- [x] 3.5 Implement current user endpoint (`GET /api/v1/auth/me`)
- [x] 3.6 Implement password change endpoint (`PUT /api/v1/auth/password`)
- [x] 3.7 Configure secure cookie settings (HttpOnly, Secure, SameSite)

## 4. Provider Management API

- [x] 4.1 Implement list providers endpoint (`GET /api/v1/providers`)
- [x] 4.2 Implement create provider endpoint (`POST /api/v1/providers`)
- [x] 4.3 Implement update provider endpoint (`PUT /api/v1/providers/:id`)
- [x] 4.4 Implement delete provider endpoint (`DELETE /api/v1/providers/:id`)
- [x] 4.5 Implement test provider connection endpoint (`POST /api/v1/providers/:id/test`)

## 5. Model Sync and Management

- [x] 5.1 Implement sync models endpoint (`POST /api/v1/providers/:id/sync`)
- [x] 5.2 Create OpenAI models fetcher (calls GET /v1/models)
- [x] 5.3 Create Anthropic built-in model list
- [x] 5.4 Implement model change detection logic
- [x] 5.5 Implement list provider models endpoint (`GET /api/v1/providers/:id/models`)
- [x] 5.6 Implement manual add model endpoint (`POST /api/v1/providers/:id/models`)
- [x] 5.7 Implement update model endpoint (`PUT /api/v1/providers/:id/models/:mid`)
- [x] 5.8 Implement delete model endpoint (`DELETE /api/v1/providers/:id/models/:mid`)

## 6. Model Mapping API

- [x] 6.1 Implement list mappings endpoint (`GET /api/v1/model-mappings`)
- [x] 6.2 Implement create mapping endpoint (`POST /api/v1/model-mappings`)
- [x] 6.3 Implement update mapping endpoint (`PUT /api/v1/model-mappings/:id`)
- [x] 6.4 Implement delete mapping endpoint (`DELETE /api/v1/model-mappings/:id`)

## 7. API Key Management

- [x] 7.1 Implement API key generation utility (sk- prefix, random string)
- [x] 7.2 Implement list API keys endpoint (`GET /api/v1/api-keys`)
- [x] 7.3 Implement create API key endpoint (`POST /api/v1/api-keys`)
- [x] 7.4 Implement delete API key endpoint (`DELETE /api/v1/api-keys/:id`)
- [x] 7.5 Implement API key validation middleware
- [ ] 7.6 Implement rate limiting middleware (optional)

## 8. Format Transformer

- [x] 8.1 Define Transformer interface (TransformRequest, TransformResponse, TransformStream)
- [x] 8.2 Implement OpenAI to Anthropic request transformer
- [x] 8.3 Implement Anthropic to OpenAI response transformer
- [x] 8.4 Implement Anthropic to OpenAI SSE stream transformer
- [x] 8.5 Implement pass-through transformer for OpenAI-compatible providers
- [x] 8.6 Add transformer factory based on provider type

## 9. Provider Adapters

- [x] 9.1 Define Provider interface (Name, Complete, Stream, NeedsTransform)
- [x] 9.2 Implement HTTP client with connection pooling and timeout
- [x] 9.3 Implement OpenAI provider adapter
- [x] 9.4 Implement Anthropic provider adapter
- [x] 9.5 Implement generic OpenAI-compatible provider adapter

## 10. API Proxy Core

- [x] 10.1 Implement model router (resolve alias to provider and actual model)
- [x] 10.2 Implement fallback/round-robin routing for multiple mappings
- [x] 10.3 Implement chat completions endpoint (`POST /v1/chat/completions`)
- [x] 10.4 Implement models list endpoint (`GET /v1/models`)
- [x] 10.5 Implement model detail endpoint (`GET /v1/models/:id`)
- [x] 10.6 Implement SSE response handling for streaming requests

## 11. Usage Tracking

- [x] 11.1 Implement usage logging middleware
- [x] 11.2 Implement usage stats endpoint (`GET /api/v1/usage/stats`)
- [x] 11.3 Implement usage logs endpoint (`GET /api/v1/usage/logs`)
- [x] 11.4 Implement dashboard metrics aggregation

## 12. Frontend - Setup and Layout

- [x] 12.1 Setup Vue Router with routes and auth guard
- [x] 12.2 Setup Pinia stores (user, app)
- [x] 12.3 Setup vue-i18n with Chinese and English locales
- [x] 12.4 Setup Element Plus with dark mode support
- [x] 12.5 Create MainLayout component (header, sidebar, theme toggle, user menu)

## 13. Frontend - Login Page

- [x] 13.1 Create Login page component
- [x] 13.2 Implement login form with validation
- [x] 13.3 Connect to login API and handle session
- [x] 13.4 Implement redirect after successful login

## 14. Frontend - Dashboard

- [x] 14.1 Create Dashboard page with stat cards
- [x] 14.2 Implement request trend chart (ECharts line chart)
- [x] 14.3 Implement provider distribution chart (ECharts pie chart)
- [x] 14.4 Implement model usage ranking list
- [x] 14.5 Connect to dashboard metrics API

## 15. Frontend - Provider Management

- [x] 15.1 Create Providers list page with table
- [x] 15.2 Create Provider form dialog (add/edit)
- [x] 15.3 Implement test connection functionality
- [x] 15.4 Create Provider detail page with model list
- [x] 15.5 Implement model sync trigger button

## 16. Frontend - Model Mapping

- [x] 16.1 Create Model Mappings list page
- [x] 16.2 Create mapping form dialog with provider/model selection
- [x] 16.3 Implement enable/disable toggle

## 17. Frontend - API Keys

- [x] 17.1 Create API Keys list page
- [x] 17.2 Create API Key form dialog
- [x] 17.3 Implement model selection for allowed_models
- [x] 17.4 Display newly created key with copy button

## 18. Frontend - Usage Statistics

- [x] 18.1 Create Usage page with filters
- [x] 18.2 Implement usage summary cards
- [x] 18.3 Implement usage trend chart
- [x] 18.4 Implement usage logs table with pagination

## 19. Frontend - Settings

- [x] 19.1 Create Settings page
- [x] 19.2 Implement password change form

## 20. Integration and Polish

- [x] 20.1 Build frontend with npm run build
- [x] 20.2 Serve static files from Go server
- [x] 20.3 Add CORS configuration
- [x] 20.4 Add request logging middleware
- [ ] 20.5 Add error handling and user-friendly error messages
- [ ] 20.6 Write README with setup and deployment instructions

## 21. Testing

- [ ] 21.1 Write unit tests for format transformers
- [ ] 21.2 Write unit tests for model router
- [ ] 21.3 Write integration tests for API endpoints
- [ ] 21.4 Write tests for SSE stream handling
