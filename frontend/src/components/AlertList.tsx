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
      active: "bg-green-100 text-green-800",
      triggered: "bg-red-100 text-red-800",
      paused: "bg-yellow-100 text-yellow-800",
    };
    return (
      <span
        className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
          styles[status] || "bg-gray-100 text-gray-800"
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
    <div className="bg-white shadow-sm rounded-lg overflow-hidden">
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <h2 className="text-lg font-semibold text-gray-900">Alerts</h2>
          <div className="flex gap-3">
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="block rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 bg-white border px-3 py-1.5"
            >
              <option value="">All Statuses</option>
              <option value="active">Active</option>
              <option value="triggered">Triggered</option>
              <option value="paused">Paused</option>
            </select>
            <select
              value={commodityFilter}
              onChange={(e) => setCommodityFilter(e.target.value)}
              className="block rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 bg-white border px-3 py-1.5"
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
      </div>

      {error && (
        <div className="px-6 py-3 bg-red-50 text-red-700 text-sm">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-600"></div>
          <span className="ml-3 text-gray-500 text-sm">Loading alerts...</span>
        </div>
      ) : alerts.length === 0 ? (
        <div className="px-6 py-8 text-center text-gray-500">
          No alerts found. Create one or adjust your filters.
        </div>
      ) : (
        <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Commodity
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Condition
              </th>
              <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                Threshold
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Notes
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Created
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {alerts.map((a) => (
              <tr key={a.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {a.commodity_name}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700">
                  {a.condition}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-700 text-right">
                  ${a.threshold_price.toFixed(4)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  {statusBadge(a.status)}
                </td>
                <td className="px-6 py-4 text-sm text-gray-600 whitespace-nowrap">
                  {a.notes || "-"}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
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
