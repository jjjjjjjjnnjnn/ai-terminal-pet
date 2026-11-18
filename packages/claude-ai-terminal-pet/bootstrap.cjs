"use strict";

// Put the overlay in the one place the reporter looks for it, once, so that
// `claude plugin install` really is the whole installation on Windows.
//
// The failure this exists for is silent by construction: pet-hook writes
// snapshots, tries to launch <data dir>\AI-Terminal-Pet.exe, finds nothing, and
// the host keeps working with no panel. A user sees "installed" and no window,
// which is indistinguishable from broken.
//
// The rules are the ones the DSH bundle and the OpenCode package learned:
//
//   * never overwrite an existing file - the user may have installed a newer
//     panel themselves, and a plugin must not downgrade it;
//   * copy through a temp sibling and rename, so a half-written exe can never
//     sit where the launcher looks;
//   * verify the vendored copy against vendor/manifest.json before trusting it;
//   * never throw and never print: this runs inside a SessionStart hook, and
//     AGENTS rule 1 (never break the host) outranks everything else.
//
// POSIX is deliberately not handled. The panel for Linux/macOS is a tarball
// plus install.sh, and dropping one binary is not the same install. This
// release is Windows-only; the README says so and --doctor reports what is
// missing elsewhere.
//
// bootstrap.test.mjs pins every rule above with a temporary directory; nothing
// in it touches the real %LOCALAPPDATA%.

const {
  chmodSync,
  copyFileSync,
  existsSync,
  mkdirSync,
  readFileSync,
  renameSync,
  rmSync,
  statSync
} = require("node:fs");
const { createHash } = require("node:crypto");
const path = require("node:path");

const packageDir = __dirname;
const VENDOR_EXE = path.join(packageDir, "vendor", "AI-Terminal-Pet.exe");
const VENDOR_MANIFEST = path.join(packageDir, "vendor", "manifest.json");

// The directory name is a cross-language contract (internal/home,
// reporters/protocol.js): moving it would orphan every existing install.
const DATA_DIRECTORY_NAME = "AI-Terminal-Pet";

/** Where pet-hook expects the panel; "" when this platform has no answer. */
function panelPath(options = {}) {
  const platform = options.platform || process.platform;
  if (platform !== "win32") return "";
  const env = options.env || process.env;
  const root = env.LOCALAPPDATA;
  return root ? path.join(root, DATA_DIRECTORY_NAME, "AI-Terminal-Pet.exe") : "";
}

function sha256(file) {
  return createHash("sha256").update(readFileSync(file)).digest("hex");
}

/** The digest vendor/manifest.json recorded for the vendored binary, or "". */
function expectedDigest(manifestPath = VENDOR_MANIFEST) {
  try {
    const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
    const value = manifest && manifest.overlay && manifest.overlay.sha256;
    return typeof value === "string" ? value : "";
  } catch {
    return "";
  }
}

/**
 * installPanel makes the panel available to the reporter and reports what it did.
 *
 * Statuses: installed, present, skipped-platform, skipped-env,
 * skipped-unsupported, failed. The distinctions matter to the tests and to a
 * future --doctor; the host never sees them.
 */
function installPanel(options = {}) {
  const platform = options.platform || process.platform;
  try {
    if (platform !== "win32") return { status: "skipped-platform" };
    if (process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL === "1") return { status: "skipped-env" };
    const target = options.target || panelPath({ platform, env: options.env });
    if (!target) return { status: "skipped-unsupported" };
    if (existsSync(target)) return { status: "present", target };
    const source = options.source || VENDOR_EXE;
    if (!existsSync(source)) return { status: "skipped-unsupported" };
    const expected = options.expectedDigest || expectedDigest(options.manifestPath);
    if (!expected) return { status: "failed", reason: "no manifest digest" };
    if (sha256(source) !== expected) {
      return { status: "failed", reason: "vendored copy does not match its manifest" };
    }

    mkdirSync(path.dirname(target), { recursive: true });
    const temp = `${target}.${process.pid}.${Date.now()}.tmp`;
    try {
      copyFileSync(source, temp);
      try {
        chmodSync(temp, 0o755);
      } catch {
        // Windows maps this to the read-only bit; a refusal is not a failure.
      }
      // rename() replaces on Windows, so re-check immediately before it.
      if (existsSync(target)) {
        rmSync(temp, { force: true });
        return { status: "present", target };
      }
      renameSync(temp, target);
    } catch (error) {
      try {
        rmSync(temp, { force: true });
      } catch {
        // Best effort.
      }
      return { status: "failed", reason: String((error && error.message) || error) };
    }
    if (statSync(target).size === 0) return { status: "failed", reason: "zero-length" };
    return { status: "installed", target };
  } catch (error) {
    // Nothing here may reach the host's hook runner.
    return { status: "failed", reason: String((error && error.message) || error) };
  }
}

module.exports = { installPanel, panelPath, expectedDigest, DATA_DIRECTORY_NAME };

if (require.main === module) {
  installPanel();
  // Always exit 0. A bootstrap that cannot install says nothing and lets the
  // reporter keep working; a SessionStart hook must not fail the session.
  process.exitCode = 0;
}
