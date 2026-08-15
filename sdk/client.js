import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const executable = process.platform === "win32" ? "halcyondtl.exe" : "halcyondtl";

export class HalcyonClientError extends Error {
  constructor(message, details = {}) {
    super(message);
    this.name = "HalcyonClientError";
    this.details = Object.freeze({ ...details });
  }
}

export class HalcyonClient {
  #binary;
  #timeout;
  #maxBuffer;

  constructor(options = {}) {
    this.#binary = options.binary ?? process.env.HALCYON_BIN ?? join(root, "out", executable);
    this.#timeout = options.timeout ?? 5_000;
    this.#maxBuffer = options.maxBuffer ?? 2 * 1024 * 1024;
    if (!Number.isSafeInteger(this.#timeout) || this.#timeout < 1 || this.#timeout > 60_000) {
      throw new TypeError("timeout must be an integer between 1 and 60000 milliseconds");
    }
    if (!Number.isSafeInteger(this.#maxBuffer) || this.#maxBuffer < 1024) {
      throw new TypeError("maxBuffer must be an integer of at least 1024 bytes");
    }
  }

  get binary() {
    return this.#binary;
  }

  listScenarios() {
    return this.#execute(["--list"]).trim().split(/\r?\n/).filter(Boolean);
  }

  runScenario(name) {
    assertScenarioName(name);
    const output = this.#execute(["scenario", name]);
    let report;
    try {
      report = JSON.parse(output);
    } catch (cause) {
      throw new HalcyonClientError("engine returned malformed JSON", { cause: String(cause) });
    }
    return validateReport(report, name);
  }

  validate(name) {
    assertScenarioName(name);
    return this.#execute(["validate", name]).trim();
  }

  #execute(args) {
    if (!existsSync(this.#binary)) {
      throw new HalcyonClientError("engine binary was not found", { binary: this.#binary });
    }
    const result = spawnSync(this.#binary, args, {
      cwd: root,
      encoding: "utf8",
      shell: false,
      timeout: this.#timeout,
      maxBuffer: this.#maxBuffer,
      windowsHide: true,
    });
    if (result.error) {
      throw new HalcyonClientError("engine execution failed", {
        code: result.error.code,
        message: result.error.message,
      });
    }
    if (result.status !== 0) {
      throw new HalcyonClientError("engine rejected the request", {
        status: result.status,
        signal: result.signal,
        stderr: String(result.stderr ?? "")
          .trim()
          .slice(0, 4096),
      });
    }
    return String(result.stdout ?? "");
  }
}

export function assertScenarioName(value) {
  if (typeof value !== "string" || !/^[a-z][a-z0-9-]{0,47}$/.test(value)) {
    throw new TypeError("scenario name has an invalid format");
  }
}

export function validateReport(report, scenario) {
  if (!report || typeof report !== "object" || Array.isArray(report)) {
    throw new HalcyonClientError("engine report must be an object");
  }
  if (report.protocol !== "HalcyonDTL" || report.scenario !== scenario) {
    throw new HalcyonClientError("engine report identity mismatch");
  }
  if (!/^[0-9a-f]{32}$/.test(report.state_digest ?? "")) {
    throw new HalcyonClientError("engine report digest is invalid");
  }
  if (!Array.isArray(report.routes) || !Array.isArray(report.positions) || !Array.isArray(report.events)) {
    throw new HalcyonClientError("engine report collections are invalid");
  }
  if (!report.invariants || Object.values(report.invariants).some((value) => value !== true)) {
    throw new HalcyonClientError("engine report contains a failed invariant");
  }
  return deepFreeze(report);
}

function deepFreeze(value) {
  if (value && typeof value === "object" && !Object.isFrozen(value)) {
    for (const child of Object.values(value)) deepFreeze(child);
    Object.freeze(value);
  }
  return value;
}
