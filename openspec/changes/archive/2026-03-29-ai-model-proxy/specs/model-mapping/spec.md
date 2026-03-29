## ADDED Requirements

### Requirement: Create model mapping

The system SHALL allow creating mappings from alias names to provider models.

#### Scenario: Create simple mapping
- **WHEN** admin creates mapping with alias "gpt-4" pointing to provider model "gpt-4-turbo"
- **THEN** system creates mapping record linking alias to provider model

#### Scenario: Multiple mappings for same alias
- **WHEN** admin creates multiple mappings with same alias to different providers
- **THEN** system stores all mappings for load balancing

### Requirement: Route by model alias

The system SHALL route requests based on alias to actual provider model.

#### Scenario: Single mapping routing
- **WHEN** client requests model "gpt-4" with single mapping
- **THEN** system routes to the mapped provider model

#### Scenario: Load balanced routing
- **WHEN** client requests model with multiple mappings
- **THEN** system selects mapping based on weight and provider priority

#### Scenario: Fallback routing
- **WHEN** primary provider is unavailable
- **THEN** system routes to next available mapping for same alias

### Requirement: Set mapping weight

The system SHALL support weight-based load balancing for multiple mappings.

#### Scenario: Weight-based selection
- **WHEN** multiple mappings exist with different weights
- **THEN** system distributes requests proportionally to weights

### Requirement: Enable/disable mapping

The system SHALL allow toggling individual mappings.

#### Scenario: Disable mapping
- **WHEN** admin disables a mapping
- **THEN** system excludes it from routing decisions

### Requirement: List model mappings

The system SHALL provide overview of all model mappings.

#### Scenario: List all mappings
- **WHEN** admin views model mappings page
- **THEN** system shows all mappings with alias, provider, actual model, and status

#### Scenario: Filter by alias
- **WHEN** admin searches for specific alias
- **THEN** system shows all mappings for that alias

### Requirement: Delete model mapping

The system SHALL allow removing model mappings.

#### Scenario: Delete mapping
- **WHEN** admin deletes a mapping
- **THEN** system removes it from database and stops routing to it
