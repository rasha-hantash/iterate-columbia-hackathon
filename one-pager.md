# Commodity Price Alert System with AI-Powered Risk Analysis

## Who Is This For?

This system is built for **commodity risk managers**, **procurement teams**, and **agricultural trading desks** — anyone responsible for monitoring price movements in agricultural commodities (corn, wheat, soybean oil) and protecting their positions from adverse market shifts.

Typical users include:

- **Risk managers** at food companies who need automated alerts when commodity prices approach stop-loss or take-profit thresholds
- **Procurement officers** managing buy-side exposure who need early warning on price spikes
- **Portfolio analysts** overseeing multi-commodity positions across a trading book

## What Does It Simulate?

The platform simulates a **real-time commodity risk monitoring environment** backed by 2023-2024 USDA wholesale terminal market data. It models:

- **Multi-tenant client isolation** — separate trading books per organization, enforced at the database level
- **Long and short position tracking** — with live P&L calculations against current market prices
- **AI-driven alert suggestions** — Claude analyzes a user's positions alongside current prices and historical seasonal patterns, then recommends protective stop-loss and take-profit alerts with specific price thresholds
- **A real-time simulation engine** — streams price movements and triggers alerts when thresholds are crossed, mimicking a production monitoring loop
- **An offline LLM evaluation framework** — a golden dataset of 10 scenarios (strict price-range checks and criteria-based qualitative checks) scored by a judge model, enabling prompt versioning and A/B testing before shipping changes

The result is an end-to-end demonstration of how an AI copilot can augment human risk decisions in commodity markets — from position analysis, to alert creation, to triggered notifications.

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
