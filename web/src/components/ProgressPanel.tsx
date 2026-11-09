import type { State } from "../api/client";

interface Props {
  state: State;
  onAdvanceWeek: () => void;
}

export function ProgressPanel({ state, onAdvanceWeek }: Props) {
  return (
    <section>
      <h2>Pažanga</h2>
      <ul>
        <li>Plano savaitė: {state.weekNumber}</li>
        <li>
          Savaitinis skyrimas: {state.allocation} (pradžioje {state.initialAllocation})
        </li>
        <li>Skyrimas sumažėjo: {state.progressPercent} %</li>
        <li>Balansas: {state.balance}</li>
        <li>Trinties lygis: {state.frictionLevel}</li>
        <li>Kitas skyrimas: {new Date(state.nextAllocationAt).toLocaleString("lt-LT")}</li>
      </ul>
      <button onClick={onAdvanceWeek}>Demonstracija: pereiti į kitą savaitę</button>
    </section>
  );
}
