## ADDED Requirements

### Requirement: Consistent API response field naming

The system SHALL use camelCase for JSON API response field names to match frontend expectations.

#### Scenario: Dashboard API returns camelCase fields
- **WHEN** frontend requests /usage/dashboard endpoint
- **THEN** system returns response with camelCase field names (todayRequests, activeProviders, etc.)

#### Scenario: Stats API returns camelCase fields
- **WHEN** frontend requests /usage/stats endpoint
- **THEN** system returns response with camelCase field names (totalRequests, successRate, etc.)
