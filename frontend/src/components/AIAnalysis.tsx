import { useEffect, useState } from "react";
import type { Position, PricePoint, AlertSuggestion } from "../types";
import {
  listPositions,
  getCurrentPrices,
  analyzePositions,
  createAlert,
} from "../api";

export default function AIAnalysis() {
  const [positions, setPositions] = useState<Position[]>([]);
  const [prices, setPrices] = useState<PricePoint[]>([]);
  const [loading, setLoading] = useState(true);
  const [analyzing, setAnalyzing] = useState(false);
  const [reasoning, setReasoning] = useState<string | null>(null);
  const [suggestions, setSuggestions] = useState<AlertSuggestion[]>([]);
  const [accepted, setAccepted] = useState<Set<number>>(new Set());
  const [accepting, setAccepting] = useState<Set<number>>(new Set());
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setReasoning(null);
    setSuggestions([]);
    setAccepted(new Set());
    setAccepting(new Set());
    setError(null);
    Promise.all([listPositions(), getCurrentPrices()])
      .then(([pos, pr]) => {
        setPositions(pos);
        setPrices(pr);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  function getPriceForCommodity(commodityId: number): number | null {
    const p = prices.find((pr) => pr.commodity_id === commodityId);
    return p ? p.price : null;
  }

  function calcPnL(pos: Position, currentPrice: number): number {
    if (pos.direction === "long") {
      return (currentPrice - pos.entry_price) * pos.volume;
    }
    return (pos.entry_price - currentPrice) * pos.volume;
  }

  async function handleAnalyze() {
    setAnalyzing(true);
    setError(null);
    setReasoning(null);
    setSuggestions([]);
    setAccepted(new Set());
    try {
      const res = await analyzePositions();
      setReasoning(res.reasoning);
      setSuggestions(res.suggestions || []);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Analysis failed");
    } finally {
      setAnalyzing(false);
    }
  }

  async function handleAccept(idx: number, suggestion: AlertSuggestion) {
    setAccepting((prev) => new Set(prev).add(idx));
    try {
      await createAlert({
        commodity_code: suggestion.commodity_code,
        condition: suggestion.condition,
        threshold_price: suggestion.threshold_price,
        notes: suggestion.notes,
      });
      setAccepted((prev) => new Set(prev).add(idx));
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to create alert"
      );
    } finally {
      setAccepting((prev) => {
        const next = new Set(prev);
        next.delete(idx);
        return next;
      });
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-10">
        <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-sand-500"></div>
        <span className="ml-2 text-[13px] text-sand-500">Loading positions...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded text-[13px]">
          {error}
        </div>
      )}

      {/* Positions Table */}
      <div className="bg-sand-50 border border-sand-200 rounded overflow-hidden">
        <div className="flex items-center justify-between h-10 px-3 border-b border-sand-200">
          <h2 className="text-[13px] font-medium text-sand-800">
            Your Positions
          </h2>
        </div>
        {positions.length === 0 ? (
          <div className="px-3 py-6 text-center text-[13px] text-sand-500">
            No positions found for this user.
          </div>
        ) : (
          <table className="min-w-full divide-y divide-sand-200">
            <thead>
              <tr className="bg-sand-100/50">
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Commodity
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Volume
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Direction
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Entry
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Current
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  P&L
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-sand-200/60">
              {positions.map((pos) => {
                const currentPrice = getPriceForCommodity(pos.commodity_id);
                const pnl =
                  currentPrice !== null ? calcPnL(pos, currentPrice) : null;
                return (
                  <tr key={pos.id} className="hover:bg-sand-100/50">
                    <td className="px-3 py-2.5 whitespace-nowrap text-[13px] font-medium text-sand-800">
                      {pos.commodity_name}
                      <span className="ml-1.5 text-[11px] text-sand-400">
                        {pos.commodity_code}
                      </span>
                    </td>
                    <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                      {pos.volume.toLocaleString()}
                    </td>
                    <td className="px-3 py-2.5 whitespace-nowrap">
                      <span
                        className={`inline-flex px-1.5 py-0.5 text-[11px] font-medium rounded-sm ${
                          pos.direction === "long"
                            ? "bg-green-50 text-green-700 border border-green-200"
                            : "bg-red-50 text-red-700 border border-red-200"
                        }`}
                      >
                        {pos.direction.toUpperCase()}
                      </span>
                    </td>
                    <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                      ${pos.entry_price.toFixed(4)}
                    </td>
                    <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                      {currentPrice !== null
                        ? `$${currentPrice.toFixed(4)}`
                        : "N/A"}
                    </td>
                    <td
                      className={`px-3 py-2.5 whitespace-nowrap text-[13px] font-medium text-right tabular-nums ${
                        pnl === null
                          ? "text-sand-400"
                          : pnl >= 0
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {pnl !== null
                        ? `${pnl >= 0 ? "+" : ""}$${pnl.toLocaleString(
                            undefined,
                            {
                              minimumFractionDigits: 2,
                              maximumFractionDigits: 2,
                            }
                          )}`
                        : "N/A"}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Analyze Button */}
      <div className="flex justify-center">
        <button
          onClick={handleAnalyze}
          disabled={analyzing || positions.length === 0}
          className="inline-flex items-center px-4 py-2 text-[13px] font-medium rounded text-sand-50 bg-sand-700 hover:bg-sand-800 focus:outline-none focus:ring-1 focus:ring-sand-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {analyzing ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-sand-200 mr-2"></div>
              Analyzing...
            </>
          ) : (
            <>
              <svg
                className="w-4 h-4 mr-1.5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                />
              </svg>
              Analyze Positions
            </>
          )}
        </button>
      </div>

      {/* Reasoning */}
      {reasoning && (
        <div className="bg-sand-50 border border-sand-200 border-l-2 border-l-sand-500 rounded p-4">
          <h3 className="text-[11px] font-medium text-sand-500 uppercase tracking-wider mb-1.5">
            AI Analysis
          </h3>
          <p className="text-[13px] text-sand-700 whitespace-pre-wrap leading-relaxed">
            {reasoning}
          </p>
        </div>
      )}

      {/* Suggestion Cards */}
      {suggestions.length > 0 && (
        <div>
          <h3 className="text-[13px] font-medium text-sand-800 mb-2">
            Suggested Alerts
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {suggestions.map((s, idx) => (
              <div
                key={idx}
                className="bg-sand-50 border border-sand-200 rounded p-3"
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-[13px] font-semibold text-sand-800">
                    {s.commodity_code}
                  </span>
                  <span
                    className={`inline-flex px-1.5 py-0.5 text-[11px] font-medium rounded-sm ${
                      s.condition === "above"
                        ? "bg-blue-50 text-blue-700 border border-blue-200"
                        : "bg-orange-50 text-orange-700 border border-orange-200"
                    }`}
                  >
                    {s.condition.toUpperCase()}
                  </span>
                </div>
                <div className="text-xl font-semibold text-sand-800 mb-1.5 tabular-nums">
                  ${s.threshold_price.toFixed(4)}
                </div>
                <p className="text-[12px] text-sand-600 mb-3 leading-relaxed">{s.notes}</p>
                <button
                  onClick={() => handleAccept(idx, s)}
                  disabled={accepted.has(idx) || accepting.has(idx)}
                  className={`w-full inline-flex items-center justify-center px-3 py-1.5 text-[12px] font-medium rounded ${
                    accepted.has(idx)
                      ? "bg-green-50 text-green-700 border border-green-200 cursor-default"
                      : "bg-sand-700 text-sand-50 hover:bg-sand-800"
                  } disabled:opacity-75`}
                >
                  {accepting.has(idx) ? (
                    <>
                      <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-sand-200 mr-1.5"></div>
                      Creating...
                    </>
                  ) : accepted.has(idx) ? (
                    <>
                      <svg
                        className="w-3 h-3 mr-1"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                      Created
                    </>
                  ) : (
                    "Accept Alert"
                  )}
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
