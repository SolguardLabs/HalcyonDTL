import { spawnSync } from "node:child_process";
import { readdirSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const files = readdirSync(join(root, "src"))
  .filter((name) => name.endsWith(".go"))
  .sort()
  .map((name) => join("src", name));
const mode = process.argv[2];
const gofmt = process.env.GOFMT_BIN ?? "gofmt";
if (!["--check", "--write"].includes(mode)) throw new Error("expected --check or --write");
const args = mode === "--write" ? ["-w", ...files] : ["-d", ...files];
const result = spawnSync(gofmt, args, { cwd: root, encoding: "utf8", shell: false });
if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
if (mode === "--check" && result.stdout) {
  process.stdout.write(result.stdout);
  process.exit(1);
}
if (result.stderr) process.stderr.write(result.stderr);
process.exit(result.status ?? 0);
