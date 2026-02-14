# Commodity Price Alert System with AI-Powered Risk Analysis

## Who Is This For?

This system is built for **commodity risk managers**, **procurement teams**, and **agricultural trading desks** — anyone responsible for monitoring price movements in agricultural commodities (corn) and protecting their positions from adverse market shifts. It's a toy version of [Edge](https://try-edge.com).

Typical users include:

- **Risk managers** at food companies who need automated alerts when commodity prices approach stop-loss or take-profit thresholds
- **Procurement officers** managing buy-side exposure who need early warning on price spikes
- **Portfolio analysts** overseeing corn positions across a trading book

## What Does It Simulate?

The platform simulates a **real-time commodity risk monitoring environment** backed by 2023-2024 USDA wholesale terminal market data for corn. It models:

- **Multi-tenant client isolation** — separate trading books per organization, enforced at the database level
- **Long and short position tracking** — with live P&L calculations against current market prices
- **AI-driven alert suggestions** — Claude analyzes a user's positions alongside current prices and historical seasonal patterns, then recommends protective stop-loss and take-profit alerts with specific price thresholds
- **A real-time simulation engine** — streams price movements and triggers alerts when thresholds are crossed, mimicking a production monitoring loop
- **An offline LLM evaluation framework** — a golden dataset of 10 scenarios (strict price-range checks and criteria-based qualitative checks) scored by a judge model, enabling prompt versioning and A/B testing before shipping changes

The result is an end-to-end demonstration of how an AI copilot can augment human risk decisions in commodity markets — from position analysis, to alert creation, to triggered notifications.

## Demo Walkthrough: What to Expect

### Seeded Positions

| User | Direction | Volume | Entry Price | Meaning |
|------|-----------|--------|-------------|---------|
| Alice (Acme) | Long | 50,000 crates | $33.00 | Profits when corn rises above $33 |
| Alice (Acme) | Short | 20,000 crates | $38.00 | Profits when corn falls below $38 |
| Bob (Acme) | Long | 30,000 crates | $34.00 | Profits when corn rises above $34 |
| Carol (Global Grain) | Long | 100,000 crates | $32.00 | Profits when corn rises above $32 |

### Seeded Alerts (Pre-loaded)

| Alert | User | Condition | Threshold | Status | Purpose |
|-------|------|-----------|-----------|--------|---------|
| 1 | Alice | Below $28.00 | $28.00 | Active | Stop-loss protecting the long position |
| 2 | Alice | Above $42.00 | $42.00 | Active | Take-profit on the short position |
| 3 | Bob | Below $27.00 | $27.00 | Active | Watching for a corn dip |
| 4 | Carol | Below $29.00 | $29.00 | Triggered | Already fired (demo of triggered state) |

### 2024 Simulation: Predicted Alert Triggers

When you start the simulation with `user_id=1` (Alice, Acme Foods), the system processes 2024 USDA corn data chronologically. Each alert fires **once** — the first date its condition is met — then flips to "triggered."

**2024 corn price shape:**
- **Jan-Feb**: Prices high ($33-$46)
- **Mar onwards**: Prices collapse to $14-$28
- **Late Dec (16-23)**: Brief spike back to $44-$46
- **Dec 26-31**: Settles to $29-$39

**Trigger timeline:**

| Order | Date | Rep. Price | Alert Triggered | What Happens |
|-------|------|-----------|-----------------|--------------|
| 1 | **Jan 23, 2024** | ~$43.57 | Alert 2 (above $42) | Corn spikes above Alice's short take-profit — alert fires |
| 2 | **Mar 11, 2024** | ~$27.69 | Alert 1 (below $28) | Corn drops below Alice's long stop-loss — alert fires |
| 3 | **Mar 12, 2024** | ~$26.75 | Alert 3 (below $27) | Corn continues falling — Bob's dip alert fires next day |

After March 12, all active alerts for Acme Foods are triggered. The simulation continues processing through December but no additional alerts fire.

**Key takeaway for the demo:** The simulation shows realistic alert behavior — a short-position take-profit fires during a January price spike, then stop-losses cascade during the March price collapse. This is exactly how a risk monitoring system should work.

## How We Apply White Circle

The AI analysis feature sends commodity positions to Claude and returns alert suggestions. Because this is a financial context with real risk implications, we integrate **White Circle** as an online evaluation layer to enforce content safety and track product quality — without adding latency to the user experience.

### Architecture

When a user requests AI analysis, the system:

1. Calls Claude to generate alert suggestions based on positions, current prices, and historical data
2. Returns the response to the user immediately
3. **Asynchronously** sends the full conversation (user request + AI response) to White Circle's `/api/session/check` endpoint in a background goroutine

This fire-and-forget pattern means White Circle adds **zero latency** to the user-facing response. If White Circle credentials are not configured, the feature is silently disabled — the system degrades gracefully.

### Policies (Real-Time Guardrails)

We define two content moderation policies that White Circle evaluates on every AI response:

| Policy | Flags | Allows |
|--------|-------|--------|
| **Financial Advice Guardrail** | Definitive trading instructions (e.g., "you must buy now") | Analytical suggestions framed as considerations for the user's review |
| **Commodity Scope Enforcement** | Off-topic responses about stocks, crypto, or non-commodity subjects | Discussions about agricultural commodity markets and related risk factors |

These policies ensure the AI stays within its intended role — a risk analysis assistant, not a financial advisor — and remains scoped to agricultural commodities.

### Metrics (Async Product Analytics)

We define three metrics that White Circle computes in the background and surfaces in its dashboard:

| Metric | What It Tracks |
|--------|---------------|
| **Take-Profit Suggestions** | How often the AI recommends alerts designed to capture gains, not just limit losses |
| **Multi-Position Coverage** | Whether the AI addresses all of a user's open positions, not just a subset |
| **Historical Data Referenced** | How often the AI cites seasonal patterns and historical price trends in its reasoning |

These metrics provide ongoing visibility into the quality and completeness of AI suggestions — without requiring manual review of every response.

### Why This Matters

In commodity risk management, a poorly scoped or overconfident AI response can lead to real financial harm. White Circle lets us:

- **Enforce guardrails** — flag responses that cross from analysis into financial advice
- **Monitor quality** — track whether the AI is thorough (covering all positions) and grounded (referencing historical data)
- **Iterate with confidence** — compare metrics across prompt versions to validate improvements before shipping

The combination of offline evals (golden dataset + LLM judge) and online evals (White Circle policies + metrics) creates a closed loop: we test changes offline, deploy them, and continuously monitor their behavior in production.
