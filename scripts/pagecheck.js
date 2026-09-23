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

// --- Undeclared-identifier audit (guards against ReferenceErrors like the
// v0.4.4 `monthlyTargetTime is not defined` bug: a bare identifier read in
// updateTimers() that was never declared and only existed as a global
// property accidentally created by renderUsage()).
//
// Approach (deliberately simple/grep-style, no DOM execution):
//   1. strip comments and string literals
//   2. collect every var/let/const/function declaration name + function/
//      callback/catch parameter
//   3. flag candidates: bare assignment targets (x =, x +=, x++, ...),
//      bare if()/while() condition identifiers, and for-loop init identifiers
//   4. whitelist known globals; anything left is reported and fails the check

const whitelist = new Set([
  // browser builtins referenced by the page
  "document", "window", "localStorage", "sessionStorage", "fetch",
  "setInterval", "setTimeout", "clearInterval", "clearTimeout", "console",
  "alert", "Date", "Math", "Number", "String", "Boolean", "Array",
  "Object", "JSON", "parseInt", "parseFloat", "isNaN", "Promise", "Error",
  "escape", "unescape", "navigator", "location", "history", "URL",
  "URLSearchParams", "FormData", "Headers", "Request", "Response",
  "Intl", "Map", "Set", "AbortController", "requestAnimationFrame",
  "cancelAnimationFrame", "structuredClone", "globalThis", "arguments",
]);

function stripLiterals(source) {
  // Remove comments, string literals, and regex literals; replace with
  // harmless placeholders so identifier scanning never sees string content.
  // Template literals are NOT fully discarded: their ${...} interpolations
  // are kept as "( ... )" so identifiers inside them stay visible.
  //
  // A "/" only starts a regex literal when it appears in expression
  // position (previous significant token is an operator/open bracket or a
  // keyword such as return/typeof). Otherwise it is division.
  let out = "";
  let i = 0;
  const n = source.length;
  // Stack of lexer contexts. Each entry: { type: "expr", depth: number } or
  // { type: "tmpl" }. The initial code runs in a never-ending expr context.
  const stack = [{ type: "expr", depth: Infinity }];
  // Last significant (non-whitespace) emitted char + last identifier word,
  // used for the regex-vs-division heuristic.
  let lastSig = "";
  let lastWord = "";
  const KEYWORDS_BEFORE_REGEX = new Set([
    "return", "typeof", "instanceof", "in", "of", "new", "delete", "void",
    "do", "else", "case", "throw", "await", "yield",
  ]);
  const emit = (text) => {
    for (const ch of text) {
      if (/\s/.test(ch)) continue;
      if (/[A-Za-z0-9_$]/.test(ch)) {
        lastWord = /[A-Za-z0-9_$]/.test(lastSig) ? lastWord + ch : ch;
      } else {
        lastWord = "";
      }
      lastSig = ch;
    }
    out += text;
  };
  const regexAllowed = () =>
    lastSig === "" ||
    "(,=:[!&|?+-*/%<>~^;{".includes(lastSig) ||
    KEYWORDS_BEFORE_REGEX.has(lastWord);

  while (i < n) {
    const ctx = stack[stack.length - 1];
    const c = source[i];
    const next = source[i + 1];

    if (ctx.type === "tmpl") {
      // Inside template literal text: skip until ` or ${ ... }
      if (c === "\\") { i += 2; continue; }
      if (c === "`") { i++; stack.pop(); emit(" "); continue; }
      if (c === "$" && next === "{") {
        i += 2;
        stack.push({ type: "expr", depth: 0 });
        emit(" ( ");
        continue;
      }
      i++;
      continue;
    }

    // code context (top-level or template ${...} expression)
    if (c === "/" && next === "/") {
      while (i < n && source[i] !== "\n") i++;
    } else if (c === "/" && next === "*") {
      i += 2;
      while (i < n && !(source[i] === "*" && source[i + 1] === "/")) i++;
      i += 2;
    } else if (c === "/" && regexAllowed()) {
      // regex literal: skip to unescaped closing / (not inside [...])
      i++;
      let inClass = false;
      while (i < n) {
        if (source[i] === "\\") { i += 2; continue; }
        if (source[i] === "[") { inClass = true; i++; continue; }
        if (source[i] === "]") { inClass = false; i++; continue; }
        if (source[i] === "/" && !inClass) { i++; break; }
        if (source[i] === "\n") break; // malformed; bail out safely
        i++;
      }
      while (i < n && /[a-z]/i.test(source[i])) i++; // flags
      emit(" / ");
    } else if (c === '"' || c === "'") {
      const quote = c;
      i++;
      while (i < n) {
        if (source[i] === "\\") { i += 2; continue; }
        if (source[i] === quote) { i++; break; }
        if (source[i] === "\n") break;
        i++;
      }
      emit(" " + quote + quote + " ");
    } else if (c === "`") {
      i++;
      stack.push({ type: "tmpl" });
      emit(" ");
    } else {
      if (c === "{") ctx.depth++;
      if (c === "}") {
        if (ctx.depth === 0) {
          // closes a template interpolation: back into template text
          stack.pop();
          i++;
          emit(" ) ");
          continue;
        }
        ctx.depth--;
      }
      emit(c);
      i++;
    }
  }
  return out;
}

function collectDeclarations(clean) {
  const declared = new Set();
  const addParamList = (raw) => {
    for (const p of raw.split(",")) {
      const name = p.trim().split(/[\s=]/)[0].replace(/^\.\.\./, "");
      if (/^[A-Za-z_$][\w$]*$/.test(name)) declared.add(name);
    }
  };
  for (const m of clean.matchAll(/\b(?:var|let|const)\s+([A-Za-z_$][\w$]*)/g))
    declared.add(m[1]);
  // function declarations/expressions: name + params
  for (const m of clean.matchAll(/\bfunction\s*([A-Za-z_$][\w$]*)?\s*\(([^()]*)\)/g)) {
    if (m[1]) declared.add(m[1]);
    addParamList(m[2]);
  }
  // arrow functions: (a, b) => and a =>
  for (const m of clean.matchAll(/\(\s*([^()]*?)\s*\)\s*=>/g)) addParamList(m[1]);
  for (const m of clean.matchAll(/(?<![\w$.(])\b([A-Za-z_$][\w$]*)\s*=>/g)) declared.add(m[1]);
  // catch (e) and destructuring catch
  for (const m of clean.matchAll(/\bcatch\s*\(?\s*\{?\s*([A-Za-z_$][\w$]*)/g)) declared.add(m[1]);
  return declared;
}

const clean = stripLiterals(js);
const declared = collectDeclarations(clean);
const flagged = new Set();

// 1. bare assignment targets / updates:  x =, x +=, x++, x--, x ??=
for (const m of clean.matchAll(/(?:^|[{};\n])\s*([A-Za-z_$][\w$]*)\s*(?:=[^=>]|[+*\/%-]?=[^=]|\+\+|\-\-)/gm)) {
  const name = m[1];
  if (!declared.has(name) && !whitelist.has(name)) flagged.add(name);
}
// 2. bare if()/while() condition identifiers
for (const m of clean.matchAll(/\b(?:if|while)\s*\(\s*(!*)\s*([A-Za-z_$][\w$]*)\s*(?:\)|&&|\|\||\?)/g)) {
  const name = m[2];
  if (!declared.has(name) && !whitelist.has(name)) flagged.add(name);
}
// 3. for-loop init without let/var: for (i = 0; ...)
for (const m of clean.matchAll(/\bfor\s*\(\s*([A-Za-z_$][\w$]*)\s*=[^=]/g)) {
  const name = m[1];
  if (!declared.has(name) && !whitelist.has(name)) flagged.add(name);
}

if (flagged.size > 0) {
  console.error("pagecheck: undeclared identifier(s) referenced in embedded JS:");
  for (const name of [...flagged].sort()) console.error("  - " + name);
  console.error(
    "pagecheck: fix by declaring with let/const (see v0.4.5 monthlyTargetTime bug); " +
    "if this is a false positive, extend scripts/pagecheck.js"
  );
  process.exit(1);
}
console.log(
  "pagecheck: undeclared-identifier scan OK (" + declared.size + " declarations checked)"
);
