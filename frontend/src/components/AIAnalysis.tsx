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
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        <span className="ml-3 text-gray-600">Loading positions...</span>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {/* Positions Table */}
      <div className="bg-white shadow-sm rounded-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">
            Your Positions
          </h2>
        </div>
        {positions.length === 0 ? (
          <div className="px-6 py-8 text-center text-gray-500">
            No positions found for this user.
          </div>
        ) : (
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Commodity
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Volume
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Direction
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Entry Price
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Current Price
                </th>
                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Unrealized P&L
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {positions.map((pos) => {
                const currentPrice = getPriceForCommodity(pos.commodity_id);
                const pnl =
                  currentPrice !== null ? calcPnL(pos, currentPrice) : null;
                return (
                  <tr key={pos.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {pos.commodity_name}
                      <span className="ml-2 text-xs text-gray-500">
                        {pos.commodity_code}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700 text-right">
                      {pos.volume.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span
                        className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          pos.direction === "long"
                            ? "bg-green-100 text-green-800"
                            : "bg-red-100 text-red-800"
                        }`}
                      >
                        {pos.direction.toUpperCase()}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700 text-right">
                      ${pos.entry_price.toFixed(4)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700 text-right">
                      {currentPrice !== null
                        ? `$${currentPrice.toFixed(4)}`
                        : "N/A"}
                    </td>
                    <td
                      className={`px-6 py-4 whitespace-nowrap text-sm font-semibold text-right ${
                        pnl === null
                          ? "text-gray-400"
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
          className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-lg shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {analyzing ? (
            <>
              <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-3"></div>
              Claude is analyzing your positions...
            </>
          ) : (
            <>
              <svg
                className="w-5 h-5 mr-2"
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
              Analyze My Positions
            </>
          )}
        </button>
      </div>

      {/* Reasoning */}
      {reasoning && (
        <div className="bg-white shadow-sm rounded-lg p-6 border-l-4 border-indigo-500">
          <h3 className="text-sm font-semibold text-indigo-700 uppercase tracking-wider mb-2">
            AI Analysis
          </h3>
          <p className="text-gray-700 whitespace-pre-wrap leading-relaxed">
            {reasoning}
          </p>
        </div>
      )}

      {/* Suggestion Cards */}
      {suggestions.length > 0 && (
        <div>
          <h3 className="text-lg font-semibold text-gray-900 mb-3">
            Suggested Alerts
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {suggestions.map((s, idx) => (
              <div
                key={idx}
                className="bg-white shadow-sm rounded-lg p-5 border border-gray-200"
              >
                <div className="flex items-center justify-between mb-3">
                  <span className="text-sm font-bold text-gray-900">
                    {s.commodity_code}
                  </span>
                  <span
                    className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                      s.condition === "above"
                        ? "bg-blue-100 text-blue-800"
                        : "bg-orange-100 text-orange-800"
                    }`}
                  >
                    {s.condition.toUpperCase()}
                  </span>
                </div>
                <div className="text-2xl font-bold text-gray-900 mb-2">
                  ${s.threshold_price.toFixed(4)}
                </div>
                <p className="text-sm text-gray-600 mb-4">{s.notes}</p>
                <button
                  onClick={() => handleAccept(idx, s)}
                  disabled={accepted.has(idx) || accepting.has(idx)}
                  className={`w-full inline-flex items-center justify-center px-4 py-2 text-sm font-medium rounded-lg ${
                    accepted.has(idx)
                      ? "bg-green-100 text-green-800 cursor-default"
                      : "bg-indigo-600 text-white hover:bg-indigo-700"
                  } disabled:opacity-75`}
                >
                  {accepting.has(idx) ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Creating...
                    </>
                  ) : accepted.has(idx) ? (
                    <>
                      <svg
                        className="w-4 h-4 mr-2"
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
                      Alert Created
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
