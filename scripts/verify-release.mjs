import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { readFileSync, readdirSync, statSync } from "node:fs";
import { extname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const expectedBlobs = new Map([
  ["src/funding.go", "15c9d1d1bf0e28d078c197b54e79bd5836a0cb65"],
  ["src/engine.go", "ae1d1af32128b4ae79ba7fb0556ba775656aa49c"],
  ["src/account.go", "0a048511e395e0f0e95e0abbd1f04621b4aac7a8"],
  ["src/rotation.go", "884ae480fbf67c42308e68d42feae8cd3948b19c"],
  ["src/position.go", "b004565f5da7de0e0ba83aac015863b6525e2fc8"],
]);
const errors = [];
const check = (condition, message) => {
  if (!condition) errors.push(message);
};
const git = (...args) => execFileSync("git", args, { cwd: root, encoding: "utf8" }).trim();

for (const [file, expected] of expectedBlobs) {
  check(git("hash-object", file) === expected, `${file} does not match its approved source blob`);
}
const docs = readdirSync(join(root, "docs"))
  .filter((name) => name.endsWith(".md"))
  .sort();
check(docs.length === 7, `expected 7 operational documents, found ${docs.length}`);
const markdown = ["README.md", "SECURITY.md", ...docs.map((name) => `docs/${name}`)];
let diagrams = 0;
for (const file of markdown)
  diagrams += (readFileSync(join(root, file), "utf8").match(/```mermaid\b/g) ?? []).length;
check(diagrams === 27, `expected 27 Mermaid diagrams, found ${diagrams}`);

const banner = readFileSync(join(root, "assets", "halcyondtl-banner.png"));
check(banner.subarray(1, 4).toString() === "PNG", "banner is not a PNG image");
check(
  banner.readUInt32BE(16) === 1672 && banner.readUInt32BE(20) === 941,
  "banner dimensions must be 1672x941",
);

const ignored = new Set([".git", "node_modules", "out", "coverage"]);
const textExtensions = new Set([".go", ".js", ".mjs", ".ts", ".md", ".json", ".yml", ".yaml", ".mod", ""]);
const restricted =
  /\b(?:ctf|labs?|laboratorios?|vulnerabil(?:ity|idad|idades)|vulnerable|bugs?|exploits?|bypass|attackers?|atacantes?)\b/i;
function walk(directory) {
  for (const name of readdirSync(directory)) {
    if (ignored.has(name)) continue;
    const full = join(directory, name);
    const info = statSync(full);
    if (info.isDirectory()) walk(full);
    else if (
      !new Set(["LICENSE", "scripts/verify-release.mjs", ...expectedBlobs.keys()]).has(
        relative(root, full).replaceAll("\\", "/"),
      ) &&
      textExtensions.has(extname(name)) &&
      restricted.test(readFileSync(full, "utf8"))
    )
      errors.push(`${relative(root, full)} contains restricted public terminology`);
  }
}
walk(root);
const packageJson = JSON.parse(readFileSync(join(root, "package.json"), "utf8"));
check(packageJson.version === "1.0.0", "package version must be 1.0.0");
check(git("check-ignore", "src/private_proof_test.go").length > 0, "private proof path must remain ignored");
const digest = createHash("sha256")
  .update(markdown.map((file) => readFileSync(join(root, file))).join("\n"))
  .digest("hex");
if (errors.length) {
  for (const error of errors) console.error(`- ${error}`);
  process.exit(1);
}
console.log(
  JSON.stringify(
    {
      protocol: "HalcyonDTL",
      version: packageJson.version,
      docs: docs.length,
      diagrams,
      protected_sources: expectedBlobs.size,
      documentation_sha256: digest,
    },
    null,
    2,
  ),
);
