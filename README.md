# Commodity Alert Manager

A commodity price alert microservice with AI-powered position analysis. Go backend with PostgreSQL, React frontend with Tailwind CSS.

## Prerequisites

- Go 1.21+
- Node.js 18+
- Docker (for PostgreSQL) or PostgreSQL running locally
- Anthropic API key

## Database Setup

### Option 1: Docker Compose (recommended)

```bash
docker compose up -d
```

This starts PostgreSQL 16 and automatically runs `db/init.sql` to create tables and seed data. Data persists in a Docker volume.

To stop: `docker compose down` (add `-v` to also delete data).

### Option 2: Existing PostgreSQL

```bash
psql -h localhost -U edge -d edge_interview < db/init.sql
```

This creates all tables and seeds test data (clients, users, commodities, positions, price history, sample alerts).

## Backend

```bash
export ANTHROPIC_API_KEY=your-key-here
cd platform && go run .
```

Server starts on http://localhost:8000. It will fail to start if `ANTHROPIC_API_KEY` is not set.

On first startup, the market data table is auto-created and populated from the included USDA CSV file (~3,700 rows).

### API Endpoints

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
| GET | /market-data | No | USDA market data (filters: `location`, `start_date`, `end_date`) |

Auth = `X-User-ID` header (1=Alice, 2=Bob, 3=Carol).

### Tests

```bash
cd platform && go test ./...
```

## Frontend

```bash
cd frontend && npm install
npm run dev
```

Dev server starts on http://localhost:5173.

### Tabs

- **AI Analysis** (default) -- View positions with P&L, click "Analyze My Positions" to get Claude-powered alert suggestions, accept them with one click
- **Alerts** -- Alert dashboard with status/commodity filters + create alert form
- **Market Data** -- Browse USDA terminal market data with location/date filters

Use the user selector in the header to switch between Alice, Bob, and Carol (demonstrates multi-tenant isolation).

## Seed Users

| ID | Name | Client | Role |
|----|------|--------|------|
| 1 | Alice Smith | Acme Foods | Risk Manager |
| 2 | Bob Jones | Acme Foods | Procurement |
| 3 | Carol Chen | Global Grain Co | Risk Manager |
