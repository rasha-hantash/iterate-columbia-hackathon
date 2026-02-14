import { useState } from "react";
import AlertList from "./AlertList";
import AlertForm from "./AlertForm";

export default function AlertsPage() {
  const [refreshKey, setRefreshKey] = useState(0);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div className="lg:col-span-2">
        <AlertList refreshKey={refreshKey} />
      </div>
      <div>
        <AlertForm onAlertCreated={() => setRefreshKey((k) => k + 1)} />
      </div>
    </div>
  );
}
