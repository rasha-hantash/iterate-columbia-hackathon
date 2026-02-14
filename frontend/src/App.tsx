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
    <div className="min-h-screen bg-sand-100">
      {/* Header */}
      <header className="bg-sand-50 border-b border-sand-200 sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-4">
          <div className="flex items-center justify-between h-11">
            <div className="flex items-center gap-6">
              <h1 className="text-[13px] font-semibold text-sand-800">
                Commodity Alerts
              </h1>
              <nav className="flex items-center gap-1">
                {tabs.map((tab) => (
                  <button
                    key={tab.key}
                    onClick={() => setActiveTab(tab.key)}
                    className={`px-2.5 py-1 rounded text-[13px] font-medium ${
                      activeTab === tab.key
                        ? "bg-sand-200 text-sand-800"
                        : "text-sand-500 hover:text-sand-700 hover:bg-sand-200/50"
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </nav>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-[12px] text-sand-500">User:</span>
              <select
                value={userId}
                onChange={handleUserChange}
                className="rounded border border-sand-300 bg-sand-50 text-[13px] text-sand-700 px-2 py-1 focus:outline-none focus:ring-1 focus:ring-sand-400 focus:border-sand-400"
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

      {/* Content */}
      <main className="max-w-6xl mx-auto px-4 py-4">
        <div key={renderKey}>
          {activeTab === "analysis" && <AIAnalysis />}
          {activeTab === "alerts" && <AlertsPage />}
          {activeTab === "market" && <MarketDataView />}
        </div>
      </main>
    </div>
  );
}
