import { readdirSync, readFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const src = join(root, "src");
let total = 0;

for (const file of readdirSync(src).filter((name) => name.endsWith(".go")).sort()) {
  const text = readFileSync(join(src, file), "utf8");
  const lines = text.split(/\r?\n/).filter((line) => line.trim().length > 0).length;
  total += lines;
  console.log(`${String(lines).padStart(4, " ")} ${file}`);
}

console.log(`${String(total).padStart(4, " ")} total`);
if (total < 3000 || total > 4000) {
  console.error(`src/ LOC must stay between 3000 and 4000, got ${total}`);
  process.exit(1);
}

