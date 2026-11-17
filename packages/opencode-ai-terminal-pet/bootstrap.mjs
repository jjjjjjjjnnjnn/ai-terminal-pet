// Drop the overlay binary into place, once, so that `opencode plugin add
// opencode-ai-terminal-pet` really is the whole installation on Windows.
//
// The failure this exists for is silent by construction: the reporter writes
// snapshots, tries to launch `<data dir>/AI-Terminal-Pet.exe`, finds nothing,
// and the host keeps working with no panel. A user sees "installed" and no
// window, which is indistinguishable from broken.
//
// The rules, all of them borrowed from the DSH bundle's bootstrap because they
// were learned the hard way there:
//
//   * never overwrite an existing file - the user may have installed a newer
//     panel themselves, and a plugin must not downgrade it;
//   * copy through a temp sibling and rename, so a half-written exe can never
//     sit where the launcher looks;
//   * verify the vendored copy against vendor/manifest.json before trusting it;
//   * return a status, never throw and never block: this runs inside the host's
//     plugin loader, and AGENTS rule 1 (never break the host) outranks it all.
//
// POSIX is deliberately not handled here yet. The panel for Linux/macOS is a
// tarball plus install.sh (fonts, AppImage-less Qt-free build, autostart entry),
// and pretending a single binary drop is the same thing would ship a half
// install. The package README says so, and --doctor reports what is missing.

import { chmodSync, copyFileSync, existsSync, mkdirSync, readFileSync, renameSync, rmSync, statSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { createRequire } from 'node:module'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const require = createRequire(import.meta.url)
const protocol = require('./protocol.cjs')

const packageDir = path.dirname(fileURLToPath(import.meta.url))
const VENDOR_EXE = path.join(packageDir, 'vendor', 'AI-Terminal-Pet.exe')
const VENDOR_MANIFEST = path.join(packageDir, 'vendor', 'manifest.json')

/** Where the reporter expects to find the panel; '' when the platform has no data directory. */
export function panelPath() {
  try {
    const root = protocol.dataDirectory()
    return root ? path.join(root, process.platform === 'win32' ? 'AI-Terminal-Pet.exe' : 'AI-Terminal-Pet') : ''
  } catch {
    return ''
  }
}

function sha256(file) {
  return createHash('sha256').update(readFileSync(file)).digest('hex')
}

/** The digest vendor/manifest.json recorded for the vendored binary, or ''. */
function expectedDigest(manifestPath = VENDOR_MANIFEST) {
  try {
    const manifest = JSON.parse(readFileSync(manifestPath, 'utf8'))
    const value = manifest && manifest.overlay && manifest.overlay.sha256
    return typeof value === 'string' ? value : ''
  } catch {
    return ''
  }
}

/**
 * installPanel makes the panel available to the reporter and reports what it did.
 *
 * Statuses: installed, present (already there, left alone), skipped-platform,
 * skipped-env (AI_TERMINAL_PET_NO_PANEL_INSTALL=1), skipped-unsupported,
 * failed. The distinctions matter to the tests and to `--doctor`-style
 * diagnostics later; the host never sees them.
 */
export function installPanel(options = {}) {
  const platform = options.platform || process.platform
  try {
    if (platform !== 'win32') return { status: 'skipped-platform' }
    if (process.env.AI_TERMINAL_PET_NO_PANEL_INSTALL === '1') return { status: 'skipped-env' }
    const target = options.target || panelPath()
    if (!target) return { status: 'skipped-unsupported' }
    if (existsSync(target)) return { status: 'present', target }
    const source = options.source || VENDOR_EXE
    if (!existsSync(source)) return { status: 'skipped-unsupported' }
    const expected = options.expectedDigest || expectedDigest(options.manifestPath)
    if (!expected) return { status: 'failed', reason: 'no manifest digest' }
    if (sha256(source) !== expected) return { status: 'failed', reason: 'vendored copy does not match its manifest' }

    mkdirSync(path.dirname(target), { recursive: true })
    const temp = `${target}.${process.pid}.${Date.now()}.tmp`
    try {
      copyFileSync(source, temp)
      try {
        chmodSync(temp, 0o755)
      } catch {
        // Windows maps this to the read-only bit; a refusal is not a failure.
      }
      // rename() replaces on Windows, so re-check immediately before it.
      if (existsSync(target)) {
        rmSync(temp, { force: true })
        return { status: 'present', target }
      }
      renameSync(temp, target)
    } catch (error) {
      try {
        rmSync(temp, { force: true })
      } catch {
        // Best effort.
      }
      return { status: 'failed', reason: String((error && error.message) || error) }
    }
    if (statSync(target).size === 0) return { status: 'failed', reason: 'zero-length' }
    return { status: 'installed', target }
  } catch (error) {
    // Nothing here may reach the host's plugin loader.
    return { status: 'failed', reason: String((error && error.message) || error) }
  }
}
