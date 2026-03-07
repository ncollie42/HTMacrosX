# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# HTMacrosX

Macro tracking web app.

## Stack

- **Go** + **Echo v4** — HTTP server, runs on `:8080`, module name `myapp`
- **templ** (`a-h/templ` v0.3.1001) — typed HTML templating, CLI at `~/go/bin/templ`
- **htmx** + **DaisyUI** (dracula theme) — frontend interactivity, embedded via `//go:embed`
- **DB** — SQLite via `modernc.org/sqlite` (pure Go, CGO-free); file at `./app.db` locally, `/data/app.db` on Fly; seeded with test users in `init.go`
- **Auth** — cookie-based sessions stored in an in-memory map (`auth/auth.go`)
- **Deployment** — Fly.io (`fly.toml`)

## Commands

```bash
# Run the server
go run .

# Build
go build .

# Regenerate templ files after editing any .templ file
templ generate

# Install matching templ CLI (must match go.mod version)
go install github.com/a-h/templ/cmd/templ@v0.3.1001
```

## Architecture

### Request Flow

All handlers follow the same pattern:
1. Extract `userID` via `c.Get("userID").(int)` (set by `validate` middleware)
2. Call DB functions to read/write data
3. Build templ components and call `.Render(context.Background(), c.Response().Writer)`

Full-page views compose components via `view.Full(nav, content...)`. Partial htmx responses render a single component directly.

### DB Package (`DB/`)

Package declaration is `database`, imported as `db "myapp/DB"`. Backed by SQLite (`modernc.org/sqlite`, pure Go). `db.Open(path)` is called in `init.go`; path comes from `DB_PATH` env var (defaults to `./app.db`). `SetMaxOpenConns(1)` serializes writes — no mutex needed.

Schema: `users`, `foods`, `meals`, `joins` tables. `DB/db.go` holds `Open()`, the schema string, all shared types, and cross-table query functions (`GetEntriessByDate`, `SumMacros`, `SumMacrosByID`).

Food macros are stored **per gram** (`protein_per_gram`, `fat_per_gram`, etc.) and scaled by `grams` at query time. Calories are computed on-the-fly (protein×4, fat×9, carb×4).

**Meals and Templates share the same `meals` table** — distinguished by `is_preset INTEGER` (0/1).

Seed guard: `init.go` calls `db.CreateUser()` for each test user; UNIQUE constraint silently skips existing users on restart.

### Auth (`auth/auth.go`)

Sessions stored in `map[string]map[string]string` keyed by a random session ID stored in the `sessionID` cookie. `GetUserFromCookie` → `validate` middleware → `c.Set("userID", userID)`.

Duplicate-submission prevention: `auth.GenToken()` / `auth.ValidateDupToken()` / `auth.ClearDupToken()` — used on the templates-to-meal flow.

### Static Assets

htmx, DaisyUI CSS, and html5-qrcode are embedded with `//go:embed` in `main.go` and served at `/htmx`, `/daisy`, `/html5qrcode`.

### Templ / Views (`view/`)

- `.templ` files are the source; `_templ.go` files are generated — never edit `_templ.go` directly
- `view.Full(components ...templ.Component)` wraps everything in the HTML shell with `hx-boost="true"` on `<body>`
- `view.GramEdit(food)` is returned inline on PUT requests so htmx can swap just that card
- Shared UI primitives live in `misc.templ`: `macroViewCompact`, `macroBar`, `GramEdit`, `buttonNav`, `liDelete`, `liEdit`

## Skills & References

When working on htmx interactions, UI components, or adding new routes/views, consult:

- **`.claude/skills/htmx-skill/SKILL.md`** — htmx patterns: core attributes, swap values, polling, inline validation, the "flat procedural function" rule, and when to use `HX-Location` vs `hx-swap`
- **`.claude/skills/htmx-skill/references/daisyui.md`** — DaisyUI 5 component class names and usage (use for all styling work)

## Conventions

- Routes follow REST semantics: `GET`=SELECT, `POST`=INSERT, `PUT/PATCH`=UPDATE, `DELETE`=DELETE
- `validate` middleware reads `userID` from cookie and sets it on the Echo context (`c.Get("userID").(int)`)
- `HX-Location` response header is used for client-side redirects after htmx actions (no full page reload)
- Food search reuses one handler and one template for both `/meal/:id/food_search` and `/template/:id/food_search`
- The `/:date` route on overview accepts a Unix timestamp as a string; empty string defaults to today
