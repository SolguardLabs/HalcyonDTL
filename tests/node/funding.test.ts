import test from "node:test";
import assert from "node:assert/strict";
import { assertCommon, byId, events, runScenario } from "../helpers/halcyon.ts";

test("funding rates become negative when route utilization exceeds target", () => {
  const report = runScenario("funding");
  assertCommon(report, "funding");

  const route = byId(report.routes, "route-atlas-eu");
  assert.equal(route.open_interest, 1);
  assert.equal(route.utilization_bps, 8200);
  assert.equal(route.epoch_history.length, 3);
  assert.ok(route.last_funding_rate_ppm < 0);
  assert.ok(route.funding_accumulator < 0);

  const position = report.positions[0];
  assert.equal(position.status, "open");
  assert.ok(position.unrealized_funding < 0);
  assert.ok(position.margin_ratio_bps < 1200);
  assert.equal(events(report, "funding.applied").length, 9);
});

test("funding scenario leaves account snapshot synced to active route", () => {
  const report = runScenario("funding");
  const alice = byId(report.accounts, "acct-alice");
  const route = byId(report.routes, "route-atlas-eu");
  assert.equal(alice.active_route, "route-atlas-eu");
  assert.equal(alice.snapshot_route, "route-atlas-eu");
  assert.equal(alice.funding_snapshot, route.funding_accumulator);
  assert.equal(report.risk.open_positions, 1);
});

