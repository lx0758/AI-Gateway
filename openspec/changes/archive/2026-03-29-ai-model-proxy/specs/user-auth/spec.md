## ADDED Requirements

### Requirement: User login

The system SHALL allow users to authenticate with username and password.

#### Scenario: Successful login
- **WHEN** user submits correct username and password
- **THEN** system creates session and returns user info

#### Scenario: Failed login
- **WHEN** user submits incorrect credentials
- **THEN** system returns 401 error with invalid credentials message

#### Scenario: Account disabled
- **WHEN** user's account is disabled
- **THEN** system returns 403 error

### Requirement: Session management

The system SHALL manage user sessions server-side with secure cookies.

#### Scenario: Session cookie
- **WHEN** user successfully logs in
- **THEN** system sets HttpOnly, Secure, SameSite cookie with session ID

#### Scenario: Session validation
- **WHEN** user accesses protected endpoint with valid session cookie
- **THEN** system allows access

#### Scenario: Session expiration
- **WHEN** user's session has expired
- **THEN** system returns 401 error requiring re-login

#### Scenario: Logout
- **WHEN** user logs out
- **THEN** system destroys session and clears cookie

### Requirement: Password management

The system SHALL allow users to change their password.

#### Scenario: Change password
- **WHEN** authenticated user submits old and new password
- **THEN** system validates old password and updates to new one

#### Scenario: Invalid old password
- **WHEN** user submits wrong old password
- **THEN** system returns error

### Requirement: User roles

The system SHALL support at least two user roles.

#### Scenario: Admin role
- **WHEN** user has admin role
- **THEN** user can access all management features

#### Scenario: Viewer role
- **WHEN** user has viewer role
- **THEN** user can only view, not modify settings

### Requirement: Protect admin endpoints

The system SHALL require authentication for all admin API endpoints.

#### Scenario: Unauthenticated access
- **WHEN** request to `/api/v1/*` lacks valid session
- **THEN** system returns 401 error

#### Scenario: Authenticated access
- **WHEN** request includes valid session cookie
- **THEN** system processes the request
