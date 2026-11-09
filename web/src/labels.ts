const outcomeLabels: Record<string, string> = {
  ACCEPTED: "Statymas priimtas",
  ACCEPTED_CAPPED: "Statymas sumažintas",
  REJECTED: "Statymas atmestas",
  DENIED: "Sesija nepradėta",
  CONFIRMATION_REQUIRED: "Reikia patvirtinimo",
  CONFIRMATION_FAILED: "Patvirtinimas nepavyko",
};

const alertLabels: Record<string, string> = {
  CHASING_SIGNAL: "Vijimosi signalas",
  CHASING_LIMIT: "Žaidimas sustabdytas iki kito pirmadienio dėl vijimosi",
  RAPID_BETTING: "Per greiti statymai",
};

export function outcomeLabel(outcome: string): string {
  return outcomeLabels[outcome] ?? outcome;
}

export function alertLabel(reason: string): string {
  return alertLabels[reason] ?? reason;
}

export function outcomeTone(outcome: string): string {
  if (outcome === "REJECTED" || outcome === "DENIED" || outcome === "CONFIRMATION_FAILED") return "danger";
  if (outcome === "ACCEPTED_CAPPED" || outcome === "CONFIRMATION_REQUIRED") return "warning";
  return "";
}
