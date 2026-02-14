import { useEffect, useState } from "react";
import type { Commodity, PricePoint } from "../types";
import { listCommodities, getCurrentPrices, createAlert } from "../api";

interface AlertFormProps {
  onAlertCreated: () => void;
}

export default function AlertForm({ onAlertCreated }: AlertFormProps) {
  const [commodities, setCommodities] = useState<Commodity[]>([]);
  const [prices, setPrices] = useState<PricePoint[]>([]);
  const [commodityCode, setCommodityCode] = useState("");
  const [condition, setCondition] = useState<"above" | "below">("above");
  const [thresholdPrice, setThresholdPrice] = useState("");
  const [notes, setNotes] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    Promise.all([listCommodities(), getCurrentPrices()]).then(
      ([comms, prs]) => {
        setCommodities(comms);
        setPrices(prs);
        if (comms.length > 0) setCommodityCode(comms[0].code);
      }
    );
  }, []);

  function getCurrentPrice(): number | null {
    const commodity = commodities.find((c) => c.code === commodityCode);
    if (!commodity) return null;
    const pp = prices.find((p) => p.commodity_id === commodity.id);
    return pp ? pp.price : null;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(false);
    const price = parseFloat(thresholdPrice);
    if (!commodityCode || isNaN(price) || price <= 0) {
      setError("Please fill in all required fields with valid values.");
      return;
    }
    setSubmitting(true);
    try {
      await createAlert({
        commodity_code: commodityCode,
        condition,
        threshold_price: price,
        notes,
      });
      setSuccess(true);
      setThresholdPrice("");
      setNotes("");
      onAlertCreated();
      setTimeout(() => setSuccess(false), 3000);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Failed to create alert");
    } finally {
      setSubmitting(false);
    }
  }

  const currentPrice = getCurrentPrice();

  return (
    <div className="bg-white shadow-sm rounded-lg p-6">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">
        Create Alert
      </h2>

      {error && (
        <div className="mb-4 bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg text-sm">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-4 bg-green-50 border border-green-200 text-green-700 px-3 py-2 rounded-lg text-sm">
          Alert created successfully!
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Commodity
          </label>
          <select
            value={commodityCode}
            onChange={(e) => setCommodityCode(e.target.value)}
            className="block w-full rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 bg-white border px-3 py-2"
          >
            {commodities.map((c) => (
              <option key={c.id} value={c.code}>
                {c.name} ({c.code})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Condition
          </label>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setCondition("above")}
              className={`flex-1 px-4 py-2 text-sm font-medium rounded-lg border ${
                condition === "above"
                  ? "bg-indigo-600 text-white border-indigo-600"
                  : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
              }`}
            >
              Above
            </button>
            <button
              type="button"
              onClick={() => setCondition("below")}
              className={`flex-1 px-4 py-2 text-sm font-medium rounded-lg border ${
                condition === "below"
                  ? "bg-indigo-600 text-white border-indigo-600"
                  : "bg-white text-gray-700 border-gray-300 hover:bg-gray-50"
              }`}
            >
              Below
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Threshold Price ($)
          </label>
          <div className="relative">
            <input
              type="number"
              step="0.0001"
              min="0"
              value={thresholdPrice}
              onChange={(e) => setThresholdPrice(e.target.value)}
              placeholder="0.0000"
              className="block w-full rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 border px-3 py-2"
            />
          </div>
          {currentPrice !== null && (
            <p className="mt-1 text-xs text-gray-500">
              Current price: ${currentPrice.toFixed(4)}
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Notes
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
            placeholder="Optional notes about this alert..."
            className="block w-full rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 border px-3 py-2"
          />
        </div>

        <button
          type="submit"
          disabled={submitting}
          className="w-full inline-flex items-center justify-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          {submitting ? (
            <>
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Creating...
            </>
          ) : (
            "Create Alert"
          )}
        </button>
      </form>
    </div>
  );
}
