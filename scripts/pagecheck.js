// pagecheck: syntax-check the embedded quota page JS (guards against
// parse-time SyntaxErrors like duplicate const that break the whole page).
const fs = require("fs");
const src = fs.readFileSync("plugin/quota_page.go", "utf8");
const m = src.match(/const QuotaPageHTML = `([\s\S]*)`/);
if (!m) {
  console.error("pagecheck: QuotaPageHTML not found");
  process.exit(1);
}
const js = [...m[1].matchAll(/<script>([\s\S]*?)<\/script>/g)]
  .map((x) => x[1])
  .join("\n");
if (!js.trim()) {
  console.error("pagecheck: no <script> content found");
  process.exit(1);
}
try {
  new Function(js);
} catch (e) {
  console.error("pagecheck: embedded JS SyntaxError:", e.message);
  process.exit(1);
}
console.log("pagecheck: embedded JS syntax OK");
