import { useState, useCallback } from "react";
import { setCurrentUserId, getCurrentUserId } from "./api";
import AIAnalysis from "./components/AIAnalysis";
import AlertsPage from "./components/AlertsPage";
import MarketDataView from "./components/MarketDataView";

const USERS = [
  { id: "1", name: "Alice Smith" },
  { id: "2", name: "Bob Jones" },
  { id: "3", name: "Carol Chen" },
];

type Tab = "analysis" | "alerts" | "market";

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>("analysis");
  const [userId, setUserId] = useState(getCurrentUserId());
  const [renderKey, setRenderKey] = useState(0);

  const handleUserChange = useCallback(
    (e: React.ChangeEvent<HTMLSelectElement>) => {
      const newId = e.target.value;
      setCurrentUserId(newId);
      setUserId(newId);
      setRenderKey((k) => k + 1);
    },
    []
  );

  const tabs: { key: Tab; label: string }[] = [
    { key: "analysis", label: "AI Analysis" },
    { key: "alerts", label: "Alerts" },
    { key: "market", label: "Market Data" },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <h1 className="text-xl font-bold text-gray-900">
              Commodity Alert Manager
            </h1>
            <div className="flex items-center gap-3">
              <label className="text-sm text-gray-600">User:</label>
              <select
                value={userId}
                onChange={handleUserChange}
                className="block rounded-md border-gray-300 shadow-sm text-sm focus:border-indigo-500 focus:ring-indigo-500 bg-white border px-3 py-1.5"
              >
                {USERS.map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      </header>

      {/* Tabs */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-4">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`whitespace-nowrap py-3 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.key
                    ? "border-indigo-600 text-indigo-600"
                    : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>
      </div>

      {/* Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div key={renderKey}>
          {activeTab === "analysis" && <AIAnalysis />}
          {activeTab === "alerts" && <AlertsPage />}
          {activeTab === "market" && <MarketDataView />}
        </div>
      </main>
    </div>
  );
}
