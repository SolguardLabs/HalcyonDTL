import test from "node:test";
import assert from "node:assert/strict";
import { assertCommon, byId, events, runScenario } from "../helpers/halcyon.ts";

test("routes scenario opens multiple routes and records benign rotation", () => {
  const report = runScenario("routes");
  assertCommon(report, "routes");
  assert.equal(events(report, "route.opened").length, 3);
  assert.equal(events(report, "position.opened").length, 3);
  assert.equal(events(report, "account.route_rotated").length, 1);

  const atlas = byId(report.routes, "route-atlas-eu");
  const boreal = byId(report.routes, "route-boreal-us");
  const cirrus = byId(report.routes, "route-cirrus-apac");
  assert.ok(atlas.utilized > 0);
  assert.ok(boreal.utilized > 0);
  assert.ok(cirrus.utilized > 0);
  assert.equal(atlas.status, "open");
});

test("route drift report highlights exposure whose account cursor points elsewhere", () => {
  const report = runScenario("routes");
  const alice = byId(report.accounts, "acct-alice");
  assert.equal(alice.active_route, "route-boreal-us");
  assert.ok(report.route_drift.some((drift) => drift.account_id === "acct-alice" && drift.route_id === "route-atlas-eu"));
});

