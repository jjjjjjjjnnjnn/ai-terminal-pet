#!/usr/bin/env node
/**
 * Keep this package's copies of the reporter byte-identical to the reference
 * implementation in reporters/.
 *
 * Why this package needs a generator at all
 * -----------------------------------------
 * OpenCode's plugin contract wants `"type": "module"`. The shared reporter core,
 * reporters/protocol.js, is CommonJS (`module.exports` at the bottom), and it has
 * to stay that way: five hooks in reporters/ require() it, and it is deployed to
 * the install directory as protocol.js by scripts/install.ps1. Inside a
 * "type": "module" package that same file would be parsed as ESM and every
 * require() of it would throw ERR_REQUIRE_ESM.
 *
 * So the package carries two generated files:
 *
 *   protocol.cjs  a byte-for-byte copy of reporters/protocol.js (only the
 *                 extension differs, because .cjs is CommonJS regardless of the
 *                 package's type)
 *   reporter.js   reporters/opencode.js with exactly one line changed - the
 *                 specifier of that require() - and a banner saying so
 *
 * Every other byte of the reporter is copied, not rewritten: this is the same
 * rule as packages/dsh-ai-terminal-pet/vendor/protocol.js, and it exists for the
 * same reason. The two implementations drifted once (handover §52.3) and a fix
 * landed in one copy and never reached the terminal. A generator is only worth
 * having if something checks its output, which is what --check is for.
 *
 * Usage:
 *   node scripts/sync.mjs --write    regenerate the two files
 *   node scripts/sync.mjs --check    fail when they do not match the source
 *
 * Exit codes: 0 everything matches (or was written), 1 drift, 2 the source no
 * longer has the shape this tool knows how to copy.
 */
import { readFileSync, writeFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const packageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
export const repoDir = path.resolve(packageDir, '..', '..')

/** The one line this generator is allowed to change, and what it becomes. */
export const SPECIFIER = 'require("./protocol.js")'
export const REPLACEMENT = 'require("./protocol.cjs")'

/**
 * The banner is part of the generated file on purpose: an empty reporter.js with
 * no provenance is how somebody decides to "just fix it here", and that edit is
 * silently overwritten by the next --write.
 */
export const BANNER = [
  '// GENERATED FILE - do not edit. Source: reporters/opencode.js',
  '// The only difference is the protocol specifier below: this package is',
  '// "type": "module", so the shared core has to be required as ./protocol.cjs.',
  '// Regenerate: node packages/opencode-ai-terminal-pet/scripts/sync.mjs --write'
].join('\n')

/** Generated files that are pure copies, with the extension as the only change. */
export const COPIES = [
  { source: 'reporters/protocol.js', target: 'protocol.cjs' },
  // The stale-row sweep is imported by reporter.js with the same specifier in
  // both trees, so it is a pure copy (no require() specifier to rewrite).
  { source: 'reporters/opencode-reconcile.js', target: 'opencode-reconcile.js' },
  // The licence ships inside the package; copying it keeps one text of it.
  { source: 'LICENSE', target: 'LICENSE' }
]

/**
 * The one generated file that is a copy plus a single edited specifier.
 *
 * It is target `reporter.js`, not `server.js`: the package entry point is
 * hand-written because it also has to install the panel on first run, and a
 * generated file is the wrong place for a decision like that.
 */
export const SERVER = { source: 'reporters/opencode.js', target: 'reporter.js' }

/**
 * renderServer turns the reference reporter into this package's entry point.
 *
 * It refuses a source it does not recognise instead of guessing: the failure
 * message has to say "the reporter moved, teach sync.mjs about it", because a
 * generator that silently produces a stale file is worse than no generator.
 */
export function renderServer(source) {
  // The double-generation case is checked first: a source that already says
  // .cjs has zero occurrences of the source specifier, so counting first would
  // report "found 0" and send the reader looking for a require() that is right
  // there in front of them.
  if (source.includes(REPLACEMENT)) {
    throw new Error(`reporters/opencode.js already mentions ${REPLACEMENT}; it is the reference copy and must not.`)
  }
  const occurrences = source.split(SPECIFIER).length - 1
  if (occurrences !== 1) {
    throw new Error(
      `reporters/opencode.js must contain exactly one ${SPECIFIER}, found ${occurrences}. ` +
        'If the require() moved or became conditional, update scripts/sync.mjs rather than hand-editing reporter.js.'
    )
  }
  return `${BANNER}\n${source.replace(SPECIFIER, REPLACEMENT)}`
}

/**
 * expectedFiles returns every generated target with the bytes it should have.
 * Reading is the only side effect, so --check and --write share one answer to
 * "what should be on disk".
 */
export function expectedFiles(root = repoDir) {
  const read = (relative) => readFileSync(path.join(root, relative))
  const files = []
  for (const copy of COPIES) {
    files.push({ target: copy.target, bytes: read(copy.source), source: copy.source })
  }
  files.push({
    target: SERVER.target,
    bytes: Buffer.from(renderServer(readFileSync(path.join(root, SERVER.source), 'utf8')), 'utf8'),
    source: SERVER.source
  })
  return files
}

/**
 * compare returns one line per target whose bytes differ, plus one per missing
 * file, in a shape the CLI and the tests can both assert on.
 */
export function compare(dir = packageDir, root = repoDir) {
  const problems = []
  for (const file of expectedFiles(root)) {
    const target = path.join(dir, file.target)
    let actual
    try {
      actual = readFileSync(target)
    } catch (error) {
      problems.push(`${file.target}: missing (${error.code || error.message}) - run node scripts/sync.mjs --write`)
      continue
    }
    if (!actual.equals(file.bytes)) {
      problems.push(
        `${file.target}: differs from ${file.source} ` +
          `(${actual.length} B on disk, ${file.bytes.length} B expected) - run node scripts/sync.mjs --write`
      )
    }
  }
  return problems
}

export function write(dir = packageDir, root = repoDir) {
  const written = []
  for (const file of expectedFiles(root)) {
    writeFileSync(path.join(dir, file.target), file.bytes)
    written.push(file.target)
  }
  return written
}

function main(argv) {
  const writeMode = argv.includes('--write')
  const checkMode = argv.includes('--check') || !writeMode
  try {
    if (writeMode) {
      const written = write()
      for (const target of written) console.log(`sync: wrote ${target}`)
      if (!checkMode) return 0
    }
    const problems = compare()
    if (problems.length === 0) {
      console.log(`sync: ${expectedFiles().length} generated files match their sources`)
      return 0
    }
    for (const problem of problems) console.error(`sync: ${problem}`)
    return 1
  } catch (error) {
    console.error(`sync: ${error.message}`)
    return 2
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  process.exitCode = main(process.argv.slice(2))
}
