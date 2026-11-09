import { useRef, useState } from "react";
import * as api from "../api/client";
import type { BetResult, Decision, State } from "../api/client";
import { outcomeLabel, outcomeTone } from "../labels";
import { ConfirmationPrompt } from "./ConfirmationPrompt";

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
  const [busy, setBusy] = useState(false);
  const betInFlight = useRef(false);

  const session = state.activeSession;
  const pickedNumber = selection !== "red" && selection !== "black";

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
    if (!session || betInFlight.current) return;
    betInFlight.current = true;
    setBusy(true);
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
      if (placed.decision.confirmation) {
        setPending({ ...placed.decision.confirmation, wager });
      } else if (placed.decision.outcome !== "CONFIRMATION_FAILED") {
        setPending(undefined);
      }
      onChange();
    } catch (cause) {
      onError(cause);
    } finally {
      betInFlight.current = false;
      setBusy(false);
    }
  }

  return (
    <section>
      <h2>Ruletė</h2>
      {!session && <button onClick={start}>Pradėti sesiją</button>}

      {session && (
        <>
          <div className="tiles">
            <div className="tile">
              <strong>{session.betCount}</strong>
              <small>Statymai</small>
            </div>
            <div className="tile">
              <strong>{session.netLoss}</strong>
              <small>Nuostolis</small>
            </div>
            <div className="tile">
              <strong>{state.chasingSignalsThisWeek} / 3</strong>
              <small>Vijimosi signalai šią savaitę</small>
            </div>
          </div>
          <div className="row">
            <button className={`red ${selection === "red" ? "selected" : ""}`} onClick={() => setSelection("red")}>
              Raudona
            </button>
            <button className={`black ${selection === "black" ? "selected" : ""}`} onClick={() => setSelection("black")}>
              Juoda
            </button>
            <input
              type="number"
              min={0}
              max={36}
              placeholder="0–36"
              aria-label="Skaičius"
              value={pickedNumber ? selection : ""}
              onChange={(event) => setSelection(event.target.value || "red")}
            />
          </div>
          <div className="row">
            <input
              type="number"
              min={1}
              aria-label="Statymas"
              value={stake}
              onChange={(event) => setStake(Number(event.target.value))}
            />
            <button disabled={busy || pending !== undefined} onClick={() => bet({ selection, stake })}>
              Statyti
            </button>
            <button className="secondary" onClick={end}>
              Baigti sesiją
            </button>
          </div>
        </>
      )}

      {pending && (
        <ConfirmationPrompt
          key={pending.id}
          kind={pending.kind}
          phrase={pending.phrase}
          waitSeconds={pending.waitSeconds}
          busy={busy}
          onCancel={() => setPending(undefined)}
          onSubmit={(text) => bet(pending.wager, { id: pending.id, text })}
        />
      )}

      {decision && (
        <div className={`notice ${outcomeTone(decision.outcome)}`}>
          <strong>{outcomeLabel(decision.outcome)}</strong>
          {decision.outcome !== "DENIED" && decision.stake > 0 && <> · statymas {decision.stake}</>}
          {result?.round && (
            <>
              {" "}
              · iškrito {result.round.outcome}, išmoka {result.round.payout}
            </>
          )}
          <small>
            {decision.outcome} · {decision.reason}
            {result?.session.endReason && ` · sesija baigta: ${result.session.endReason}`}
            {decision.nextAllocationAt &&
              ` · kitas skyrimas ${new Date(decision.nextAllocationAt).toLocaleDateString("lt-LT")}`}
          </small>
        </div>
      )}
    </section>
  );
}
