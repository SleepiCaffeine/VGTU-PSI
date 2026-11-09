import createClient from "openapi-fetch";
import type { components, paths } from "./schema";

export type Tenant = components["schemas"]["Tenant"];
export type State = components["schemas"]["State"];
export type Decision = components["schemas"]["Decision"];
export type BetRequest = components["schemas"]["BetRequest"];
export type BetResult = components["schemas"]["BetResult"];
export type SessionStartResult = components["schemas"]["SessionStartResult"];
export type Alert = components["schemas"]["Alert"];

const client = createClient<paths>({ baseUrl: "/api" });

function tenantOptions(tenantId: number) {
  return {
    params: { path: { tenantId } },
    headers: { "X-Tenant-ID": String(tenantId) },
  };
}

function failure(error: unknown): Error {
  const message = (error as { error?: string } | undefined)?.error;
  return new Error(message ?? "Užklausa nepavyko");
}

export async function listTenants(): Promise<Tenant[]> {
  const { data, error } = await client.GET("/tenants");
  if (!data) throw failure(error);
  return data;
}

export async function getState(tenantId: number): Promise<State> {
  const { data, error } = await client.GET("/tenants/{tenantId}/state", tenantOptions(tenantId));
  if (!data) throw failure(error);
  return data;
}

export async function listAlerts(tenantId: number): Promise<Alert[]> {
  const { data, error } = await client.GET("/tenants/{tenantId}/alerts", tenantOptions(tenantId));
  if (!data) throw failure(error);
  return data;
}

export async function advanceWeek(tenantId: number): Promise<State> {
  const { data, error } = await client.POST("/tenants/{tenantId}/demo/advance-week", tenantOptions(tenantId));
  if (!data) throw failure(error);
  return data;
}

export async function startSession(tenantId: number): Promise<SessionStartResult> {
  const { data, error } = await client.POST("/tenants/{tenantId}/sessions", tenantOptions(tenantId));
  if (data) return data;
  if (error && "decision" in error) return error;
  throw failure(error);
}

export async function endSession(tenantId: number, sessionId: number): Promise<void> {
  const { error } = await client.POST("/tenants/{tenantId}/sessions/{sessionId}/end", {
    params: { path: { tenantId, sessionId } },
    headers: { "X-Tenant-ID": String(tenantId) },
  });
  if (error) throw failure(error);
}

export async function placeBet(tenantId: number, sessionId: number, body: BetRequest): Promise<BetResult> {
  const { data, error } = await client.POST("/tenants/{tenantId}/sessions/{sessionId}/bets", {
    params: { path: { tenantId, sessionId } },
    headers: { "X-Tenant-ID": String(tenantId) },
    body,
  });
  if (data) return data;
  if (error && "decision" in error) return error;
  throw failure(error);
}
