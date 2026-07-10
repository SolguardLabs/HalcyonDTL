import test from "node:test";
import assert from "node:assert/strict";
import { assertCommon, bucket, byId, events, runScenario, signedBucket } from "../helpers/halcyon.ts";

test("normal close settles accumulated negative funding and releases route utilization", () => {
  const report = runScenario("open-close");
  assertCommon(report, "open-close");
  assert.equal(events(report, "position.closed").length, 1);
  assert.equal(report.risk.open_positions, 0);
  assert.equal(report.risk.closed_positions, 1);

  const route = byId(report.routes, "route-atlas-eu");
  assert.equal(route.utilized, 0);
  assert.equal(route.open_interest, 0);

  const position = report.positions[0];
  assert.equal(position.status, "closed");
  assert.equal(position.close_reason, "normal_close");
  assert.ok(position.funding_paid < 0);
  assert.equal(position.close_fee, 608);
});

test("normal close records user funding debit and protocol fee", () => {
  const report = runScenario("open-close");
  const alice = byId(report.accounts, "acct-alice");
  assert.ok(signedBucket(alice.realized_funding, "hUSD") < 0);
  assert.equal(bucket(alice.fees_paid, "hUSD"), 608);
  assert.equal(bucket(report.pool.fees_collected, "hUSD"), 608);
  assert.equal(report.pool.uncollected_funding.length, 0);
});

