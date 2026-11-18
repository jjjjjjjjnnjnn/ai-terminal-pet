import test from "node:test";
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { existsSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import path from "node:path";

const require = createRequire(import.meta.url);
const { installPanel, panelPath, expectedDigest, DATA_DIRECTORY_NAME } = require("./bootstrap.cjs");

function sha256(file) {
  return createHash("sha256").update(readFileSync(file)).digest("hex");
}

function fixture() {
  const root = mkdtempSync(path.join(tmpdir(), "pet-claude-bootstrap-"));
  const source = path.join(root, "vendor", "AI-Terminal-Pet.exe");
  mkdirSync(path.dirname(source), { recursive: true });
  writeFileSync(source, Buffer.from("panel-bytes"));
  const manifest = path.join(root, "vendor", "manifest.json");
  writeFileSync(manifest, JSON.stringify({ overlay: { sha256: sha256(source) } }));
  const target = path.join(root, "local", DATA_DIRECTORY_NAME, "AI-Terminal-Pet.exe");
  return { root, source, manifest, target };
}

test("installs the panel once from the vendored copy", () => {
  const { source, manifest, target } = fixture();
  const result = installPanel({ platform: "win32", source, manifestPath: manifest, target });
  assert.equal(result.status, "installed");
  assert.equal(readFileSync(target).toString(), readFileSync(source).toString());
  assert.deepEqual(readdirSync(path.dirname(target)).filter((name) => name.endsWith(".tmp")), []);
});

test("never overwrites a panel that is already there", () => {
  const { source, manifest, target } = fixture();
  mkdirSync(path.dirname(target), { recursive: true });
  writeFileSync(target, "the user's own newer panel");
  const result = installPanel({ platform: "win32", source, manifestPath: manifest, target });
  assert.equal(result.status, "present");
  assert.equal(readFileSync(target).toString(), "the user's own newer panel");
});

test("refuses a vendored copy that does not match its manifest", () => {
  const { source, manifest, target } = fixture();
  writeFileSync(manifest, JSON.stringify({ overlay: { sha256: "0".repeat(64) } }));
  const result = installPanel({ platform: "win32", source, manifestPath: manifest, target });
  assert.equal(result.status, "failed");
  assert.equal(result.reason, "vendored copy does not match its manifest");
  assert.equal(existsSync(target), false);
});

test("refuses to trust a missing or unreadable manifest", () => {
  const { source, target } = fixture();
  const result = installPanel({
    platform: "win32",
    source,
    manifestPath: path.join(path.dirname(source), "absent.json"),
    target
  });
  assert.equal(result.status, "failed");
  assert.equal(result.reason, "no manifest digest");
  assert.equal(existsSync(target), false);
});

test("skips platforms with no panel install path", () => {
  const { source, manifest, target } = fixture();
  for (const platform of ["linux", "darwin"]) {
    const result = installPanel({ platform, source, manifestPath: manifest, target });
    assert.equal(result.status, "skipped-platform");
  }
});

test("answers with the data-directory contract the reporter uses", () => {
  const env = { LOCALAPPDATA: path.join("C:", "Users", "someone", "AppData", "Local") };
  assert.equal(
    panelPath({ platform: "win32", env }),
    path.join(env.LOCALAPPDATA, DATA_DIRECTORY_NAME, "AI-Terminal-Pet.exe")
  );
  assert.equal(panelPath({ platform: "win32", env: {} }), "");
  assert.equal(panelPath({ platform: "linux", env }), "");
});

test("the environment escape hatch is honoured", () => {
  const { source, manifest, target } = fixture();
  process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL = "1";
  try {
    const result = installPanel({ platform: "win32", source, manifestPath: manifest, target });
    assert.equal(result.status, "skipped-env");
    assert.equal(existsSync(target), false);
  } finally {
    delete process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL;
  }
});

test("expectedDigest reads the vendored manifest and tolerates damage", () => {
  const { source, manifest } = fixture();
  assert.equal(expectedDigest(manifest), sha256(source));
  assert.equal(expectedDigest(path.join(path.dirname(manifest), "absent.json")), "");
  writeFileSync(manifest, "{ not json");
  assert.equal(expectedDigest(manifest), "");
});
