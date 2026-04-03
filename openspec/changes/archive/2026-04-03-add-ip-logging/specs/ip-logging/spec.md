## ADDED Requirements

### Requirement: Record client IP address

The system SHALL record the client IP address for each API call in the `usage_logs` table:
- `client_ip`: The trusted original client IP address (string)
- `forwarded_chain`: The complete X-Forwarded-For header value (string)

#### Scenario: Direct connection without proxy

- **WHEN** a client connects directly to the gateway without any proxy
- **THEN** system records `client_ip` as the client's TCP connection IP and `forwarded_chain` as empty

#### Scenario: Single proxy layer (Nginx)

- **WHEN** a request passes through a trusted Nginx proxy with X-Forwarded-For header
- **THEN** system records `client_ip` as the first IP in X-Forwarded-For and `forwarded_chain` as the complete header value

#### Scenario: Multiple proxy layers

- **WHEN** a request passes through multiple trusted proxies with X-Forwarded-For: "client_ip, proxy1, proxy2"
- **THEN** system records `client_ip` as the first IP in chain and `forwarded_chain` as the complete chain

#### Scenario: Untrusted proxy with spoofed IP

- **WHEN** an untrusted proxy sends request with spoofed X-Forwarded-For header
- **THEN** system ignores the spoofed header and records `client_ip` as the proxy's IP, with `forwarded_chain` showing the spoofed value for audit

### Requirement: Configure trusted proxies

The system SHALL support configuration of trusted proxy IP ranges via environment variable `AG_TRUSTED_PROXIES`.

#### Scenario: Default configuration

- **WHEN** `AG_TRUSTED_PROXIES` is not set
- **THEN** system uses default trusted CIDR ranges: "10.0.0.0/8", "192.168.0.0/16", "172.16.0.0/12"

#### Scenario: Custom proxy configuration

- **WHEN** `AG_TRUSTED_PROXIES="nginx_ip,cloudflare_cidr"` is set
- **THEN** system trusts only the specified IP addresses/ranges for parsing X-Forwarded-For headers

#### Scenario: Disable all proxy trust

- **WHEN** `AG_TRUSTED_PROXIES=""` is set to empty string
- **THEN** system trusts no proxies and always uses RemoteAddr as client_ip

### Requirement: Display IP in usage logs

The system SHALL display client IP in the usage logs table with forwarding chain tooltip.

#### Scenario: View IP column in logs table

- **WHEN** user opens the usage logs page
- **THEN** system displays a "客户端IP" column showing `client_ip` value after the "Key" column

#### Scenario: View forwarding chain tooltip

- **WHEN** user hovers over the info icon next to an IP address
- **THEN** system displays tooltip with the complete `forwarded_chain` value

#### Scenario: No forwarding chain

- **WHEN** a log entry has empty `forwarded_chain`
- **THEN** system displays only the IP address without info icon

### Requirement: Aggregate IP statistics

The system SHALL aggregate and display IP usage statistics in a dedicated card.

#### Scenario: View IP statistics card

- **WHEN** user opens the usage page
- **THEN** system displays an "IP 统计" card showing all unique IPs with call count, tokens, and average latency

#### Scenario: IP statistics sorting

- **WHEN** the IP statistics are displayed
- **THEN** IPs are sorted by call count in descending order

#### Scenario: IP with forwarding chain in statistics

- **WHEN** an IP in the statistics table has forwarding chain
- **THEN** system displays the IP with info icon showing the forwarding chain in tooltip