## ADDED Requirements

### Requirement: Persist login state across page refresh

The system SHALL maintain user login state when the page is refreshed.

#### Scenario: Refresh page while logged in
- **WHEN** user refreshes the page while logged in
- **THEN** system restores user session from backend and stays on current page

#### Scenario: Session expired on refresh
- **WHEN** user refreshes the page with expired session
- **THEN** system redirects to login page

### Requirement: Check session before route guard decision

The system SHALL verify session status before making routing decisions.

#### Scenario: Route guard checks session
- **WHEN** route guard evaluates protected route access
- **THEN** system first attempts to fetch user info from backend before redirecting to login
