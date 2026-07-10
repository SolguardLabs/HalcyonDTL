import test from "node:test";
import assert from "node:assert/strict";
import { listScenarios, runBinary, validateScenario } from "../helpers/halcyon.ts";

test("cli lists deterministic public scenarios", () => {
  assert.deepEqual(listScenarios(), ["funding", "liquidation", "open-close", "rotation", "routes"]);
});

test("cli rejects unknown scenarios", () => {
  const result = runBinary(["scenario", "missing"], false);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /unknown scenario/);
});

test("all public scenarios validate invariants", () => {
  for (const scenario of listScenarios()) {
    assert.equal(validateScenario(scenario), `ok ${scenario}`);
  }
});

