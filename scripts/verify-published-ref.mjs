import { execFileSync } from "node:child_process";

const run = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();
const event = process.env.VERIFY_EVENT ?? "local";
const refType = process.env.VERIFY_REF_TYPE ?? "";
const refName = process.env.VERIFY_REF_NAME ?? "";

if (refName === "production" || refType === "tag" || event === "release") {
  run("fetch", "origin", "main", "production", "--tags", "--force");
  const main = run("rev-parse", "origin/main^{commit}");
  const production = run("rev-parse", "origin/production^{commit}");
  if (main !== production) throw new Error("main and production do not identify the same commit");
  if (refType === "tag" || event === "release") {
    const tag = refName || process.env.GITHUB_REF_NAME;
    const kind = run("cat-file", "-t", `refs/tags/${tag}`);
    if (kind !== "tag") throw new Error(`${tag} must be an annotated tag`);
    const tagged = run("rev-parse", `refs/tags/${tag}^{commit}`);
    if (tagged !== main) throw new Error(`${tag} does not identify the production commit`);
  }
}

console.log(JSON.stringify({ event, refType, refName, commit: run("rev-parse", "HEAD") }));
