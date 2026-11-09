import { useState } from "react";
import * as api from "../api/client";
import type { BetResult, Decision, State } from "../api/client";

interface Props {
  tenantId: number;
  state: State;
  onChange: () => void;
  onError: (cause: unknown) => void;
}

interface Wager {
  selection: string;
  stake: number;
}

type Pending = NonNullable<Decision["confirmation"]> & { wager: Wager };

export function GamblerView({ tenantId, state, onChange, onError }: Props) {
  const [selection, setSelection] = useState("red");
  const [stake, setStake] = useState(20);
  const [decision, setDecision] = useState<Decision>();
  const [result, setResult] = useState<BetResult>();
  const [pending, setPending] = useState<Pending>();
  const [typed, setTyped] = useState("");

  const session = state.activeSession;

  async function start() {
    try {
      const started = await api.startSession(tenantId);
      setDecision(started.decision);
      setResult(undefined);
      onChange();
    } catch (cause) {
      onError(cause);
    }
  }

  async function end() {
    if (!session) return;
    try {
      await api.endSession(tenantId, session.id);
      setPending(undefined);
      onChange();
    } catch (cause) {
      onError(cause);
    }
  }

  async function bet(wager: Wager, confirmation?: { id: number; text: string }) {
    if (!session) return;
    try {
      const placed = await api.placeBet(tenantId, session.id, {
        game: "roulette",
        selection: wager.selection,
        stake: wager.stake,
        confirmationId: confirmation?.id,
        confirmationText: confirmation?.text,
      });
      setDecision(placed.decision);
      setResult(placed);
      setTyped("");
      if (placed.decision.confirmation) {
        setPending({ ...placed.decision.confirmation, wager });
      } else if (placed.decision.outcome !== "CONFIRMATION_FAILED") {
        setPending(undefined);
      }
      onChange();
    } catch (cause) {
      onError(cause);
    }
  }

  return (
    <section>
      <h2>Ruletė</h2>
      {!session && <button onClick={start}>Pradėti sesiją</button>}

      {session && (
        <>
          <p>
            Sesija #{session.id}: statymų {session.betCount}, nuostolis {session.netLoss}, vijimosi signalų{" "}
            {session.chasingSignals}
          </p>
          <label>
            Statymas{" "}
            <input type="number" min={1} value={stake} onChange={(event) => setStake(Number(event.target.value))} />
          </label>
          <label>
            Pasirinkimas{" "}
            <input value={selection} onChange={(event) => setSelection(event.target.value)} size={6} />
          </label>
          <small>red, black arba skaičius 0–36</small>
          <p>
            <button onClick={() => bet({ selection, stake })}>Statyti</button>
            <button onClick={end}>Baigti sesiją</button>
          </p>
        </>
      )}

      {pending && (
        <div>
          <h3>Reikia patvirtinimo</h3>
          {pending.waitSeconds > 0 && <p>Prieš patvirtindami palaukite {pending.waitSeconds} s.</p>}
          {pending.kind === "CLICK" ? (
            <button onClick={() => bet(pending.wager, { id: pending.id, text: "" })}>{pending.phrase}</button>
          ) : (
            <>
              <p>Įveskite sakinį: „{pending.phrase}“</p>
              <input value={typed} onChange={(event) => setTyped(event.target.value)} size={60} />
              <button onClick={() => bet(pending.wager, { id: pending.id, text: typed })}>Patvirtinti</button>
            </>
          )}
        </div>
      )}

      {decision && (
        <div>
          <h3>Sprendimas</h3>
          <ul>
            <li>Rezultatas: {decision.outcome}</li>
            <li>Priežastis: {decision.reason}</li>
            {decision.outcome !== "DENIED" && <li>Statymas: {decision.stake}</li>}
            {decision.nextAllocationAt && (
              <li>Kitas skyrimas: {new Date(decision.nextAllocationAt).toLocaleString("lt-LT")}</li>
            )}
            {result?.round && (
              <li>
                Iškrito: {result.round.outcome}, išmoka {result.round.payout}
              </li>
            )}
            {result?.session.endReason && <li>Sesija baigta: {result.session.endReason}</li>}
          </ul>
        </div>
      )}
    </section>
  );
}
