# History And Back-Button Notes

## Goal

Treat the app more like a native flow:

- Main
- Add something
- Ingredient picker or scan
- Meal view

After the user commits an add, pressing Back from the meal view should return to Main, not back into the transient picker step.

## Core Principle

Only durable resource states should get history entries.

Examples of durable states in this app:

- `/`
- `/meal/:id/`
- `/template/:id/`

Examples of transient task states:

- `/meal/new/food_search`
- `/meal/:id/food_search`
- `/scan?meal=:id`

Transient states are useful UI, but they usually should not remain in browser history after the user completes the task.

## Current Direction

The current implementation already uses a good app-like pattern in one key place:

- Ingredient add responds with `HX-Replace-Url` to replace the transient picker URL with `/meal/:id/`

That means:

- Main -> ingredient picker
- Add ingredient
- URL becomes `/meal/:id/`
- Back goes to Main instead of back to the picker

This is the right model for committed transitions.

## Options

### Option 1: Keep Picker As A Separate Screen, But Treat It As Transient

This is the smallest change from the current architecture.

Approach:

- Open ingredient picker and scan as task screens
- When the user commits an add, replace history with the meal URL
- Avoid pushing new history entries for intermediate task UI when possible

Pros:

- Minimal refactor
- Works well with HTMX
- Matches current route structure

Cons:

- Still exposes picker routes as first-class pages
- History can still feel odd if those pages are entered directly or reloaded

### Option 2: Move Ingredient Picker Into The Meal Screen

This is the cleaner long-term app model.

Approach:

- Make `/meal/new/` or `/meal/:id/` the canonical screen
- Open ingredient search as a modal, drawer, or inline panel
- Keep the user visually inside the meal flow
- After the first ingredient is added, keep the durable URL as the meal URL

Pros:

- Strongest app-like feel
- Cleaner back-button behavior
- Better mental model: picker is a tool inside meal creation, not a destination

Cons:

- Bigger UI refactor
- More HTMX/template restructuring

## Practical Rule

Use this rule when deciding whether something should push history:

- If it represents a thing the user may want to revisit directly, push history
- If it is only a temporary step in completing another task, do not preserve it in history after commit

## Recommendation

Near term:

- Keep the current separate-screen flow
- Continue using `HX-Replace-Url` for successful add/commit transitions
- Treat picker and scan as transient steps

Later:

- Consider refactoring ingredient picker into the meal editor itself

## Follow-Up Work

- Audit all HTMX navigation for `hx-boost`, `HX-Location`, and `HX-Replace-Url`
- Decide which routes are truly durable vs transient
- Add a consistent navigation policy for meal creation, scan, and saved-meal duplication
