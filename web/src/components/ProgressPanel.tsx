import type { State } from "../api/client";

const frictionLevels = [1, 2, 3, 4];

interface Props {
  state: State;
  onAdvanceWeek: () => void;
}

export function ProgressPanel({ state, onAdvanceWeek }: Props) {
  const nextAllocation = new Date(state.nextAllocationAt).toLocaleDateString("lt-LT");

  return (
    <section>
      <h2>Pažanga</h2>
      <div className="tiles">
        <Tile label="Balansas" value={state.balance} />
        <Tile label="Skyrimas" value={state.allocation} />
        <Tile label="Sumažėjo" value={`${state.progressPercent} %`} />
        <Tile label="Savaitė" value={state.weekNumber} />
      </div>
      <div className="bar">
        <div style={{ width: `${state.progressPercent}%` }} />
      </div>
      <small>
        Kelias nuo {state.initialAllocation} taškų iki nulio · kitas skyrimas {nextAllocation}
      </small>
      <div className="row">
        <span>Trintis</span>
        <span className="pips" aria-label={`Trinties lygis ${state.frictionLevel} iš 4`}>
          {frictionLevels.map((level) => (
            <span key={level} className={level <= state.frictionLevel ? "on" : ""} />
          ))}
        </span>
        <button className="secondary" onClick={onAdvanceWeek}>
          Demonstracija: kita savaitė
        </button>
      </div>
    </section>
  );
}

function Tile({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="tile">
      <strong>{value}</strong>
      <small>{label}</small>
    </div>
  );
}
