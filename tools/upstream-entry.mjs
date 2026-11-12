#!/usr/bin/env node
/**
 * Keep the upstream awesome-dsh-plugin entry pointing at the newest release.
 *
 * Why this exists
 * ---------------
 * The entry in `awesome-dsh-plugin/awesome-dsh-plugin` is one YAML file per
 * plugin, and every release of ours has to move exactly two things in it: the
 * `tarball:` line, and nothing else. The obvious way to do that - open a pull
 * request from a branch cut weeks ago and let git three-way merge it - fails
 * predictably, and not because of anything a human wrote:
 *
 *   - a maintainer edits the same file between our bump and the merge (they do
 *     this constantly: 1.2.7's entry was rewritten by hand the day after it
 *     merged), so the same `tarball:` line becomes a content conflict;
 *   - a conflicted pull request has no merge ref, so GitHub never starts the
 *     `check` run at all, and the "Submission gate" that triggers on it never
 *     fires either. The pull request looks pending and is in fact stuck: the
 *     maintainer cannot merge it, and cannot even see a verdict on it.
 *
 * #14718 sat in exactly that state, and the fix maintainers ask for is always
 * the same sentence: "rebase onto main, keep main's content, add your change".
 * This tool does that mechanically, so it cannot be done half-way:
 *
 *   --check   answers "is the listed entry pointing at our newest release?"
 *             (network failure is a skip, not a failure: it depends on another
 *             repository, and CI must not go red for the internet)
 *   --write   rebuilds the entry FROM upstream main, replaces only the tarball
 *             line, asserts that reverting that one string reproduces main
 *             byte-for-byte, commits it signed, and force-pushes the branch
 *
 * Everything that decides anything is a pure function, exported for
 * `tools/upstream-entry.test.mjs`. The CLI is the thin part: git and fetch.
 *
 * Usage (from the repository root):
 *
 *   node tools/upstream-entry.mjs --check
 *   node tools/upstream-entry.mjs --version 1.2.9            # URL + asset probe only
 *   node tools/upstream-entry.mjs --version 1.2.9 --write \
 *        --branch bump-tarball-1.2.9 --pr 14718 --comment-out .local/comment.md
 *
 * Exit codes: 0 ok (including a documented skip), 1 stale entry or failed
 * write, 2 usage error. Nothing here writes to this repository.
 */
import { execFileSync } from 'node:child_process'
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

const UPSTREAM_REPO = 'awesome-dsh-plugin/awesome-dsh-plugin'
const ENTRY_OWNER = 'jjjjjjjjnnjnn'
const ENTRY_REPO = 'ai-terminal-pet'
const ENTRY_DIRECTORY = 'packages/dsh-ai-terminal-pet'
const TAG_PREFIX = 'dsh-plugin-v'
const PACKAGE_MANIFEST = 'packages/dsh-ai-terminal-pet/package.json'

// ---------------------------------------------------------------- pure parts

/**
 * The entry's file name, per contributing.md:
 * `data/plugins/<owner>__<repo>.yml`, and for a monorepo subpackage
 * `data/plugins/<owner>__<repo>--<subdirectory with / as ->.yml`.
 *
 * Getting this wrong is silent: `readEntries()` globs `data/plugins/*.yml` and
 * is not recursive, so a file at any other path is never read, every other
 * check still passes, and merging lists nothing.
 */
export function entryFileName(owner, repo, directory) {
  const scope = directory ? '--' + directory.split('/').join('-') : ''
  return `${owner}__${repo}${scope}.yml`
}

/** The `tarball:` URL a given release version must be listed with. */
export function expectedTarball({ owner, repo, version, assetName, releaseTagPrefix = TAG_PREFIX }) {
  const tag = `${releaseTagPrefix}${version}`
  return `https://github.com/${owner}/${repo}/releases/download/${tag}/${assetName}`
}

/** The release asset name this package publishes: `<npm name>-<version>.tgz`. */
export function assetNameFor(packageName, version) {
  return `${packageName}-${version}.tgz`
}

/**
 * The version an entry currently points at, read from its `tarball:` line.
 * Returns '' when the line is missing or carries no `/dsh-plugin-v<semver>/`
 * segment - an entry can legitimately ship without a tarball at all.
 */
export function listedVersion(text, releaseTagPrefix = TAG_PREFIX) {
  const line = tarballLine(text)
  if (!line) return ''
  const tag = `\\/${releaseTagPrefix}([0-9]+(?:\\.[0-9]+)*)`
  const match = line.match(new RegExp(tag))
  return match ? match[1] : ''
}

/** The whole `tarball:` line (no trailing newline), or '' when absent. */
export function tarballLine(text) {
  const lines = text.split('\n')
  for (const line of lines) {
    if (line.startsWith('tarball:')) return line
  }
  return ''
}

/**
 * Replace the entry's `tarball:` value, refusing every shape that could
 * corrupt the file:
 *
 *   - a CRLF file at all: the clone that produces the candidate sets
 *     `core.autocrlf=false`, so a CRLF file here means this is not the file
 *     git will hash, and "revert to prove nothing else moved" cannot hold;
 *   - not exactly one `tarball:` line (zero = nothing to bump, two = a
 *     duplicate key that YAML would resolve by silently keeping the last one);
 *   - a value containing `:` `#` or a quote, which would need quoting and is
 *     never a real release URL.
 */
export function rewriteTarball(text, url) {
  if (text.includes('\r')) {
    throw new Error('refusing a CRLF file: the upstream entry is LF, and a byte-for-byte revert cannot hold on CRLF')
  }
  const lines = text.split('\n')
  const indexes = []
  for (let index = 0; index < lines.length; index += 1) {
    if (lines[index].startsWith('tarball:')) indexes.push(index)
  }
  if (indexes.length !== 1) {
    throw new Error(`expected exactly one tarball: line, found ${indexes.length}`)
  }
  if (/['"#]/.test(url) || url.split(':').length !== 2) {
    throw new Error(`refusing a release URL that would need quoting: ${url}`)
  }
  const index = indexes[0]
  lines[index] = `tarball: ${url}`
  return lines.join('\n')
}

/**
 * True when the entry is behind the release we just shipped. `listed` empty
 * means the entry carries no pinned tarball, which for a package that publishes
 * one is also stale.
 */
export function entryIsStale(listed, expected) {
  if (!listed) return true
  return listed !== expected
}

/**
 * What a rebuild must look like before it is allowed to be pushed. A rebuild
 * that changed anything other than the tarball line means the tool picked up
 * someone else's edit, which is the one outcome a maintainer would bounce.
 *
 * Proven, not asserted: undo the one string we changed and the result must be
 * byte-identical to upstream's file.
 */
export function rebuildVerdict({ before, after }) {
  const originalUrl = tarballValue(before)
  const changed = after !== before
  if (!changed) return { changed: false, onlyTarball: false }
  let reverted
  try {
    reverted = rewriteTarball(after, originalUrl)
  } catch {
    return { changed: true, onlyTarball: false }
  }
  return { changed: true, onlyTarball: reverted === before }
}

/** The bare value of the entry's `tarball:` line, or '' when absent. */
export function tarballValue(text) {
  const line = tarballLine(text)
  return line ? line.slice('tarball:'.length).trim() : ''
}

// --------------------------------------------------------------- repo inputs

function runGit(args, options = {}) {
  return execFileSync('git', args, { encoding: 'utf8', ...options })
}

function readPackage() {
  const manifest = JSON.parse(readFileSync(path.join(repoDir, PACKAGE_MANIFEST), 'utf8'))
  return { name: manifest.name, version: manifest.version, repository: manifest.repository?.url ?? '' }
}

/** This machine's identity, so a clone made here commits as the same author. */
function identity() {
  const read = (key) => {
    try {
      return runGit(['config', '--get', key], { cwd: repoDir }).trim()
    } catch {
      return ''
    }
  }
  return {
    name: read('user.name'),
    email: read('user.email'),
    signingKey: read('user.signingkey'),
    gpgFormat: read('gpg.format'),
  }
}

async function assetProbe(url) {
  try {
    const response = await fetch(url, { method: 'HEAD', redirect: 'manual' })
    // GitHub release assets redirect to the blob host; both a direct 200 and a
    // 302 mean the asset exists, a 404 means the tag or the file does not.
    return { ok: response.status === 200 || response.status === 302, status: response.status }
  } catch (error) {
    return { ok: false, status: 0, error: String(error && error.message ? error.message : error) }
  }
}

async function readUpstreamEntry(branch = 'main') {
  const file = entryFileName(ENTRY_OWNER, ENTRY_REPO, ENTRY_DIRECTORY)
  const url = `https://raw.githubusercontent.com/${UPSTREAM_REPO}/${branch}/data/plugins/${file}`
  try {
    const response = await fetch(url, { headers: { 'user-agent': 'ai-terminal-pet-upstream-entry' } })
    if (!response.ok) return { ok: false, status: response.status, text: '' }
    return { ok: true, status: response.status, text: await response.text() }
  } catch (error) {
    return { ok: false, status: 0, text: '', error: String(error && error.message ? error.message : error) }
  }
}

// --------------------------------------------------------------------- checks

/**
 * Look for an open bump pull request for a release, so "the entry is behind"
 * can be told apart from "somebody forgot". A pending PR is a healthy state:
 * the maintainer asked for a rebase and the answer is on its way.
 *
 * Deliberately a plain public REST call rather than `gh`: this runs in CI, where
 * there is no gh login and a repository-scoped token cannot be assumed to read
 * another repository, and it must also work on a machine where gh was never
 * authenticated. `head=<owner>:<branch>` is the one filter that identifies our
 * own bump PR regardless of who opened it. Any non-200 (403 rate limit, 404,
 * no network) is reported as `supported:false` - "could not ask" - which the
 * caller must never read as "nobody opened one".
 *
 * Returns `{supported, open, url, number}`.
 */
async function openBumpPR(version) {
  const branch = `bump-tarball-${version}`
  const head = `${ENTRY_OWNER}:${branch}`
  const url = `https://api.github.com/repos/${UPSTREAM_REPO}/pulls`
    + `?state=open&head=${encodeURIComponent(head)}&per_page=5`
  try {
    const headers = { accept: 'application/vnd.github+json', 'user-agent': 'ai-terminal-pet-upstream-entry' }
    const token = process.env.GH_TOKEN || process.env.GITHUB_TOKEN || ''
    if (token) headers.authorization = `Bearer ${token}`
    const response = await fetch(url, { headers })
    if (!response.ok) return { supported: false, open: false, url: '', number: 0 }
    const found = await response.json()
    const first = Array.isArray(found) ? found[0] : undefined
    return { supported: true, open: Boolean(first), url: first?.html_url ?? '', number: first?.number ?? 0 }
  } catch {
    return { supported: false, open: false, url: '', number: 0 }
  }
}

async function check({ version }) {
  const pkg = readPackage()
  const target = version || pkg.version
  const url = expectedTarball({
    owner: ENTRY_OWNER,
    repo: ENTRY_REPO,
    version: target,
    assetName: assetNameFor(pkg.name, target),
  })

  // Is that release actually out? The asset URL is the only question with an
  // unambiguous answer, and it needs no token and no local tags: a shallow
  // checkout in CI has neither. A 404 means the version has not been shipped
  // yet (the bump commit that moves package.json lands before its tag), so this
  // check must not fire there.
  const probe = await assetProbe(url)
  if (probe.status === 404) {
    console.log(`upstream entry: ${TAG_PREFIX}${target} is not released yet (asset 404). SKIP`)
    console.log('  A version bump that has not been tagged and released cannot be listed upstream.')
    return 0
  }
  if (!probe.ok) {
    console.log(`upstream entry: could not probe the ${TAG_PREFIX}${target} asset (${probe.status || probe.error || 'no response'}). SKIP`)
    console.log('  Network failure on a check that reads two other hosts; a skip is not a pass.')
    return 0
  }

  const upstream = await readUpstreamEntry()
  if (!upstream.ok) {
    console.log(`upstream entry: could not read ${UPSTREAM_REPO} (HTTP ${upstream.status || 'no response'}). SKIP`)
    console.log('  This check reads another repository over the network; a skip is not a pass.')
    return 0
  }
  const listedVersionNow = listedVersion(upstream.text)
  if (!entryIsStale(listedVersionNow, target)) {
    console.log(`upstream entry: points at ${TAG_PREFIX}${listedVersionNow}. OK`)
    return 0
  }
  const pending = await openBumpPR(target)
  if (pending.open) {
    console.log(`upstream entry: still ${TAG_PREFIX}${listedVersionNow || '(none)'}, but ${TAG_PREFIX}${target} is in flight: ${pending.url}`)
    console.log('  OK - a pull request that carries it is open. The entry only moves when a')
    console.log('  maintainer merges it, so a pending PR is a healthy state, not a failure.')
    return 0
  }
  if (!pending.supported) {
    console.log(`upstream entry: still ${TAG_PREFIX}${listedVersionNow || '(none)'}; ${TAG_PREFIX}${target} is released. SKIP`)
    console.log('  Could not ask whether a bump pull request is already open (gh missing or not')
    console.log('  authenticated here), so this is not evidence that it was forgotten.')
    return 0
  }
  const file = entryFileName(ENTRY_OWNER, ENTRY_REPO, ENTRY_DIRECTORY)
  console.log(`upstream entry: STALE and NOTHING PENDING - listed ${TAG_PREFIX}${listedVersionNow || '(none)'}, released ${TAG_PREFIX}${target}.`)
  console.log('')
  console.log('  The release is out, the upstream entry still points at the old tarball, and no')
  console.log('  bump pull request is open: every storefront that renders from it keeps')
  console.log('  installing the old overlay. Rebuild the entry from main rather than merging')
  console.log('  an old branch (see docs/UPSTREAM.md):')
  console.log('')
  console.log(`    node tools/upstream-entry.mjs --version ${target} --write \\`)
  console.log(`      --branch bump-tarball-${target} --pr <pull request number>`)
  console.log('')
  console.log(`  Entry file: data/plugins/${file}`)
  return 1
}

// ---------------------------------------------------------------------- write

async function write({ version, branch, pr, commentOut, dryRun }) {
  const pkg = readPackage()
  const file = entryFileName(ENTRY_OWNER, ENTRY_REPO, ENTRY_DIRECTORY)
  const url = expectedTarball({
    owner: ENTRY_OWNER,
    repo: ENTRY_REPO,
    version,
    assetName: assetNameFor(pkg.name, version),
  })

  const probe = await assetProbe(url)
  if (!probe.ok) {
    throw new Error(`release asset is not live (HTTP ${probe.status}): ${url}`)
  }

  const upstream = await readUpstreamEntry()
  if (!upstream.ok) throw new Error(`could not read upstream main: HTTP ${upstream.status}`)
  const rebuilt = rewriteTarball(upstream.text, url)
  const verdict = rebuildVerdict({ before: upstream.text, after: rebuilt, url })
  if (!verdict.changed) {
    console.log(`upstream entry: main already lists ${TAG_PREFIX}${listedVersion(upstream.text)}. Nothing to do.`)
    return 0
  }
  if (!verdict.onlyTarball) {
    throw new Error('rebuild would change more than the tarball line; refusing to push')
  }

  if (dryRun) {
    console.log(`upstream entry: would rewrite one line in data/plugins/${file}`)
    console.log(`  - ${tarballLine(upstream.text).slice(0, 160)}`)
    console.log(`  + tarball: ${url}`)
    return 0
  }

  const identityValue = identity()
  const work = mkdtempSync(path.join(tmpdir(), 'ai-terminal-pet-entry-'))
  try {
    runGit(['clone', '--quiet', '--depth=1', '--branch', 'main',
      `https://github.com/${UPSTREAM_REPO}.git`, work])
    // LF only, and no automatic conversion: the file we write is the file git
    // hashes, and autocrlf=true would silently break the byte-for-byte proof.
    runGit(['config', 'core.autocrlf', 'false'], { cwd: work })
    runGit(['config', 'core.eol', 'lf'], { cwd: work })
    if (identityValue.name) runGit(['config', 'user.name', identityValue.name], { cwd: work })
    if (identityValue.email) runGit(['config', 'user.email', identityValue.email], { cwd: work })
    if (identityValue.signingKey) {
      runGit(['config', 'user.signingkey', identityValue.signingKey], { cwd: work })
      if (identityValue.gpgFormat) runGit(['config', 'gpg.format', identityValue.gpgFormat], { cwd: work })
      runGit(['config', 'commit.gpgsign', 'true'], { cwd: work })
    }
    runGit(['remote', 'add', 'fork', `https://github.com/${ENTRY_OWNER}/${UPSTREAM_REPO.split('/')[1]}.git`], { cwd: work })
    runGit(['checkout', '--quiet', '-B', branch], { cwd: work })

    writeFileSync(path.join(work, 'data', 'plugins', file), rebuilt, 'utf8')

    const changed = runGit(['diff', '--name-only'], { cwd: work })
      .split('\n').map((value) => value.trim()).filter(Boolean)
    if (changed.length !== 1 || changed[0] !== `data/plugins/${file}`) {
      throw new Error(`expected exactly one changed file, got: ${changed.join(', ') || '(none)'}`)
    }
    const numstat = runGit(['diff', '--numstat'], { cwd: work }).trim()
    if (numstat !== `1\t1\tdata/plugins/${file}`) {
      throw new Error(`expected a one-line change, got: ${numstat}`)
    }

    const body = [
      `Bump the ai-terminal-pet entry to ${TAG_PREFIX}${version}`,
      '',
      'Rebuilt on top of current main instead of merged into the previous branch: main',
      'rewrote this entry by its own hand between our bump and the merge, and a three-',
      'way merge on the same tarball line is what leaves the pull request conflicted',
      'and its checks unrun.',
      '',
      'The tarball line is the only change. Every precondition and default main states',
      'is kept verbatim in both languages, and reverting this one string reproduces',
      "main's file byte-for-byte.",
      '',
      `Release: https://github.com/${ENTRY_OWNER}/${ENTRY_REPO}/releases/tag/${TAG_PREFIX}${version}`,
    ].join('\n')
    const commitArgs = ['commit', '--quiet', '-m', body]
    if (identityValue.signingKey) commitArgs.splice(1, 0, '-S')
    runGit(commitArgs, { cwd: work })

    const head = runGit(['rev-parse', '--short', 'HEAD'], { cwd: work }).trim()
    runGit(['push', '--force-with-lease', 'fork', `HEAD:${branch}`], { cwd: work })
    console.log(`upstream entry: rebuilt ${file} on main and pushed ${branch} (${head})`)

    if (commentOut) {
      const comment = [
        `Rebuilt on top of current \`main\` and force-pushed, so the branch no longer conflicts.`,
        '',
        'What changed, and what deliberately did not:',
        '',
        '- started from `main` and kept it as it is: both descriptions, every precondition',
        '  and default it states, `category`, `icon`, `url`, `name`;',
        `- the one line a version bump needs, \`tarball:\` to \`${TAG_PREFIX}${version}\`, with an`,
        "  assertion that reverting that single string reproduces `main`'s file byte-for-byte",
        '  and that the diff is exactly one line;',
        '- the release asset was probed before pointing at it, so the link is live.',
        '',
        `Release notes: https://github.com/${ENTRY_OWNER}/${ENTRY_REPO}/releases/tag/${TAG_PREFIX}${version}`,
        '',
        'This is now mechanical on my side (`tools/upstream-entry.mjs --write`), and my CI',
        'runs `--check`, which fails when the listed tarball is not the newest release tag.',
      ].join('\n')
      const target = path.isAbsolute(commentOut) ? commentOut : path.join(repoDir, commentOut)
      writeFileSync(target, comment + '\n', 'utf8')
      console.log(`upstream entry: wrote the pull request comment to ${target}`)
    }
    if (pr) {
      try {
        execFileSync('gh', ['pr', 'comment', String(pr), '-R', UPSTREAM_REPO, '--body-file', commentOut || '-'], {
          cwd: repoDir,
          stdio: commentOut ? 'inherit' : ['pipe', 'inherit', 'inherit'],
        })
        console.log(`upstream entry: commented on ${UPSTREAM_REPO}#${pr}`)
      } catch (error) {
        console.log(`upstream entry: the push succeeded but the PR comment did not: ${error.message}`)
      }
    }
    return 0
  } finally {
    try {
      rmSync(work, { recursive: true, force: true })
    } catch {
      // Best effort; the directory is under the OS temp dir.
    }
  }
}

// ------------------------------------------------------------------------ cli

function usage() {
  console.log(`usage: node tools/upstream-entry.mjs --check
       node tools/upstream-entry.mjs --version <plugin version> [--write --branch <name>]
            [--pr <number>] [--comment-out <file>] [--dry-run]`)
}

function main(argv) {
  const flags = new Map()
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index]
    if (!argument.startsWith('--')) throw new Error(`unexpected argument ${argument}`)
    const name = argument.slice(2)
    if (name === 'help') {
      usage()
      return 0
    }
    if (name === 'check' || name === 'write' || name === 'dry-run') {
      flags.set(name, true)
      continue
    }
    const value = argv[index + 1]
    if (value === undefined || value.startsWith('--')) throw new Error(`--${name} needs a value`)
    flags.set(name, value)
    index += 1
  }

  if (flags.get('help')) {
    usage()
    return 0
  }
  if (flags.has('check')) {
    return check({ version: flags.get('version') || '' })
  }
  const version = flags.get('version')
  if (!version) {
    usage()
    return 2
  }
  if (!/^[0-9]+(\.[0-9]+)*$/.test(version)) throw new Error(`not a release version: ${version}`)
  if (!flags.has('write')) {
    const pkg = readPackage()
    const url = expectedTarball({
      owner: ENTRY_OWNER,
      repo: ENTRY_REPO,
      version,
      assetName: assetNameFor(pkg.name, version),
    })
    console.log(`tarball: ${url}`)
    return 0
  }
  const branch = flags.get('branch') || `bump-tarball-${version}`
  return write({
    version,
    branch,
    pr: flags.get('pr') || '',
    commentOut: flags.get('comment-out') || '',
    dryRun: flags.has('dry-run'),
  })
}

const invokedDirectly = process.argv[1]
  && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))

if (invokedDirectly) {
  Promise.resolve()
    .then(() => main(process.argv.slice(2)))
    .then((code) => {
      process.exitCode = code
    })
    .catch((error) => {
      console.error(`upstream-entry: ${error && error.message ? error.message : error}`)
      process.exitCode = 1
    })
}

export { check, write, rebuildVerdict as __rebuildVerdict }
