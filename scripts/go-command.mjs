import { spawnSync } from "node:child_process";

const go = process.env.GO_BIN ?? "go";
const result = spawnSync(go, process.argv.slice(2), { encoding: "utf8", stdio: "inherit", shell: false });
if (result.error) {
  console.error(`Unable to execute ${go}: ${result.error.message}`);
  process.exit(1);
}
process.exit(result.status ?? 1);
