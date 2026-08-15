import assert from "node:assert/strict";
import test from "node:test";
import { HalcyonClient, HalcyonClientError, assertScenarioName, validateReport } from "../../sdk/index.js";
import { binaryPath, ensureBuilt } from "../helpers/halcyon.ts";

test("client lists and validates deterministic scenarios", () => {
  ensureBuilt();
  const client = new HalcyonClient({ binary: binaryPath() });
  assert.deepEqual(client.listScenarios(), ["funding", "liquidation", "open-close", "rotation", "routes"]);
  assert.equal(client.validate("funding"), "ok funding");
});

test("client returns an immutable validated report", () => {
  ensureBuilt();
  const report = new HalcyonClient({ binary: binaryPath() }).runScenario("routes");
  assert.equal(report.protocol, "HalcyonDTL");
  assert.equal(Object.isFrozen(report), true);
  assert.equal(Object.isFrozen(report.routes), true);
});

test("client rejects malformed scenario names before execution", () => {
  for (const value of ["", "../routes", "Routes", "funding;exit", "a".repeat(49)]) {
    assert.throws(() => assertScenarioName(value), TypeError);
  }
});

test("client validates timeout and buffer bounds", () => {
  assert.throws(() => new HalcyonClient({ timeout: 0 }), TypeError);
  assert.throws(() => new HalcyonClient({ maxBuffer: 512 }), TypeError);
});

test("client reports missing binaries without invoking a shell", () => {
  const client = new HalcyonClient({ binary: "Z:/absent/halcyondtl" });
  assert.throws(() => client.listScenarios(), HalcyonClientError);
});

test("report validation rejects identity and invariant mismatches", () => {
  const base = {
    protocol: "HalcyonDTL",
    scenario: "funding",
    state_digest: "a".repeat(32),
    routes: [],
    positions: [],
    events: [],
    invariants: { accounting: true },
  };
  assert.throws(() => validateReport({ ...base, protocol: "other" }, "funding"), HalcyonClientError);
  assert.throws(
    () => validateReport({ ...base, invariants: { accounting: false } }, "funding"),
    HalcyonClientError,
  );
});
