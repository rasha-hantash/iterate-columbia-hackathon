# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Commodity price alert microservice — a Go backend with PostgreSQL for managing price alerts on agricultural commodities (CORN, WHEAT, SOYBEAN_OIL). Multi-tenant via client-based isolation; authentication via `X-User-ID` header.

## Commands

```bash
# Database setup (PostgreSQL must be running locally)
psql -h localhost -U edge -d edge_interview < db/init.sql

# Run the server (listens on :8000)
cd platform && go run .

# Build
cd platform && go build -o edge-alerts .

# Run tests (from platform/)
cd platform && go test ./...

# Run a single test
cd platform && go test -run TestName ./...

# Run LLM evals (requires ANTHROPIC_API_KEY)
cd eval && uv sync
ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --verbose

# Run only the analyzer (skip judging)
cd eval && ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --skip-judge -v

# Run only the judge on existing responses
cd eval && ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --skip-analysis -v

# Test a new analyzer prompt version
cd eval && ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --analyzer-prompt v2 -v
```

## Architecture

All Go source lives in `platform/` as a single `main` package (no sub-packages).

**Three-layer structure:**
- `main.go` — HTTP server setup, routing (`net/http` stdlib), auth middleware (`getCurrentUser` looks up user by `X-User-ID` header)
- `handler.go` — `AlertHandler` with request parsing, validation, JSON responses. Shared helpers: `respondJSON()`, `respondError()`
- `service.go` — `AlertService` with business logic and direct SQL via `database/sql`. Uses transactions for create/trigger operations. Writes to `alert_history` for audit trail

**Database** (`db/init.sql`): Schema + seed data in one file. Key tables: `clients`, `users`, `commodities`, `positions`, `price_data`, `price_alerts`, `alert_history`. Soft deletes via `deleted_at` on `price_alerts`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check (pings DB) |
| POST | `/alerts` | Create a price alert |
| GET | `/alerts?status=&commodity_code=` | List alerts for client |
| POST | `/alerts/{id}/trigger` | Trigger an alert |

## Eval Framework

Offline LLM-as-judge evaluation for the AI position analysis feature. Lives in `eval/`. See `eval/EVAL.md` for full documentation.

**Two-phase pipeline:**
1. **Analyzer** — Calls Claude with positions + `create_alert` tool, captures alert suggestions
2. **Judge** — Separate Claude call evaluates suggestions against golden dataset ground truth

**Golden dataset** (`eval/golden/scenarios.csv`): 10 CORN scenarios — 5 strict (exact price range checks) and 5 criteria-based (qualitative rule checks). Covers long/short positions, profit/loss states, large positions, and multi-position portfolios.

**Prompt versioning:** Create `eval/analyzer_prompts/v2.txt` or `eval/judge_prompts/v2.txt` to A/B test prompt changes. Results save to `eval/eval_results/` with timestamps.

## Key Conventions

- DB connection string is hardcoded in `main.go` (host=localhost, user=edge, db=edge_interview)
- No external HTTP framework — pure `net/http` with manual path routing
- Only dependency: `github.com/lib/pq` (PostgreSQL driver)
- Alert conditions are `above` or `below`; statuses are `active`, `triggered`, `paused`
- All alert queries filter by `client_id` and `deleted_at IS NULL` for tenant isolation
