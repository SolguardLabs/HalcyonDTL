import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const root = resolve(fileURLToPath(new URL("../..", import.meta.url)));
const exeName = process.platform === "win32" ? "halcyondtl.exe" : "halcyondtl";
const defaultBin = join(root, "out", exeName);

export type AmountBucket = {
  asset: string;
  amount: number;
};

export type SignedBucket = {
  asset: string;
  amount: number;
};

export type FundingRecord = {
  epoch: number;
  route_id: string;
  asset: string;
  utilization_bps: number;
  target_bps: number;
  rate_ppm: number;
  accumulator_before: number;
  accumulator_after: number;
  open_interest: number;
  utilized: number;
  liquidity: number;
};

export type RouteReport = {
  id: string;
  source_vault: string;
  destination_vault: string;
  operator: string;
  asset: string;
  liquidity: number;
  utilized: number;
  reserved: number;
  capacity: number;
  utilization_bps: number;
  target_utilization_bps: number;
  max_utilization_bps: number;
  last_funding_rate_ppm: number;
  funding_accumulator: number;
  last_epoch: number;
  open_interest: number;
  status: string;
  epoch_history: FundingRecord[];
};

export type AccountReport = {
  id: string;
  owner: string;
  collateral: AmountBucket[];
  reserved_margin: AmountBucket[];
  realized_funding: SignedBucket[];
  fees_paid: AmountBucket[];
  positions: string[];
  active_route: string;
  funding_snapshot: number;
  snapshot_epoch: number;
  snapshot_route: string;
  status: string;
};

export type PositionReport = {
  id: string;
  account_id: string;
  route_id: string;
  asset: string;
  notional: number;
  margin: number;
  entry_accumulator: number;
  exit_accumulator: number;
  open_epoch: number;
  close_epoch: number;
  funding_paid: number;
  close_fee: number;
  liquidation_fee: number;
  socialized_debt: number;
  status: string;
  close_reason: string;
  unrealized_funding: number;
  margin_ratio_bps: number;
};

export type HalcyonReport = {
  lab: string;
  scenario: string;
  network_id: string;
  clock: number;
  state_digest: string;
  assets: Array<Record<string, unknown>>;
  operators: Array<Record<string, unknown>>;
  vaults: Array<Record<string, any>>;
  accounts: AccountReport[];
  routes: RouteReport[];
  positions: PositionReport[];
  funding: {
    records: FundingRecord[];
    last_epoch: number;
    negative_notional: number;
    positive_notional: number;
  };
  pool: {
    fees_collected: AmountBucket[];
    insurance_balance: AmountBucket[];
    socialized_debt: AmountBucket[];
    uncollected_funding: AmountBucket[];
  };
  route_drift: Array<Record<string, any>>;
  totals: Record<string, AmountBucket[]>;
  risk: Record<string, number>;
  invariants: Record<string, boolean>;
  events: Array<Record<string, any>>;
  notes: string[];
};

export function binaryPath(): string {
  return process.env.HALCYON_BIN ?? defaultBin;
}

export function ensureBuilt(): void {
  if (process.env.HALCYON_BIN) return;
  if (existsSync(defaultBin)) return;
  const result = spawnSync(process.execPath, ["scripts/build.mjs"], {
    cwd: root,
    encoding: "utf8",
    stdio: "pipe",
  });
  if (result.status !== 0) {
    throw new Error(["build failed", result.stdout.trim(), result.stderr.trim()].filter(Boolean).join("\n"));
  }
}

export function runBinary(args: string[], expectSuccess = true) {
  ensureBuilt();
  const result = spawnSync(binaryPath(), args, {
    cwd: root,
    encoding: "utf8",
    stdio: "pipe",
  });
  if (expectSuccess && result.status !== 0) {
    throw new Error(
      [`halcyondtl ${args.join(" ")} failed`, result.stdout.trim(), result.stderr.trim()]
        .filter(Boolean)
        .join("\n"),
    );
  }
  return result;
}

export function listScenarios(): string[] {
  return runBinary(["--list"]).stdout.trim().split(/\r?\n/).filter(Boolean);
}

export function runScenario(name: string): HalcyonReport {
  const result = runBinary(["scenario", name]);
  return JSON.parse(result.stdout) as HalcyonReport;
}

export function validateScenario(name: string): string {
  return runBinary(["validate", name]).stdout.trim();
}

export function byId<T extends { id: string }>(items: T[], id: string): T {
  const found = items.find((item) => item.id === id);
  assert.ok(found, `missing id ${id}`);
  return found;
}

export function bucket(entries: AmountBucket[], asset: string): number {
  const found = entries.find((entry) => entry.asset === asset);
  assert.ok(found, `missing asset ${asset}`);
  return found.amount;
}

export function signedBucket(entries: SignedBucket[], asset: string): number {
  const found = entries.find((entry) => entry.asset === asset);
  assert.ok(found, `missing signed asset ${asset}`);
  return found.amount;
}

export function events(report: HalcyonReport, kind: string): Array<Record<string, any>> {
  return report.events.filter((event) => event.kind === kind);
}

export function assertDigest(value: unknown): void {
  assert.equal(typeof value, "string");
  assert.match(value as string, /^[0-9a-f]{32}$/);
}

export function assertCommon(report: HalcyonReport, scenario: string): void {
  assert.equal(report.lab, "HalcyonDTL");
  assert.equal(report.scenario, scenario);
  assert.equal(report.network_id, "halcyon-local-funding");
  assertDigest(report.state_digest);
  assert.ok(report.assets.length >= 1);
  assert.ok(report.operators.length >= 3);
  assert.ok(report.vaults.length >= 6);
  assert.ok(report.routes.length >= 3);
  assert.ok(Array.isArray(report.events));
  assert.equal(report.invariants.accounts_non_negative, true);
  assert.equal(report.invariants.vaults_non_negative, true);
  assert.equal(report.invariants.routes_within_max_utilization, true);
  assert.equal(report.invariants.positions_linked, true);
  assert.equal(report.invariants.funding_records_link_routes, true);
  assert.equal(report.invariants.active_snapshots_link_open_routes, true);
  assert.equal(report.invariants.closed_positions_not_in_accounts, true);
}

