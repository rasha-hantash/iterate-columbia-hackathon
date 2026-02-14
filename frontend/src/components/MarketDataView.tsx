import { useEffect, useState } from "react";
import type { MarketDataRow } from "../types";
import { getMarketData } from "../api";

export default function MarketDataView() {
  const [data, setData] = useState<MarketDataRow[]>([]);
  const [locations, setLocations] = useState<string[]>([]);
  const [locationFilter, setLocationFilter] = useState("");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    getMarketData(
      locationFilter || undefined,
      startDate || undefined,
      endDate || undefined
    )
      .then((rows) => {
        setData(rows);
        if (locations.length === 0 && rows.length > 0) {
          const unique = [...new Set(rows.map((r) => r.location))].sort();
          setLocations(unique);
        }
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [locationFilter, startDate, endDate]);

  function formatCurrency(val: number | null) {
    if (val == null) return "-";
    return val.toLocaleString("en-US", {
      style: "currency",
      currency: "USD",
    });
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  }

  return (
    <div className="bg-sand-50 border border-sand-200 rounded overflow-hidden">
      <div className="flex items-center justify-between h-10 px-3 border-b border-sand-200">
        <h2 className="text-[13px] font-medium text-sand-800">Market Data</h2>
        <div className="flex flex-wrap gap-2">
          <select
            value={locationFilter}
            onChange={(e) => setLocationFilter(e.target.value)}
            className="rounded border border-sand-300 bg-sand-50 text-[12px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          >
            <option value="">All Locations</option>
            {locations.map((loc) => (
              <option key={loc} value={loc}>
                {loc}
              </option>
            ))}
          </select>
          <input
            type="date"
            value={startDate}
            onChange={(e) => setStartDate(e.target.value)}
            className="rounded border border-sand-300 bg-sand-50 text-[12px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          />
          <input
            type="date"
            value={endDate}
            onChange={(e) => setEndDate(e.target.value)}
            className="rounded border border-sand-300 bg-sand-50 text-[12px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          />
        </div>
      </div>

      {error && (
        <div className="px-3 py-2 bg-red-50 border-b border-red-200 text-red-700 text-[13px]">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-sand-500"></div>
          <span className="ml-2 text-[13px] text-sand-500">Loading...</span>
        </div>
      ) : data.length === 0 ? (
        <div className="px-3 py-8 text-center text-[13px] text-sand-500">
          No market data found. Adjust filters or check the backend.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-sand-200">
            <thead>
              <tr className="bg-sand-100/50">
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Date
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Location
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Origin
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Low
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  High
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Properties
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-sand-200/60">
              {data.map((row) => (
                <tr key={row.id} className="hover:bg-sand-100/50">
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-600">
                    {formatDate(row.report_date)}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] font-medium text-sand-800">
                    {row.location}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-600">
                    {row.origin}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                    {formatCurrency(row.low_price)}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                    {formatCurrency(row.high_price)}
                  </td>
                  <td className="px-3 py-2.5 text-[12px] text-sand-500 max-w-xs truncate">
                    {row.properties || "-"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
