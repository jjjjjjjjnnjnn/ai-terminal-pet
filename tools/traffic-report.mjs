#!/usr/bin/env node
/**
 * One page of numbers about this repository, measured now, with the machine
 * traffic separated from the human traffic.
 *
 * Why this exists
 * ---------------
 * This project has been quoting its own npm download count in launch copy and
 * in docs/LAUNCH.md as "the only trustworthy number". It is not trustworthy,
 * and the reason is measurable rather than philosophical: last week
 * dsh-ai-terminal-pet reported 4,586 downloads spread almost evenly over 25
 * versions (137-227 each), including 166 for `0.0.0-stage` - a version nobody
 * can install. That is a crawler fleet fetching every version once, not users.
 * The comparison that proves it is the package next door: dsh-pet reports
 * 86,416 last week with 68% of it on its two newest versions, which is what
 * humans look like.
 *
 * So the report below separates three kinds of number:
 *
 *   reach      GitHub traffic. Usable, but clones are mostly our own CI
 *              (213 runs x 4 jobs in the window this file was written) and
 *              page views are the only human-ish counter here.
 *   installs   release-asset downloads, minus a baseline: every release we
 *              download ~2 assets back to verify them, so anything near 2 per
 *              release is us.
 *   npm        total, plus the "latest share": humans concentrate on the newest
 *              version; crawlers spread evenly. This is the one derived metric
 *              worth watching. Measured baseline: this package 6%, the
 *              comparable dsh-pet 58%.
 *
 * Usage: node tools/traffic-report.mjs [--json]
 * Exit codes: 0 always (this is a report, not a gate), 2 when a source cannot
 * be read - in which case it says which one and prints nothing invented.
 */
import { execFileSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const REPO = 'jjjjjjjjnnjnn/ai-terminal-pet'
const NPM_PACKAGES = ['dsh-ai-terminal-pet', 'opencode-ai-terminal-pet']
// The competitor that occupies the same catalogue shelf and the same audience.
// Only used for the "what humans look like" comparison, never as a target.
const REFERENCE_PACKAGE = 'dsh-pet'

/** gh runs one API call and returns parsed JSON. */
function gh(endpoint) {
  const output = execFileSync('gh', ['api', endpoint], { encoding: 'utf8', maxBuffer: 32 << 20 })
  return JSON.parse(output)
}

/** npmPoint returns { downloads } or null when the registry has no data yet. */
async function npmPoint(name, period) {
  const response = await fetch(`https://api.npmjs.org/downloads/point/${period}/${name}`)
  if (!response.ok) return null
  const body = await response.json()
  return typeof body.downloads === 'number' ? body.downloads : null
}

/** npmPerVersion returns { total, versions: {version: downloads} } or null. */
async function npmPerVersion(name, period = 'last-week') {
  const response = await fetch(`https://api.npmjs.org/versions/${name}/${period}`)
  if (!response.ok) return null
  const body = await response.json()
  const versions = body.downloads && typeof body.downloads === 'object' ? body.downloads : {}
  const total = Object.values(versions).reduce((sum, value) => sum + (Number(value) || 0), 0)
  return { total, versions }
}

/**
 * latestShare is the share of downloads carried by the two newest published
 * versions. The newest two, not one, because a user who installed last week has
 * not upgraded yet - and because a single version's share is noisy when a
 * release is hours old.
 *
 * Exported because it is the one number in this file that is a judgement, and a
 * judgement should be testable.
 */
export function latestShare(versions, rank = versionRank) {
  const names = Object.keys(versions)
  if (names.length === 0) return null
  const total = names.reduce((sum, name) => sum + (Number(versions[name]) || 0), 0)
  if (total === 0) return null
  const newest = names
    .slice()
    .sort((a, b) => rank(b) - rank(a))
    .slice(0, 2)
  const top = newest.reduce((sum, name) => sum + (Number(versions[name]) || 0), 0)
  return { share: top / total, top, total, newest }
}

/** versionRank orders semver strings, ignoring prerelease suffixes for order. */
export function versionRank(version) {
  const core = String(version).split('-')[0].split('.').map((part) => Number(part) || 0)
  return core[0] * 1e6 + core[1] * 1e3 + core[2]
}

/**
 * assetSummary groups release assets by the thing they install and reports the
 * per-release rate, which is the only way to see our own verification in the
 * numbers: each release downloads about two of its own assets back.
 */
export function assetSummary(assets) {
  const kinds = [
    { name: 'DSH plugin tarball', match: (file) => /^dsh-ai-terminal-pet-.*\.tgz$/.test(file) },
    { name: 'Windows setup', match: (file) => file === 'pet-setup.exe' },
    { name: 'Windows zip', match: (file) => /win-x64\.zip$/.test(file) },
    { name: 'POSIX tarball', match: (file) => /-(linux|darwin)-(amd64|arm64)\.tar\.gz$/.test(file) }
  ]
  const grouped = kinds.map((kind) => {
    const files = assets.filter((asset) => kind.match(asset.name))
    const downloads = files.reduce((sum, asset) => sum + (Number(asset.downloads) || 0), 0)
    return {
      name: kind.name,
      releases: files.length,
      downloads,
      perRelease: files.length ? downloads / files.length : 0
    }
  })
  const total = assets.reduce((sum, asset) => sum + (Number(asset.downloads) || 0), 0)
  return { grouped, total }
}

function loadLocalVersion() {
  const source = readFileSync(path.join(repoDir, 'internal/app/version.go'), 'utf8')
  const match = source.match(/^const Version = "(\d+\.\d+\.\d+)"$/m)
  return match ? match[1] : 'unknown'
}

function fmt(number) {
  return Number(number).toLocaleString('en-US')
}

function pct(value) {
  return `${Math.round(value * 100)}%`
}

async function main() {
  const asJson = process.argv.includes('--json')
  let meta, views, clones, referrers, paths, releases, catalog, packages
  try {
    meta = gh(`repos/${REPO}`)
    views = gh(`repos/${REPO}/traffic/views`)
    clones = gh(`repos/${REPO}/traffic/clones`)
    referrers = gh(`repos/${REPO}/traffic/popular/referrers`)
    paths = gh(`repos/${REPO}/traffic/popular/paths`)
    releases = gh(`repos/${REPO}/releases?per_page=100`)
    const catalogResponse = await fetch('https://awesome-dsh-plugin.com/plugins.json')
    catalog = catalogResponse.ok ? await catalogResponse.json() : null
    packages = {}
    for (const name of NPM_PACKAGES) {
      packages[name] = {
        week: await npmPerVersion(name, 'last-week'),
        month: await npmPoint(name, 'last-month')
      }
    }
  } catch (error) {
    console.error(`traffic: cannot read a source: ${error.message}`)
    process.exit(2)
  }

  const assets = []
  for (const release of releases) {
    for (const asset of release.assets || []) {
      assets.push({ name: asset.name, downloads: asset.download_count, tag: release.tag_name })
    }
  }
  const assetsSummary = assetSummary(assets)

  const reference = await npmPerVersion(REFERENCE_PACKAGE, 'last-week')

  // CI clones, counted rather than guessed: four jobs per run, one checkout each.
  let ciRuns = 0
  try {
    const runs = JSON.parse(
      execFileSync('gh', ['run', 'list', '--limit', '300', '--json', 'createdAt'], { encoding: 'utf8' })
    )
    const since = Date.now() - 14 * 24 * 3600 * 1000
    ciRuns = runs.filter((run) => Date.parse(run.createdAt) > since).length
  } catch {
    ciRuns = 0
  }

  const notify = catalog ? (catalog.plugins || []).filter((entry) => entry.category === 'notify') : []
  const ours = notify.find((entry) => String(entry.name).includes('ai-terminal-pet'))
  const notifySorted = notify.slice().sort((a, b) => (Number(b.downloads) || 0) - (Number(a.downloads) || 0))
  const rank = ours ? notifySorted.findIndex((entry) => entry.name === ours.name) + 1 : 0
  // Entries a crawler has not reached yet carry no number at all; counting them
  // as zero would drag the median down and make our shelf position look better
  // than it is.
  const everyone = notify
    .map((entry) => Number(entry.downloads) || 0)
    .filter((value) => value > 0)
    .sort((a, b) => a - b)
  const median = everyone.length ? everyone[Math.floor(everyone.length / 2)] : 0
  const leader = notifySorted[0]

  const report = {
    measuredAt: new Date().toISOString().slice(0, 10),
    localVersion: loadLocalVersion(),
    repo: {
      stars: meta.stargazers_count,
      forks: meta.forks_count,
      subscribers: meta.subscribers_count,
      issues: meta.open_issues_count,
      views: { count: views.count, uniques: views.uniques },
      clones: { count: clones.count, uniques: clones.uniques, ciRuns, ciClones: ciRuns * 4 },
      referrers,
      paths: paths.slice(0, 5)
    },
    assets: assetsSummary,
    npm: packages,
    reference: reference ? { name: REFERENCE_PACKAGE, ...latestShare(reference.versions) } : null,
    catalog: ours
      ? { name: ours.name, downloads: ours.downloads, rank, entries: notify.length, median, leader: leader?.name, leaderDownloads: leader?.downloads }
      : null
  }

  if (asJson) {
    console.log(JSON.stringify(report, null, 2))
    return
  }

  console.log(`traffic report: ${REPO}  (measured ${report.measuredAt}, local overlay ${report.localVersion})`)
  console.log('')
  console.log('REACH (last 14 days)')
  console.log(`  page views      ${fmt(views.count)} (${fmt(views.uniques)} unique)`)
  console.log(
    `  git clones      ${fmt(clones.count)} (${fmt(clones.uniques)} unique) - ${fmt(ciRuns)} CI runs x 4 jobs = ` +
      `${fmt(ciRuns * 4)} of them are ours`
  )
  if (referrers.length) {
    console.log('  referrers       ' + referrers.map((entry) => `${entry.referrer} ${entry.count}`).join(' · '))
  }
  console.log(
    `  stars ${fmt(meta.stargazers_count)} · forks ${fmt(meta.forks_count)} · watchers ${fmt(meta.subscribers_count)} · ` +
      `open issues/PRs ${fmt(meta.open_issues_count)}`
  )
  console.log('')
  console.log('INSTALLS (release assets, all time)')
  for (const kind of assetsSummary.grouped) {
    console.log(
      `  ${kind.name.padEnd(20)} ${fmt(kind.downloads).padStart(7)}  over ${kind.releases} release(s), ` +
        `${kind.perRelease.toFixed(1)}/release`
    )
  }
  console.log(
    '  baseline            this project downloads ~2 assets per release back to verify them; ' +
      'anything at that rate is us, not a user'
  )
  console.log('')
  console.log('NPM (last week)')
  for (const name of NPM_PACKAGES) {
    const week = packages[name].week
    if (!week || Object.keys(week.versions).length === 0) {
      console.log(`  ${name}: no data yet (a version published today has none)`)
      continue
    }
    const share = latestShare(week.versions)
    const values = Object.values(week.versions).map(Number)
    const min = Math.min(...values)
    const max = Math.max(...values)
    console.log(
      `  ${name}: ${fmt(week.total)} over ${Object.keys(week.versions).length} version(s), ` +
        `${fmt(min)}-${fmt(max)} each`
    )
    console.log(`    latest share ${share ? pct(share.share) : 'n/a'} (humans ≈ 50%+, a crawler fleet ≈ 3%)`)
  }
  if (report.reference) {
    console.log(
      `  reference: ${fmt(reference.total)} for ${REFERENCE_PACKAGE} with ` +
        `${pct(report.reference.share)} on its two newest versions`
    )
  }
  console.log('')
  if (report.catalog) {
    console.log('CATALOGUE SHELF (awesome-dsh-plugin, category notify)')
    console.log(
      `  rank ${rank} / ${notify.length} by downloads (ours ${fmt(report.catalog.downloads)}, ` +
        `median ${fmt(report.catalog.median)})`
    )
    console.log(`  leader: ${report.catalog.leader} ${fmt(report.catalog.leaderDownloads)}`)
  }
  console.log('')
  console.log('Read this as: views and the latest share are the human-ish numbers here.')
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    console.error(`traffic: ${error.message}`)
    process.exit(2)
  })
}
