# Commodity Price Alert System with AI-Powered Risk Analysis

## Who Is This For?

Built for **commodity risk managers** and **procurement teams** monitoring agricultural commodity prices and protecting positions from adverse market shifts. A toy version of [Edge](https://try-edge.com).

## What Does It Do?

The platform is a **commodity risk monitoring system** backed by real 2023-2024 USDA wholesale terminal market data for corn. It combines:

- **Multi-tenant client isolation** — separate trading books per organization, enforced at the database and query level
- **Long and short position tracking** — with live P&L calculations against current market prices
- **AI-driven alert suggestions** — Claude analyzes a user's positions alongside current prices and 2023 seasonal patterns, then recommends protective stop-loss and take-profit alerts via a structured `create_alert` tool
- **A simulation engine** — replays 2024 USDA price data chronologically through a goroutine + channel pipeline, triggering alerts when thresholds are crossed
- **An offline eval framework** — 10 golden scenarios scored by an LLM-as-judge, with prompt versioning for A/B testing
- **A React frontend** — three-tab UI (AI Analysis, Alerts, Market Data) with user switching to demonstrate multi-tenant isolation

## Seeded Positions

| User | Direction | Volume | Entry Price | Meaning |
|------|-----------|--------|-------------|---------|
| Alice (Acme) | Long | 50,000 | $33.00 | Profits when corn rises above $33 |
| Alice (Acme) | Short | 20,000 | $38.00 | Profits when corn falls below $38 |
| Bob (Acme) | Long | 30,000 | $34.00 | Profits when corn rises above $34 |
| Carol (Global Grain) | Long | 100,000 | $32.00 | Profits when corn rises above $32 |

No alerts are pre-loaded. Alerts are created either manually or by accepting AI suggestions from the analysis endpoint.

## Demo Flow

1. **View positions** — The AI Analysis tab shows each user's positions with current price and unrealized P&L
2. **Generate alerts** — Click "Analyze My Positions" to have Claude suggest stop-loss and take-profit alerts based on positions + 2023 seasonal price patterns
3. **Accept suggestions** — Accept the AI's alert suggestions to create them, or create alerts manually
4. **Run simulation** — Start the 2024 simulation to replay a year of USDA corn prices; watch alerts trigger as thresholds are crossed
5. **Switch users** — Use the header dropdown to switch between Alice, Bob, and Carol to see tenant isolation in action

## How We Apply White Circle

The AI analysis endpoints send Claude's responses to [White Circle](https://whitecircle.ai) for real-time policy evaluation. This runs as a fire-and-forget goroutine — **zero added latency** to the user response. If credentials are not configured, it's silently disabled.

### Policies (Real-Time Guardrails)

| Policy | Flags | Allows |
|--------|-------|--------|
| **Financial Advice Guardrail** | Definitive trading instructions (e.g., "you must buy now") | Analytical suggestions framed as considerations for the user's review |
| **Commodity Scope Enforcement** | Off-topic responses about stocks, crypto, or non-commodity subjects | Discussions about agricultural commodity markets and related risk factors |

### Metrics (Async Product Analytics)

| Metric | What It Tracks |
|--------|---------------|
| **Take-Profit Suggestions** | How often the AI recommends alerts designed to capture gains |
| **Multi-Position Coverage** | Whether the AI addresses all of a user's open positions |
| **Historical Data Referenced** | How often the AI cites seasonal patterns and historical price trends |

### Why This Matters

In commodity risk management, a poorly scoped or overconfident AI response can lead to real financial harm. White Circle lets us enforce guardrails, monitor quality across prompt versions, and iterate with confidence. The combination of offline evals (golden dataset + LLM judge) and online evals (White Circle policies + metrics) creates a closed loop: test changes offline, deploy them, and continuously monitor behavior in production.
