import { useEffect, useState } from "react";
import type { Alert, Commodity } from "../types";
import { listAlerts, listCommodities } from "../api";

export default function AlertList({ refreshKey }: { refreshKey: number }) {
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [commodities, setCommodities] = useState<Commodity[]>([]);
  const [statusFilter, setStatusFilter] = useState("");
  const [commodityFilter, setCommodityFilter] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listCommodities()
      .then(setCommodities)
      .catch(() => {});
  }, []);

  useEffect(() => {
    setLoading(true);
    setError(null);
    listAlerts(statusFilter || undefined, commodityFilter || undefined)
      .then(setAlerts)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [statusFilter, commodityFilter, refreshKey]);

  function statusBadge(status: string) {
    const styles: Record<string, string> = {
      active: "bg-green-50 text-green-700 border border-green-200",
      triggered: "bg-red-50 text-red-700 border border-red-200",
      paused: "bg-yellow-50 text-yellow-700 border border-yellow-200",
    };
    return (
      <span
        className={`inline-flex px-1.5 py-0.5 text-[11px] font-medium rounded-sm ${
          styles[status] || "bg-sand-100 text-sand-600 border border-sand-200"
        }`}
      >
        {status}
      </span>
    );
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
        <h2 className="text-[13px] font-medium text-sand-800">Alerts</h2>
        <div className="flex gap-2">
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="rounded border border-sand-300 bg-sand-50 text-[12px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          >
            <option value="">All Statuses</option>
            <option value="active">Active</option>
            <option value="triggered">Triggered</option>
            <option value="paused">Paused</option>
          </select>
          <select
            value={commodityFilter}
            onChange={(e) => setCommodityFilter(e.target.value)}
            className="rounded border border-sand-300 bg-sand-50 text-[12px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
          >
            <option value="">All Commodities</option>
            {commodities.map((c) => (
              <option key={c.id} value={c.code}>
                {c.name}
              </option>
            ))}
          </select>
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
      ) : alerts.length === 0 ? (
        <div className="px-3 py-8 text-center text-[13px] text-sand-500">
          No alerts found. Create one or adjust your filters.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-sand-200">
            <thead>
              <tr className="bg-sand-100/50">
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Commodity
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Condition
                </th>
                <th className="px-3 py-2 text-right text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Threshold
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Notes
                </th>
                <th className="px-3 py-2 text-left text-[11px] font-medium text-sand-500 uppercase tracking-wider">
                  Created
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-sand-200/60">
              {alerts.map((a) => (
                <tr key={a.id} className="hover:bg-sand-100/50">
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] font-medium text-sand-800">
                    {a.commodity_name}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-600">
                    {a.condition}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[13px] text-sand-700 text-right tabular-nums">
                    ${a.threshold_price.toFixed(4)}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap">
                    {statusBadge(a.status)}
                  </td>
                  <td className="px-3 py-2.5 text-[13px] text-sand-500 whitespace-nowrap">
                    {a.notes || "-"}
                  </td>
                  <td className="px-3 py-2.5 whitespace-nowrap text-[12px] text-sand-400">
                    {formatDate(a.created_at)}
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
