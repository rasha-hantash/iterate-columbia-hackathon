# Commodity Alert Manager

A commodity price alert microservice with AI-powered position analysis and real-time risk monitoring. Go backend with PostgreSQL, React frontend with Tailwind CSS.

## Quick Start

### 1. Start the Database

**Docker Compose (recommended):**

```bash
docker compose up -d
```

**Or existing PostgreSQL:**

```bash
psql -h localhost -U edge -d edge_interview < db/init.sql
```

### 2. Start the Backend

```bash
export ANTHROPIC_API_KEY=your-key-here
cd platform && go run .
```

Server starts on http://localhost:8000. On first startup, both 2023 and 2024 USDA market data CSVs are auto-imported (~8,300 rows total).

### 3. Start the Frontend

```bash
cd frontend && npm install && npm run dev
```

Dev server starts on http://localhost:5173.

## Running Tests

```bash
cd platform && go test ./...
```

## Risk Monitoring System

The risk monitoring system uses 2023 wholesale terminal market data to establish typical corn prices, then simulates real-time processing of 2024 data to check and trigger alerts.

### Step 1: View Monthly Price Analysis

Get typical monthly corn prices from 2023 data (computed via SQL aggregation):

```bash
curl http://localhost:8000/market-data/monthly-analysis?year=2023&commodity=corn
```

Returns monthly avg, min, max, and sample counts. No third-party tools needed.

### Step 2: Generate Alerts with AI

Use Claude to suggest alerts based on your positions and 2023 price patterns:

```bash
curl -X POST -H "X-User-ID: 1" http://localhost:8000/analyze-positions-market
```

Or create alerts manually:

```bash
curl -X POST -H "X-User-ID: 1" -H "Content-Type: application/json" \
  -d '{"commodity_code":"CORN","condition":"below","threshold_price":28}' \
  http://localhost:8000/alerts
```

### Step 3: Run the Simulation

Process 2024 data as a real-time price feed (goroutine + channel pipeline simulating a websocket):

```bash
# Start simulation (speed = ms between each date, default 500)
curl -X POST "http://localhost:8000/simulation/start?speed=100&user_id=1"

# Check progress and triggered alerts
curl http://localhost:8000/simulation/status

# Stop mid-simulation
curl -X POST http://localhost:8000/simulation/stop

# Reset triggered alerts back to active for re-run
curl -X POST "http://localhost:8000/simulation/reset?user_id=1"
```

The simulation processes each 2024 date chronologically, computes a representative price, and auto-triggers any active alerts whose conditions are met.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /health | No | Health check |
| GET | /commodities | No | List all commodities |
| GET | /prices | No | Latest price per commodity |
| GET | /positions | Yes | User's positions with commodity info |
| POST | /alerts | Yes | Create a price alert |
| GET | /alerts | Yes | List alerts (filters: `status`, `commodity_code`) |
| POST | /alerts/{id}/trigger | Yes | Trigger an alert |
| POST | /analyze-positions | Yes | AI analysis of user positions |
| POST | /analyze-positions-market | Yes | AI analysis with 2023 wholesale price context |
| GET | /market-data | No | USDA market data (filters: `location`, `start_date`, `end_date`) |
| GET | /market-data/monthly-analysis | No | Monthly price statistics (params: `year`, `commodity`) |
| POST | /simulation/start | No | Start 2024 simulation (params: `speed`, `user_id`) |
| GET | /simulation/status | No | Current simulation progress and results |
| POST | /simulation/stop | No | Stop running simulation |
| POST | /simulation/reset | No | Reset triggered alerts for re-run (param: `user_id`) |

Auth = `X-User-ID` header (1=Alice, 2=Bob, 3=Carol).

## LLM Evals

Offline evaluation framework that tests the quality of Claude's alert suggestions using an LLM-as-judge approach. Lives in `eval/`.

```bash
cd eval && uv sync
ANTHROPIC_API_KEY=your-key uv run python scripts/run_eval.py --verbose
```

Requires Python 3.11+ and [uv](https://docs.astral.sh/uv/).

Two-phase pipeline against a golden dataset of 10 CORN scenarios:

1. **Analyzer** -- Sends positions + prices to Claude with a `create_alert` tool, captures alert suggestions
2. **Judge** -- A separate Claude call evaluates whether the suggestions are correct

```bash
# Only run analyzer, skip judging
uv run python scripts/run_eval.py --skip-judge -v

# Only run judge on existing responses
uv run python scripts/run_eval.py --skip-analysis -v

# A/B test a new analyzer prompt
uv run python scripts/run_eval.py --analyzer-prompt v2 -v
```

Results save to `eval/eval_results/`. See `eval/EVAL.md` for full documentation.

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (for PostgreSQL) or PostgreSQL running locally
- Anthropic API key

## Architecture

All Go source lives in `platform/` as a single `main` package.

- `main.go` -- HTTP server, routing, auth middleware
- `handler.go` -- Request parsing, validation, JSON responses
- `service.go` -- Business logic, SQL queries, monthly price analysis
- `ai_handler.go` -- Claude API integration for position analysis
- `csv_import.go` -- Auto-imports 2023/2024 USDA market data CSVs on startup
- `simulation.go` -- Real-time simulation engine with goroutine + channel pipeline

Database schema and seed data live in `db/init.sql`. Key tables: `clients`, `users`, `commodities`, `positions`, `price_data`, `price_alerts`, `alert_history`, `market_data`.

## Frontend Tabs

- **AI Analysis** (default) -- View positions with P&L, get Claude-powered alert suggestions
- **Alerts** -- Alert dashboard with status/commodity filters + create alert form
- **Market Data** -- Browse USDA terminal market data with location/date filters

Use the user selector in the header to switch between Alice, Bob, and Carol (demonstrates multi-tenant isolation).

## Seed Users

| ID | Name | Client | Role |
|----|------|--------|------|
| 1 | Alice Smith | Acme Foods | Risk Manager |
| 2 | Bob Jones | Acme Foods | Procurement |
| 3 | Carol Chen | Global Grain Co | Risk Manager |
