import { useEffect, useState } from "react";
import * as api from "../api/client";
import type { Alert, State } from "../api/client";
import { alertLabel } from "../labels";

const refreshIntervalMs = 5000;

interface Props {
  tenantId: number;
  state: State;
  onRefresh: () => void;
  onError: (cause: unknown) => void;
}

export function SupervisorView({ tenantId, state, onRefresh, onError }: Props) {
  const [alerts, setAlerts] = useState<Alert[]>([]);

  useEffect(() => {
    const load = () => api.listAlerts(tenantId).then(setAlerts).catch(onError);
    load();
    const timer = setInterval(() => {
      load();
      onRefresh();
    }, refreshIntervalMs);
    return () => clearInterval(timer);
  }, [tenantId, onRefresh, onError]);

  const signals = alerts.filter((alert) => alert.reason === "CHASING_SIGNAL").length;
  const rapid = alerts.filter((alert) => alert.reason === "RAPID_BETTING").length;
  const stopped = alerts.some((alert) => alert.reason === "CHASING_LIMIT");
  const session = state.activeSession;

  return (
    <section>
      <h2>Prižiūrėtojo įspėjimai</h2>
      <Banner signals={signals} rapid={rapid} stopped={stopped} />
      {session && (
        <div className="tiles">
          <div className="tile">
            <strong>{state.chasingSignalsThisWeek} / 3</strong>
            <small>Signalai šią savaitę</small>
          </div>
          <div className="tile">
            <strong>{session.rapidFlags}</strong>
            <small>Per greiti statymai</small>
          </div>
          <div className="tile">
            <strong>{session.betCount}</strong>
            <small>Statymai</small>
          </div>
          <div className="tile">
            <strong>{session.netLoss}</strong>
            <small>Sesijos nuostolis</small>
          </div>
        </div>
      )}
      <h3>Istorija</h3>
      {alerts.length === 0 && <p className="muted">Įrašų nėra.</p>}
      <ul className="alerts">
        {alerts.map((alert) => (
          <li key={alert.id} className={`notice ${alert.reason === "CHASING_LIMIT" ? "danger" : "warning"}`}>
            {alertLabel(alert.reason)}
            <small>
              {new Date(alert.createdAt).toLocaleString("lt-LT")} · sesija #{alert.sessionId}
            </small>
          </li>
        ))}
      </ul>
    </section>
  );
}

function Banner({ signals, rapid, stopped }: { signals: number; rapid: number; stopped: boolean }) {
  if (stopped) {
    return <div className="banner danger">Sesija buvo sustabdyta dėl vijimosi. Signalų iki sustabdymo: {signals}</div>;
  }
  if (signals > 0 || rapid > 0) {
    return (
      <div className="banner warning">
        Vijimosi signalų: {signals} · per greitų statymų įspėjimų: {rapid}
      </div>
    );
  }
  return <div className="banner">Įspėjamųjų ženklų nėra</div>;
}
