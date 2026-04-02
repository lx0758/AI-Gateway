## Why

The current login page uses a typical AI-style purple gradient background that feels generic and overused. The dashboard layout has structural issues where 5 stat cards with `span="4"` exceed the 24-column grid, causing layout problems. This change simplifies the UI to be more professional and clean.

## What Changes

- **Login Page**: Replace the purple gradient with a clean, minimal design using a solid light background
- **Dashboard**: Fix grid layout issues and improve visual hierarchy

## Capabilities

### New Capabilities
- `login-ui-simplification`: Simplified login page with clean, professional styling
- `dashboard-layout-fix`: Fixed dashboard grid layout and improved stat card display

### Modified Capabilities
- None

## Impact

- Frontend: `web/src/views/Login/index.vue` - styling changes
- Frontend: `web/src/views/Dashboard/index.vue` - layout fixes and styling improvements
