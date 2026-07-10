import test from "node:test";
import assert from "node:assert/strict";
import { assertCommon, bucket, byId, events, runScenario } from "../helpers/halcyon.ts";

test("liquidation closes an under-maintenance position", () => {
  const report = runScenario("liquidation");
  assertCommon(report, "liquidation");
  assert.equal(events(report, "position.liquidated").length, 1);
  assert.equal(report.risk.liquidated_positions, 1);
  assert.equal(report.risk.open_positions, 0);

  const position = report.positions[0];
  assert.equal(position.status, "liquidated");
  assert.equal(position.close_reason, "liquidation");
  assert.ok(position.funding_paid < 0);
  assert.ok(position.liquidation_fee > 0);
  assert.ok(position.socialized_debt > 0);
});

test("liquidation pushes residual debt into the pool and frees route utilization", () => {
  const report = runScenario("liquidation");
  const route = byId(report.routes, "route-atlas-eu");
  assert.equal(route.utilized, 0);
  assert.ok(bucket(report.pool.socialized_debt, "hUSD") > 0);
  assert.ok(bucket(report.pool.fees_collected, "hUSD") > 0);

  const bob = byId(report.accounts, "acct-bob");
  assert.equal(bob.status, "restricted");
  assert.equal(bob.positions.length, 0);
});

