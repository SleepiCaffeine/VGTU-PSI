import { useCallback, useEffect, useState } from "react";
import * as api from "./api/client";
import type { State, Tenant } from "./api/client";
import { GamblerView } from "./components/GamblerView";
import { HelpContacts } from "./components/HelpContacts";
import { ProgressPanel } from "./components/ProgressPanel";
import { SupervisorView } from "./components/SupervisorView";

type Mode = "gambler" | "supervisor";

export function App() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenantId, setTenantId] = useState<number>();
  const [mode, setMode] = useState<Mode>("gambler");
  const [state, setState] = useState<State>();
  const [error, setError] = useState<string>();

  const report = useCallback((cause: unknown) => {
    setError(cause instanceof Error ? cause.message : "Įvyko klaida");
  }, []);

  useEffect(() => {
    api
      .listTenants()
      .then((list) => {
        setTenants(list);
        setTenantId(list[0]?.id);
      })
      .catch(report);
  }, [report]);

  const refresh = useCallback(() => {
    if (tenantId === undefined) return;
    setError(undefined);
    api.getState(tenantId).then(setState).catch(report);
  }, [tenantId, report]);

  useEffect(refresh, [refresh]);

  return (
    <main>
      <h1>Nuolydis</h1>
      <p>Eksperimentinė priemonė. Tai nėra gydymas. Žaidžiama tik virtualiais taškais.</p>

      <section>
        <label>
          Pora{" "}
          <select value={tenantId ?? ""} onChange={(event) => setTenantId(Number(event.target.value))}>
            {tenants.map((tenant) => (
              <option key={tenant.id} value={tenant.id}>
                {tenant.name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Veiksena{" "}
          <select value={mode} onChange={(event) => setMode(event.target.value as Mode)}>
            <option value="gambler">Lošiantis asmuo</option>
            <option value="supervisor">Prižiūrėtojas</option>
          </select>
        </label>
      </section>

      {error && <p className="error">{error}</p>}

      {tenantId !== undefined && state && (
        <>
          <ProgressPanel
            state={state}
            onAdvanceWeek={() => api.advanceWeek(tenantId).then(setState).catch(report)}
          />
          {mode === "gambler" ? (
            <GamblerView key={tenantId} tenantId={tenantId} state={state} onChange={refresh} onError={report} />
          ) : (
            <SupervisorView key={tenantId} tenantId={tenantId} onError={report} />
          )}
        </>
      )}

      <HelpContacts />
    </main>
  );
}
