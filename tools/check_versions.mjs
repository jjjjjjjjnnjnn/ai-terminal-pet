#!/usr/bin/env node
/**
 * Fail when a document states a version that the source of truth disagrees with.
 *
 * Why this exists
 * ---------------
 * This repository ships two version lines - the overlay (`pet-v*`, the constant
 * in internal/app/app_windows.go) and the DSH plugin package (`dsh-plugin-v*`,
 * packages/dsh-ai-terminal-pet/package.json) - and about a dozen sentences in
 * the documentation name one of them. Nothing used to read those sentences, so
 * they rotted exactly as fast as nobody looked:
 *
 *   - the README badge sat at 1.0.2 while the overlay went 1.0.3 -> 1.0.7;
 *   - both README "the package is now" paragraphs said 1.2.8 while npm served
 *     1.2.13 (three released versions of drift);
 *   - docs/LAUNCH.md still told the reader to link the `pet-v1.0.0` release;
 *   - the ROADMAP baseline table carried a 302-test count and a 1.0.6 binary.
 *
 * Every one of those was found by accident, while bumping something else. This
 * tool is that "while" made automatic: it knows the *shapes* of a currency claim
 * (a badge, a release link, an install command, a "current version is" sentence,
 * a baseline row) and requires each to state the version the code states.
 *
 * What it catches, honestly
 * -------------------------
 *   1. Every known claim shape, and it requires the expected NUMBER of them, so
 *      deleting a sentence and leaving another stale one cannot pass silently.
 *   2. A generic net for a different mistake: no document may name a version
 *      *newer* than the current one in its own line (1.0.N for the overlay,
 *      1.2.N for the plugin). That is a half-finished bump or a typo, e.g.
 *      1.0.9 while the constant says 1.0.7.
 *   3. The CHANGELOG must have a heading for the current overlay version, must
 *      not have one for a newer overlay version, and `[未发布]` - if present -
 *      must be the newest heading (a released tree that still says "unreleased"
 *      one heading lower is the classic half-done release commit).
 *
 * What it does NOT catch: a *new* sentence that names a version in a way nobody
 * wrote down here. When you add a version sentence to a document, add its shape
 * to CLAIMS below; when you reword one, the failure will say the shape went
 * missing, and updating the regex is part of the edit rather than an annoyance.
 *
 * Usage: node tools/check_versions.mjs
 * Exit codes: 0 every claim agrees, 1 at least one does not, 2 a file is missing.
 */
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

const OVERLAY_SOURCE = 'internal/app/app_windows.go'
const ROOT_MANIFEST = 'package.json'
const PACKAGE_MANIFEST = 'packages/dsh-ai-terminal-pet/package.json'
const PACKAGE_README = 'packages/dsh-ai-terminal-pet/README.md'
const MARKET_INDEX = 'docs/market-index.json'

// Documents that may name a version at all. Used by the generic sweep, so a new
// document that starts naming versions has to be listed here to be watched.
const DOCUMENTS = [
  'README.md',
  PACKAGE_README,
  'CHANGELOG.md',
  MARKET_INDEX,
  'docs/ROADMAP.md',
  'docs/HANDOVER.md',
  'docs/LAUNCH.md'
]

/**
 * Every known "this is the current version" shape.
 *
 * `expect` is one entry per capture group: 'overlay' or 'plugin'. `count` is how
 * many times the shape must appear - not "at least", because a document that
 * keeps an English and a Chinese copy of a claim has exactly two of them, and a
 * count that silently drops is how the README badge survived four releases.
 */
export const CLAIMS = [
  {
    id: 'readme badge',
    file: 'README.md',
    pattern: /img\.shields\.io\/badge\/version-(\d+\.\d+\.\d+)-blue/g,
    expect: ['overlay'],
    count: 1
  },
  {
    id: 'readme release link (en + zh)',
    file: 'README.md',
    pattern: /releases\/tag\/pet-v(\d+\.\d+\.\d+)/g,
    expect: ['overlay'],
    count: 2
  },
  {
    id: 'readme portable zip name (en + zh)',
    file: 'README.md',
    pattern: /ai-terminal-pet-(\d+\.\d+\.\d+)-win-x64\.zip/g,
    expect: ['overlay'],
    count: 2
  },
  {
    id: 'readme tarball install URL (en + zh)',
    file: 'README.md',
    pattern: /releases\/download\/dsh-plugin-v(\d+\.\d+\.\d+)\/dsh-ai-terminal-pet-(\d+\.\d+\.\d+)\.tgz/g,
    expect: ['plugin', 'plugin'],
    count: 2
  },
  {
    id: 'readme git install tag (en + zh)',
    file: 'README.md',
    pattern: /git#dsh-plugin-v(\d+\.\d+\.\d+)&path:/g,
    expect: ['plugin'],
    count: 2
  },
  {
    id: 'readme "the package is now" (en)',
    file: 'README.md',
    pattern: /The package is now \*\*(\d+\.\d+\.\d+)\*\*/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'readme "当前版本是" (zh)',
    file: 'README.md',
    pattern: /当前版本是 \*\*(\d+\.\d+\.\d+)\*\*/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'readme npm version link (en)',
    file: 'README.md',
    pattern: /npmjs\.com\/package\/dsh-ai-terminal-pet\) \((\d+\.\d+\.\d+)\)/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'readme npm version link (zh)',
    file: 'README.md',
    pattern: /package\/dsh-ai-terminal-pet\)，(\d+\.\d+\.\d+)）/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'readme --version example',
    file: 'README.md',
    pattern: /打印版本信息（v(\d+\.\d+\.\d+)）/g,
    expect: ['overlay'],
    count: 1
  },
  {
    id: 'packaged readme tarball install URL',
    file: PACKAGE_README,
    pattern: /releases\/download\/dsh-plugin-v(\d+\.\d+\.\d+)\/dsh-ai-terminal-pet-(\d+\.\d+\.\d+)\.tgz/g,
    expect: ['plugin', 'plugin'],
    count: 1
  },
  {
    id: 'packaged readme git install tag',
    file: PACKAGE_README,
    pattern: /git#dsh-plugin-v(\d+\.\d+\.\d+)&path:/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'roadmap baseline row',
    file: 'docs/ROADMAP.md',
    pattern: /\| 版本 \| `(\d+\.\d+\.\d+)`（`internal\/app\/app_windows\.go:29` 的 `const Version`）；插件包 `(\d+\.\d+\.\d+)` \|/g,
    expect: ['overlay', 'plugin'],
    count: 1
  },
  {
    id: 'handover status heading',
    file: 'docs/HANDOVER.md',
    pattern: /浮窗 \*\*(\d+\.\d+\.\d+)\*\*、插件包 \*\*(\d+\.\d+\.\d+)\*\*/g,
    expect: ['overlay', 'plugin'],
    count: 1
  },
  {
    id: 'handover version bullet',
    file: 'docs/HANDOVER.md',
    pattern: /浮窗 \*\*`(\d+\.\d+\.\d+)`\*\*（`internal\/app\/app_windows\.go:29` 的 `const Version`），插件包 \*\*`(\d+\.\d+\.\d+)`\*\*/g,
    expect: ['overlay', 'plugin'],
    count: 1
  },
  {
    id: 'launch posting pointer',
    file: 'docs/LAUNCH.md',
    pattern: /releases\/tag\/pet-v(\d+\.\d+\.\d+)` 这个链接/g,
    expect: ['overlay'],
    count: 1
  },
  {
    id: 'launch prerequisites line',
    file: 'docs/LAUNCH.md',
    pattern: /npm `latest` = (\d+\.\d+\.\d+)/g,
    expect: ['plugin'],
    count: 1
  },
  {
    id: 'launch baseline table',
    file: 'docs/LAUNCH.md',
    pattern: /\*\*(\d+\.\d+\.\d+)\*\*（`dist-tags\.latest`/g,
    expect: ['plugin'],
    count: 1
  }
]

// ---------------------------------------------------------------- pure parts

/**
 * extractOverlayVersion reads the constant the overlay reports from --version.
 * Exactly one declaration is allowed: a second one is how a build starts
 * reporting a version nobody meant to ship.
 */
export function extractOverlayVersion(source) {
  const matches = [...source.matchAll(/^const Version = "(\d+\.\d+\.\d+)"$/gm)]
  if (matches.length !== 1) {
    return { error: `expected exactly 1 'const Version = "x.y.z"' line, found ${matches.length}` }
  }
  return { version: matches[0][1] }
}

export function compareVersions(a, b) {
  const left = a.split('.').map(Number)
  const right = b.split('.').map(Number)
  for (let index = 0; index < 3; index += 1) {
    if (left[index] !== right[index]) return left[index] < right[index] ? -1 : 1
  }
  return 0
}

function lineOf(text, index) {
  let line = 1
  for (let cursor = 0; cursor < index; cursor += 1) {
    if (text.charCodeAt(cursor) === 10) line += 1
  }
  return line
}

/**
 * checkClaims compares every known claim shape against the versions the code
 * states. It returns human-readable failures, never throws, so the CLI and the
 * unit tests read the same verdict.
 */
export function checkClaims(documents, versions, claims = CLAIMS) {
  const failures = []
  for (const claim of claims) {
    const text = documents[claim.file]
    if (text === undefined) {
      failures.push(`fail: ${claim.id}: ${claim.file} was not read`)
      continue
    }
    const matches = [...text.matchAll(claim.pattern)]
    if (matches.length !== claim.count) {
      failures.push(
        `fail: ${claim.id}: found ${matches.length} match(es) in ${claim.file}, expected ${claim.count}` +
          (matches.length === 0 ? ' (reworded? then update CLAIMS in tools/check_versions.mjs)' : '')
      )
      continue
    }
    for (const match of matches) {
      claim.expect.forEach((kind, group) => {
        const found = match[group + 1]
        const want = versions[kind]
        if (found !== want) {
          failures.push(
            `fail: ${claim.id}: ${claim.file}:${lineOf(text, match.index)} says ${found}, ` +
              `the ${kind} version is ${want}`
          )
        }
      })
    }
  }
  return failures
}

/**
 * stripCode blanks fenced blocks and inline code spans, keeping the line count so
 * reported line numbers still point at the real line.
 *
 * The sweep below is about prose that names the current version. A fence is a
 * command and a backtick span is a literal, and either may name an unreleased
 * version on purpose - docs/HANDOVER.md demonstrates `tools/upstream-entry.mjs
 * --check --version 1.2.99` to show that an untagged version SKIPs instead of
 * failing the build. The version strings that matter inside code (install URLs,
 * asset names, baseline rows) are covered by CLAIMS, which do not strip anything.
 */
export function stripCode(text) {
  let fenced = false
  return text
    .split('\n')
    .map((line) => {
      if (line.trimStart().startsWith('```')) {
        fenced = !fenced
        return ''
      }
      if (fenced) return ''
      return line.replace(/`[^`\n]*`/g, '``')
    })
    .join('\n')
}

/**
 * sweepFutureVersions is the generic net, deliberately narrow: it only flags a
 * version literal that is NEWER than the current one in the same release line.
 * Older versions are history and must stay; a newer one is always a mistake.
 */
export function sweepFutureVersions(documents, versions) {
  const failures = []
  for (const file of Object.keys(documents)) {
    const prose = stripCode(documents[file])
    for (const match of prose.matchAll(/\b(1\.(?:0|2)\.(\d+))\b/g)) {
      const kind = match[1].startsWith('1.0.') ? 'overlay' : 'plugin'
      if (compareVersions(match[1], versions[kind]) > 0) {
        failures.push(
          `fail: ${file}:${lineOf(prose, match.index)} names ${match[1]}, which is newer than the ` +
            `${kind} version ${versions[kind]} (half-finished bump or a typo)`
        )
      }
    }
  }
  return failures
}

/**
 * checkChangelog pins the shape of the newest headings. This file documents both
 * release lines, so a heading may be an overlay version (1.0.N) or a plugin
 * version (1.2.N) - the rule is only that neither may name a version newer than
 * its own line's current one, and that the overlay's current version is there.
 *
 * `[未发布]` is legitimate in a working tree, so it is not an error - but it must
 * not sit below an unreleased *overlay* version, which is the shape of a
 * half-done release commit. A released plugin version above it is fine: the two
 * lines ship independently (1.2.15 is a plugin-only release while the overlay
 * batch below it is still unshipped).
 */
export function checkChangelog(text, versions) {
  const failures = []
  const headings = [...text.matchAll(/^## \[([^\]]+)\]/gm)].map((match) => match[1])
  if (headings.length === 0) {
    return ['fail: CHANGELOG.md has no "## [version]" heading at all']
  }
  if (!headings.includes(versions.overlay)) {
    failures.push(
      `fail: CHANGELOG.md has no "## [${versions.overlay}]" heading for the current overlay version`
    )
  }
  for (const heading of headings) {
    if (!/^1\.[02]\.\d+$/.test(heading)) continue
    const kind = heading.startsWith('1.0.') ? 'overlay' : 'plugin'
    if (compareVersions(heading, versions[kind]) > 0) {
      failures.push(
        `fail: CHANGELOG.md has a "## [${heading}]" heading, newer than the ${kind} version ${versions[kind]}`
      )
    }
  }
  const unreleased = headings.indexOf('未发布')
  if (unreleased > 0) {
    const overlayAbove = headings.slice(0, unreleased).filter((heading) => /^1\.0\.\d+$/.test(heading))
    if (overlayAbove.length > 0) {
      failures.push(
        `fail: CHANGELOG.md lists [未发布] below ${overlayAbove.join(', ')}: ` +
          'an unreleased overlay batch cannot sit under a released overlay version'
      )
    }
  }
  return failures
}

// --------------------------------------------------------------------- CLI

export function check(documents, manifests, overlaySource, claims = CLAIMS) {
  const parsed = extractOverlayVersion(overlaySource)
  if (parsed.error) return [`fail: ${OVERLAY_SOURCE}: ${parsed.error}`]

  const failures = []
  if (manifests.root !== manifests.package_) {
    failures.push(
      `fail: two plugin manifests disagree: ${ROOT_MANIFEST} says ${manifests.root}, ` +
        `${PACKAGE_MANIFEST} says ${manifests.package_}`
    )
  }
  const versions = { overlay: parsed.version, plugin: manifests.root }
  if (manifests.index !== versions.plugin) {
    failures.push(`fail: ${MARKET_INDEX} lists ${manifests.index}, the packaged version is ${versions.plugin}`)
  }
  failures.push(...checkClaims(documents, versions, claims))
  failures.push(...sweepFutureVersions(documents, versions))
  failures.push(...checkChangelog(documents['CHANGELOG.md'] ?? '', versions))
  return failures
}

function readJson(relative) {
  return JSON.parse(readFileSync(path.join(repoDir, relative), 'utf8'))
}

function main() {
  const documents = {}
  const missing = []
  for (const file of DOCUMENTS) {
    try {
      documents[file] = readFileSync(path.join(repoDir, file), 'utf8')
    } catch {
      missing.push(file)
    }
  }
  if (missing.length > 0) {
    console.log(`versions: cannot read ${missing.join(', ')} - refusing to guess`)
    process.exit(2)
  }

  const overlaySource = readFileSync(path.join(repoDir, OVERLAY_SOURCE), 'utf8')
  const manifests = {
    root: readJson(ROOT_MANIFEST).version,
    package_: readJson(PACKAGE_MANIFEST).version,
    index: readJson(MARKET_INDEX).plugins[0].versions[0].version
  }

  const failures = check(documents, manifests, overlaySource)
  const overlay = extractOverlayVersion(overlaySource).version
  for (const failure of failures) console.log(failure)
  if (failures.length > 0) {
    console.log(`\nversions: ${failures.length} claim(s) disagree with the code (overlay ${overlay}, plugin ${manifests.root})`)
    process.exit(1)
  }
  console.log(
    `versions: ${CLAIMS.length} claim shapes agree - overlay ${overlay}, plugin ${manifests.root}, ` +
      `index ${manifests.index}, changelog and the future-version sweep clean`
  )
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main()
}
