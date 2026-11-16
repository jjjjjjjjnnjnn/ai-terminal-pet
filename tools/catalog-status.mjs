#!/usr/bin/env node
/**
 * Is our row on awesome-dsh-plugin.com actually current?
 *
 * docs/UPSTREAM.md §4 states the rule: "merged" is not "visible", and the
 * authority is the site's own index file rather than a PR page. This tool reads
 * that file and compares it with npm and GitHub, so the answer is one command
 * instead of three browser tabs.
 *
 *   node tools/catalog-status.mjs                 # live: site + npm + GitHub
 *   node tools/catalog-status.mjs --file f.json   # offline: site file only
 *   node tools/catalog-status.mjs --json          # machine-readable
 *
 * Why it exists (measured 2026-11-16, handover §64):
 *
 *   - the site carried `version: 1.2.16` for our entry while npm had published
 *     1.2.21 and 1.2.22, because the catalogue has been frozen since
 *     2026-11-14T05:27Z: the nightly `Build` step failed on 2026-11-15 and again
 *     on 2026-11-16 with `MarketIndexError: $.plugins[N].name duplicates
 *     dsh-dot-mcp`. That is upstream's bug (PR #15579), not a stale entry of
 *     ours - `tarball:` in the yml was already 1.2.21.
 *   - the site's 30-day `downloads` is short by exactly the last day of the
 *     requested window, upstream-wide, because the probe runs at ~03:1x UTC
 *     before npm has aggregated the day that ended three hours earlier
 *     (issue #15864). A re-probe does not fix it; only a window change does.
 *
 * So the two fields fail for two different reasons, and this tool distinguishes
 * them instead of printing "out of date" once. A number that is wrong for
 * everyone is not evidence that we did something wrong.
 *
 * Exit codes: 0 current (or a documented upstream cause), 1 our row is stale in
 * a way we can fix, 2 the site file could not be read.
 */
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const HERE = path.dirname(fileURLToPath(import.meta.url))
const REPO = path.resolve(HERE, '..')

export const NPM_NAME = 'dsh-ai-terminal-pet'
export const REPO_URL = 'https://github.com/jjjjjjjjnnjnn/ai-terminal-pet'
export const SITE_INDEX = 'https://awesome-dsh-plugin.com/plugins.json'
export const NPM_REGISTRY = 'https://api.npmjs.org/downloads'

/** Compare dotted numeric versions. Same shape as tools/check_versions.mjs. */
export function compareVersions(left, right) {
  const a = String(left).split('.').map(Number)
  const b = String(right).split('.').map(Number)
  for (let index = 0; index < 3; index += 1) {
    if (a[index] !== b[index]) return a[index] < b[index] ? -1 : 1
  }
  return 0
}

/** The site keys one row per npm name; find ours without depending on order. */
export function findEntry(index, npmName = NPM_NAME) {
  const plugins = index?.plugins
  if (!Array.isArray(plugins)) return null
  return plugins.find((entry) => entry?.npm === npmName) ?? null
}

/**
 * The download shortfall, as two numbers rather than a boolean.
 *
 * `site` is what the catalogue published for its own recorded window; `now` is
 * the same window asked again; `withoutLastDay` is that window minus its final
 * day. When site == withoutLastDay and now > site, the tail day was not yet
 * aggregated when the probe ran - the upstream-wide cause in issue #15864.
 */
export function downloadShortfall(site, days) {
  if (!Number.isFinite(site) || !Array.isArray(days) || days.length === 0) return null
  const now = days.reduce((total, day) => total + (Number(day?.downloads) || 0), 0)
  const withoutLastDay = days.slice(0, -1).reduce((total, day) => total + (Number(day?.downloads) || 0), 0)
  const lastDay = Number(days[days.length - 1]?.downloads) || 0
  return {
    site,
    now,
    withoutLastDay,
    lastDay,
    // The signature of the aggregation-lag bug: we are short by exactly the
    // final day, so the published total equals the window minus that day.
    lagging: site === withoutLastDay && now > site,
    shortfall: Math.max(0, now - site),
  }
}

/**
 * The verdict for one row. Pure: every fact it needs is passed in, so a test can
 * describe the frozen-catalogue case and the lagging-downloads case without a
 * network.
 */
export function verdict({ entry, npmVersion, publishedVersion, downloads, localVersion }) {
  if (!entry) return { level: 'fail', reason: 'our row is not in the published index' }
  // The row's `version` is a cached npm probe from the last successful nightly.
  // When it is behind npm, nothing we publish can fix it: the catalogue has to
  // be rebuilt, and that is upstream's job.
  if (npmVersion && entry.version && compareVersions(entry.version, npmVersion) < 0) {
    return {
      level: 'upstream',
      reason: `row says ${entry.version}, npm has ${npmVersion} - the catalogue has not been rebuilt since its last successful nightly`,
      fixableByUs: false,
    }
  }
  if (entry.downloads != null && downloads?.lagging) {
    return {
      level: 'upstream',
      reason: `downloads ${downloads.site} vs ${downloads.now} for the same window: short by exactly the final day (upstream issue #15864)`,
      fixableByUs: false,
    }
  }
  // The tarball half is the one WE control (tools/upstream-entry.mjs writes it;
  // the maintainer's merge is the only other hand on it). Offline, compare it
  // against this working tree's package version; online, against npm.
  const newest = npmVersion ?? localVersion ?? null
  if (publishedVersion && newest && compareVersions(publishedVersion, newest) < 0) {
    return {
      level: npmVersion ? 'fail' : 'pending',
      reason: npmVersion
        ? `our newest release is ${npmVersion} but the entry's tarball points at ${publishedVersion}`
        : `the entry's tarball points at ${publishedVersion}; this working tree is at ${localVersion}, so either the bump PR is in flight or none was opened (tools/upstream-entry.mjs --check answers which)`,
      // A bump PR is the answer, and that is ours to open; whether it has been
      // merged is not. Both levels therefore mean "look at the PR".
      fixableByUs: true,
    }
  }
  return { level: 'ok', reason: 'row is current' }
}

/**
 * One entry the site would have to publish for us to be current, in the shape
 * `verdict` wants. Pure, so the test can build the three cases (current,
 * lagging downloads, frozen catalogue) without a network.
 */
export function describeEntry(entry) {
  if (!entry) return null
  return {
    name: entry.name ?? null,
    version: entry.version ?? null,
    tarball: entry.tarball ?? null,
    publishedVersion: tarballVersion(entry.tarball),
    downloads: entry.downloads ?? null,
    window: entry.downloadsStart && entry.downloadsEnd
      ? `${entry.downloadsStart}:${entry.downloadsEnd}`
      : null,
  }
}

function readLocalVersion() {
  const source = fs.readFileSync(path.join(REPO, 'internal', 'app', 'version.go'), 'utf8')
  const match = source.match(/^const Version = "(\d+\.\d+\.\d+)"$/m)
  return match ? match[1] : null
}

function readPluginVersion() {
  const manifest = JSON.parse(fs.readFileSync(path.join(REPO, 'packages', 'dsh-ai-terminal-pet', 'package.json'), 'utf8'))
  return manifest.version
}

async function fetchJson(url) {
  const response = await fetch(url, { headers: { 'user-agent': 'ai-terminal-pet catalog-status' } })
  if (!response.ok) throw new Error(`${url} -> ${response.status}`)
  return response.json()
}

/** Version referenced by the entry's own `tarball:` URL, if it carries one. */
export function tarballVersion(tarball) {
  if (typeof tarball !== 'string') return null
  const match = tarball.match(/dsh-ai-terminal-pet-(\d+\.\d+\.\d+)\.tgz/)
  return match ? match[1] : null
}

async function main(argv) {
  const fileIndex = argv.indexOf('--file')
  const asJson = argv.includes('--json')
  const offline = fileIndex !== -1

  let index
  try {
    index = offline
      ? JSON.parse(fs.readFileSync(argv[fileIndex + 1], 'utf8'))
      : await fetchJson(SITE_INDEX)
  } catch (error) {
    console.error(`catalog-status: cannot read the site index: ${error.message}`)
    return 2
  }

  const entry = findEntry(index)
  const lines = []
  const push = (level, message) => lines.push({ level, message })

  push(entry ? 'ok' : 'fail', entry ? `row present: ${entry.name}` : `row absent from ${index?.plugins?.length ?? 0} entries`)

  if (entry) {
    const described = describeEntry(entry)
    push('info', `site version ${described.version ?? 'null'} · tarball ${described.tarball ?? 'null'}`)
    push('info', `site downloads ${entry.downloads ?? 'null'} (${entry.downloadsStart ?? '?'} → ${entry.downloadsEnd ?? '?'}) · stars ${entry.stars ?? 'null'}`)
  }

  const localOverlay = readLocalVersion()
  const localPlugin = readPluginVersion()
  push('info', `local: overlay ${localOverlay}, plugin ${localPlugin}`)

  let npmVersion = null
  let downloads = null
  if (offline) {
    push('info', 'offline: npm was not consulted, so only the tarball half can be judged')
  } else {
    try {
      npmVersion = await fetchJson(`https://registry.npmjs.org/${NPM_NAME}/latest`).then((data) => data.version)
    } catch (error) {
      push('info', `npm version unavailable: ${error.message}`)
    }
    if (entry?.downloadsStart && entry?.downloadsEnd) {
      try {
        const range = await fetchJson(`${NPM_REGISTRY}/range/${entry.downloadsStart}:${entry.downloadsEnd}/${NPM_NAME}`)
        downloads = downloadShortfall(entry.downloads, range.downloads)
      } catch (error) {
        push('info', `npm download range unavailable: ${error.message}`)
      }
    }
  }

  const result = verdict({
    entry,
    npmVersion,
    publishedVersion: tarballVersion(entry?.tarball),
    downloads,
    // The site's `version` is a cached npm probe; the tarball is what we set.
    // Offline, the newest thing we know is this working tree's package version.
    localVersion: localPlugin,
  })
  push(result.level === 'ok' ? 'ok' : result.level, result.reason)
  if (downloads) {
    push('info', `downloads for the site's own window: published ${downloads.site}, now ${downloads.now}, window-minus-last-day ${downloads.withoutLastDay}, final day ${downloads.lastDay}`)
  }

  if (asJson) {
    console.log(JSON.stringify({ entry, npmVersion, localOverlay, localPlugin, downloads, verdict: result, lines }, null, 2))
  } else {
    for (const line of lines) {
      const label = line.level === 'info' ? '     ' : `[${line.level}]`.padEnd(5)
      console.log(`${label} ${line.message}`)
    }
    console.log('')
    console.log({
      ok: 'catalog-status: current',
      pending: 'catalog-status: the tarball is behind this working tree - check whether the bump PR is still open',
      upstream: 'catalog-status: stale for an upstream reason (a re-release would not change it)',
      fail: 'catalog-status: ours to fix',
    }[result.level])
  }
  if (result.level === 'ok') return 0
  if (result.level === 'pending' || result.level === 'upstream') return 0
  return 1
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))) {
  process.exitCode = await main(process.argv.slice(2))
}
