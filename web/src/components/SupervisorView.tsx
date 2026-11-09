import { useEffect, useState } from "react";
import * as api from "../api/client";
import type { Alert } from "../api/client";

interface Props {
  tenantId: number;
  onError: (cause: unknown) => void;
}

export function SupervisorView({ tenantId, onError }: Props) {
  const [alerts, setAlerts] = useState<Alert[]>([]);

  useEffect(() => {
    api.listAlerts(tenantId).then(setAlerts).catch(onError);
  }, [tenantId, onError]);

  return (
    <section>
      <h2>Įspėjimai</h2>
      {alerts.length === 0 && <p>Įspėjimų nėra.</p>}
      <ul>
        {alerts.map((alert) => (
          <li key={alert.id}>
            {new Date(alert.createdAt).toLocaleString("lt-LT")}: {alert.reason} (sesija #{alert.sessionId})
          </li>
        ))}
      </ul>
    </section>
  );
}
