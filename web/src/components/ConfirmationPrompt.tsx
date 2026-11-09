import { useEffect, useState } from "react";

interface Props {
  kind: string;
  phrase: string;
  waitSeconds: number;
  busy: boolean;
  onSubmit: (text: string) => void;
  onCancel: () => void;
}

export function ConfirmationPrompt({ kind, phrase, waitSeconds, busy, onSubmit, onCancel }: Props) {
  const [remaining, setRemaining] = useState(waitSeconds);
  const [typed, setTyped] = useState("");
  const locked = remaining > 0 || busy;

  useEffect(() => {
    if (remaining <= 0) return;
    const timer = setTimeout(() => setRemaining(remaining - 1), 1000);
    return () => clearTimeout(timer);
  }, [remaining]);

  const countdown = remaining > 0 ? ` (${remaining} s)` : "";
  const cancel = (
    <button className="secondary" onClick={onCancel}>
      Atšaukti
    </button>
  );

  if (kind === "CLICK") {
    return (
      <div className="notice warning">
        <button disabled={locked} onClick={() => onSubmit("")}>
          {phrase}
          {countdown}
        </button>{" "}
        {cancel}
      </div>
    );
  }

  return (
    <div className="notice warning">
      <p>Įveskite sakinį: „{phrase}“</p>
      <div className="row">
        <input value={typed} disabled={locked} onChange={(event) => setTyped(event.target.value)} size={50} />
        <button
          disabled={locked}
          onClick={() => {
            onSubmit(typed);
            setTyped("");
          }}
        >
          Patvirtinti{countdown}
        </button>
        {cancel}
      </div>
    </div>
  );
}
