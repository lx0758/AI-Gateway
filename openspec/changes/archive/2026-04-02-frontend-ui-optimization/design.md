## Context

The login page currently uses a purple gradient background that creates an "AI" aesthetic, but feels generic and unprofessional. The gradient (#667eea to #764ba2) is a common pattern associated with AI products.

The dashboard has a layout issue: 5 stat cards use `span="4"` which equals 20, but this causes improper wrapping behavior in Element Plus grid system. Additionally, the charts and tables could benefit from improved visual hierarchy.

## Goals / Non-Goals

**Goals:**
- Simplify login page to use clean, minimal styling
- Fix dashboard grid layout to display all 5 stat cards properly
- Improve visual hierarchy and spacing

**Non-Goals:**
- Do not change functionality (only visual changes)
- Do not add new features
- Do not modify backend or API

## Decisions

### Login Page Design

**Decision**: Replace purple gradient with clean light gray (#f5f5f5) background

**Alternatives considered**:
- Pure white (#fff): Too stark, lacks depth
- Dark mode: Out of scope for this change
- Gradient (subtle): Still looks "AI-ish"

**Rationale**: Light gray (#f5f5f5) provides:
- Professional, minimal appearance
- Good contrast with white login card
- Works well with both light/dark theme

### Dashboard Layout Fix

**Decision**: Change 5 stat cards to use `span="24"` each in a stacked layout or `span="8"` for 3-column layout

**Issue**: 5 cards × span-4 = 20, but element-plus uses 24 columns
- Cards wrap unpredictably or display incorrectly

**Solution**: 
- Use `xs="24" sm="12" md="8"` responsive layout for each card
- This displays 1 card per row on mobile, 2 on tablet, 3 on desktop
- 3 cards per row looks cleaner than 5 cramped cards

### Stat Card Styling

**Decision**: Increase card size and improve typography hierarchy

**Changes**:
- Increase stat-value font size for better readability
- Add subtle shadow on hover
- Improve spacing between elements

## Risks / Trade-offs

- **Risk**: User may prefer the "AI" aesthetic
  - **Mitigation**: The new design is cleaner and more professional
  - **Mitigation**: Both light and dark themes are supported

- **Risk**: Breaking responsive layout on existing screens
  - **Mitigation**: Using Element Plus built-in responsive breakpoints
