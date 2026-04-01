## ADDED Requirements

### Requirement: Login page

The system SHALL provide a login page for user authentication.

#### Scenario: Display login form
- **WHEN** unauthenticated user accesses the application
- **THEN** system shows login page with username and password fields

#### Scenario: Login success redirect
- **WHEN** user successfully logs in
- **THEN** system redirects to dashboard

### Requirement: Dashboard page

The system SHALL provide a dashboard with overview statistics.

#### Scenario: Display statistics
- **WHEN** user views dashboard
- **THEN** system shows total requests, today's requests, active providers, active keys

#### Scenario: Display charts
- **WHEN** user views dashboard
- **THEN** system shows request trend chart, provider distribution, model usage ranking

### Requirement: Provider management page

The system SHALL provide UI for managing providers.

#### Scenario: List providers
- **WHEN** user navigates to providers page
- **THEN** system shows table of all providers with status

#### Scenario: Create provider form
- **WHEN** user clicks "Add Provider"
- **THEN** system shows form with name, type, base_url, api_key fields

#### Scenario: Provider detail page
- **WHEN** user clicks on a provider
- **THEN** system shows provider details with model list and mappings

### Requirement: Model mapping page

The system SHALL provide UI for managing model mappings.

#### Scenario: List all mappings
- **WHEN** user navigates to models page
- **THEN** system shows all mappings with alias, provider, actual model

#### Scenario: Create mapping
- **WHEN** user creates new mapping
- **THEN** system shows form to select alias and provider model

### Requirement: API key management page

The system SHALL provide UI for managing API keys.

#### Scenario: List API keys
- **WHEN** user navigates to API keys page
- **THEN** system shows all keys with name, masked key, usage, status

#### Scenario: Create key form
- **WHEN** user clicks "Create Key"
- **THEN** system shows form with name, models, quota, rate_limit fields

### Requirement: Usage statistics page

The system SHALL provide UI for viewing usage statistics.

#### Scenario: View statistics
- **WHEN** user navigates to usage page
- **THEN** system shows filters, summary stats, trend chart, and log table

### Requirement: Internationalization

The system SHALL support Chinese and English languages.

#### Scenario: Switch language
- **WHEN** user clicks language toggle
- **THEN** all UI text updates to selected language

#### Scenario: Persist language preference
- **WHEN** user selects a language
- **THEN** system remembers preference for future visits

### Requirement: Dark mode

The system SHALL support light and dark themes.

#### Scenario: Toggle theme
- **WHEN** user clicks theme toggle
- **THEN** UI switches between light and dark mode

#### Scenario: Persist theme preference
- **WHEN** user selects a theme
- **THEN** system remembers preference for future visits

### Requirement: Responsive layout

The system SHALL provide responsive sidebar navigation.

#### Scenario: Desktop layout
- **WHEN** viewing on desktop
- **THEN** sidebar is visible with full menu

#### Scenario: Mobile layout
- **WHEN** viewing on mobile device
- **THEN** sidebar collapses to hamburger menu
