# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

## Project

Personal portfolio site for plinehan.com — a Go web app deployed on Google App Engine. It serves a
minimalist image gallery with an OAuth2-based admin interface.

## Commands

```bash
# Development (hot-reload via Air)
air

# Build
go build -o ./tmp/main .

# Run tests
go test ./...

# Run a single test
go test -run TestName ./...

# Deploy to App Engine
./deploy.sh
```

## Architecture

**Entry point:** `main.go` — registers routes, serves embedded templates/images, and applies request
logging middleware.

**Route chain:** The site is a linked navigation sequence: `/` → `/ofcourse` → `/funnyman` →
`/brown` → `/nurbs` → `/thenextlevel` → `/dog`. Each step renders `template.html` with page-specific
colors and images defined in `main.go`.

**Session management:** `sess/sessions.go` — HMAC-SHA256 signed cookies. Only `plinehan@gmail.com`
is authorized. Requires `SESSION_SECRET` env var in production (panics on GAE if unset).

**OAuth2:** `oauth.go` — Google OAuth2 flow. `/sudo/login` initiates it; `/sudo/callback` validates
state, exchanges code for token, verifies email, and sets session cookie.

**Assets:** HTML templates and images under `images/` are embedded into the binary at build time
using `go:embed`.

## Environment Variables

| Variable                     | Required   | Purpose                         |
| ---------------------------- | ---------- | ------------------------------- |
| `GOOGLE_OAUTH_CLIENT_ID`     | Yes        | OAuth2 client ID                |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Yes        | OAuth2 client secret            |
| `SESSION_SECRET`             | Yes (prod) | HMAC key for signing cookies    |
| `PORT`                       | No         | Server port (default 8080)      |
| `GAE_ENV`                    | Auto (GAE) | Triggers production-mode checks |
