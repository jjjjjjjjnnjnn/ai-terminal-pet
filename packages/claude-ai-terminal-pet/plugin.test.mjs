import test from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const packageDir = path.dirname(fileURLToPath(import.meta.url));
const repoDir = path.resolve(packageDir, "..", "..");

const readJson = (relative) => JSON.parse(readFileSync(path.join(packageDir, relative), "utf8"));
const sha256 = (file) => createHash("sha256").update(readFileSync(file)).digest("hex");

// The eight events internal/connect/snippet.go documents and has measured.
// Keep this list in the same order: the compatibility check in the app reads
// the host config and expects to see all of them.
const EVENTS = [
  "SessionStart",
  "UserPromptSubmit",
  "PreToolUse",
  "PostToolUse",
  "PermissionRequest",
  "Notification",
  "Stop",
  "SessionEnd"
];

function overlayVersion() {
  const source = readFileSync(path.join(repoDir, "internal", "app", "version.go"), "utf8");
  const match = /^const Version = "(\d+\.\d+\.\d+)"/m.exec(source);
  assert.ok(match, "internal/app/version.go must declare const Version");
  return match[1];
}

test("the plugin version is the overlay version", () => {
  const manifest = readJson(".claude-plugin/plugin.json");
  assert.equal(manifest.version, overlayVersion());
});

test("the marketplace points at this package and states the same version", () => {
  const marketplace = JSON.parse(readFileSync(path.join(repoDir, ".claude-plugin", "marketplace.json"), "utf8"));
  const entry = marketplace.plugins.find((plugin) => plugin.name === "ai-terminal-pet");
  assert.ok(entry, "marketplace must list ai-terminal-pet");
  assert.equal(entry.source, "./packages/claude-ai-terminal-pet");
  assert.equal(entry.version, overlayVersion());
  assert.ok(existsSync(path.join(repoDir, entry.source, ".claude-plugin", "plugin.json")));
});

test("every documented event is wired to the native reporter", () => {
  const hooks = readJson(path.join("hooks", "hooks.json")).hooks;
  assert.deepEqual(Object.keys(hooks).sort(), [...EVENTS].sort());
  for (const event of EVENTS) {
    const commands = hooks[event].flatMap((group) => group.hooks.map((hook) => hook.command));
    const native = commands.filter((command) => command.includes("${CLAUDE_PLUGIN_ROOT}/vendor/pet-hook.exe"));
    assert.equal(native.length, 1, `${event} must run pet-hook exactly once`);
    assert.match(native[0], /--mode=claude$/);
  }
});

test("SessionStart installs the bundled panel before reporting", () => {
  const hooks = readJson(path.join("hooks", "hooks.json")).hooks;
  const commands = hooks.SessionStart[0].hooks.map((hook) => hook.command);
  assert.equal(commands[0], 'node "${CLAUDE_PLUGIN_ROOT}/bootstrap.cjs"');
  assert.match(commands[1], /vendor\/pet-hook\.exe/);
});

test("the vendored panel matches its manifest and bin/", () => {
  const manifest = readJson(path.join("vendor", "manifest.json"));
  const vendored = path.join(packageDir, "vendor", "AI-Terminal-Pet.exe");
  assert.equal(manifest.schemaVersion, 1);
  assert.equal(manifest.overlay.bytes, readFileSync(vendored).length);
  assert.equal(manifest.overlay.sha256, sha256(vendored));
  const built = path.join(repoDir, "bin", "AI-Terminal-Pet.exe");
  if (existsSync(built)) assert.equal(sha256(built), sha256(vendored), "vendor/ must be copied from bin/");
});

test("the vendored hook matches its manifest and bin/", () => {
  const manifest = readJson(path.join("vendor", "manifest.json"));
  const vendored = path.join(packageDir, "vendor", "pet-hook.exe");
  assert.equal(manifest.hook.bytes, readFileSync(vendored).length);
  assert.equal(manifest.hook.sha256, sha256(vendored));
  const built = path.join(repoDir, "bin", "pet-hook.exe");
  if (existsSync(built)) assert.equal(sha256(built), sha256(vendored), "vendor/ must be copied from bin/");
});

test("no machine-specific path leaked into the package", () => {
  const files = [
    ".claude-plugin/plugin.json",
    "hooks/hooks.json",
    "bootstrap.cjs",
    "README.md"
  ];
  for (const file of files) {
    const text = readFileSync(path.join(packageDir, file), "utf8");
    assert.doesNotMatch(text, /[A-Za-z]:\\Users\\/);
    assert.doesNotMatch(text, /\/Users\/[^/]+\//);
  }
});
