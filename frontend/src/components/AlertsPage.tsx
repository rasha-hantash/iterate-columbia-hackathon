import { useState } from "react";
import AlertList from "./AlertList";

export default function AlertsPage() {
  const [refreshKey] = useState(0);

  return (
    <div>
      <AlertList refreshKey={refreshKey} />
    </div>
  );
}
