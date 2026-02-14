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
    <div className="bg-sand-50 border border-sand-200 rounded p-4">
      <h2 className="text-[13px] font-medium text-sand-800 mb-3">
        Create Alert
      </h2>

      {error && (
        <div className="mb-3 bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded text-[13px]">
          {error}
        </div>
      )}

      {success && (
        <div className="mb-3 bg-green-50 border border-green-200 text-green-700 px-3 py-2 rounded text-[13px]">
          Alert created successfully!
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="block text-[12px] font-medium text-sand-600 mb-1">
            Commodity
          </label>
          <select
            value={commodityCode}
            onChange={(e) => setCommodityCode(e.target.value)}
            className="block w-full rounded border border-sand-300 bg-sand-50 text-[13px] text-sand-700 px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          >
            {commodities.map((c) => (
              <option key={c.id} value={c.code}>
                {c.name} ({c.code})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-[12px] font-medium text-sand-600 mb-1">
            Condition
          </label>
          <div className="flex gap-1.5">
            <button
              type="button"
              onClick={() => setCondition("above")}
              className={`flex-1 px-3 py-1.5 text-[13px] font-medium rounded border ${
                condition === "above"
                  ? "bg-sand-700 text-sand-50 border-sand-700"
                  : "bg-sand-50 text-sand-600 border-sand-300 hover:bg-sand-100"
              }`}
            >
              Above
            </button>
            <button
              type="button"
              onClick={() => setCondition("below")}
              className={`flex-1 px-3 py-1.5 text-[13px] font-medium rounded border ${
                condition === "below"
                  ? "bg-sand-700 text-sand-50 border-sand-700"
                  : "bg-sand-50 text-sand-600 border-sand-300 hover:bg-sand-100"
              }`}
            >
              Below
            </button>
          </div>
        </div>

        <div>
          <label className="block text-[12px] font-medium text-sand-600 mb-1">
            Threshold Price ($)
          </label>
          <input
            type="number"
            step="0.0001"
            min="0"
            value={thresholdPrice}
            onChange={(e) => setThresholdPrice(e.target.value)}
            placeholder="0.0000"
            className="block w-full rounded border border-sand-300 bg-sand-50 text-[13px] text-sand-700 px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          />
          {currentPrice !== null && (
            <p className="mt-1 text-[11px] text-sand-400 tabular-nums">
              Current: ${currentPrice.toFixed(4)}
            </p>
          )}
        </div>

        <div>
          <label className="block text-[12px] font-medium text-sand-600 mb-1">
            Notes
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            rows={3}
            placeholder="Optional notes..."
            className="block w-full rounded border border-sand-300 bg-sand-50 text-[13px] text-sand-700 px-2.5 py-1.5 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          />
        </div>

        <button
          type="submit"
          disabled={submitting}
          className="w-full inline-flex items-center justify-center px-3 py-1.5 text-[13px] font-medium rounded text-sand-50 bg-sand-700 hover:bg-sand-800 focus:outline-none focus:ring-1 focus:ring-sand-500 disabled:opacity-50"
        >
          {submitting ? (
            <>
              <div className="animate-spin rounded-full h-3.5 w-3.5 border-b-2 border-sand-200 mr-1.5"></div>
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
