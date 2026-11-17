#!/usr/bin/env node
/**
 * Compare a packed .tgz against the tree it was supposed to come from.
 *
 * Why this exists
 * ---------------
 * 1.2.28 shipped a Release asset whose packaged README told users to install
 * `dsh-plugin-v1.2.27`. The order was wrong, not the content: `npm pack` ran
 * *before* the commit that moved the version inside `packages/dsh-ai-terminal-pet/
 * README*.md`, so the tarball froze the pre-bump README - while npm, which
 * re-packs at publish time, served the correct one. Nothing in the working tree
 * looked wrong afterwards, which is exactly why a working-tree check cannot
 * catch this: the only thing that can is looking at the artifact.
 *
 * The rule that follows is in docs/UPSTREAM.md: pack after the version bump,
 * never before, and check the artifact. This is that check.
 *
 * Zero dependencies: gzip is node's zlib, and a tar is 512-byte headers, so the
 * release machine needs nothing that a clean checkout does not already have.
 *
 * Usage: node tools/check-tarball.mjs [path/to/file.tgz] [--package <dir>]
 * Exit codes: 0 every file matches, 1 a difference, 2 the tarball or the tree
 * cannot be read.
 *
 * `--package` exists because there are two npm packages in this repository
 * (packages/dsh-ai-terminal-pet and packages/opencode-ai-terminal-pet) and the
 * rule bought by the 1.2.28 incident is the same for both: compare the
 * artifact, never the working tree.
 */
import { readFileSync, readdirSync, existsSync } from 'node:fs'
import path from 'node:path'
import zlib from 'node:zlib'
import { fileURLToPath } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')
const PACKAGE_DIR = 'packages/dsh-ai-terminal-pet'
const PREFIX = 'package/'

/** readTar returns { entries: Map<name, Buffer>, order: string[] }. */
export function readTar(gzipped) {
  const data = zlib.gunzipSync(gzipped)
  const entries = new Map()
  const order = []
  let offset = 0
  while (offset + 512 <= data.length) {
    const header = data.subarray(offset, offset + 512)
    if (header.every((byte) => byte === 0)) break // end-of-archive marker
    const name = header.subarray(0, 100).toString('utf8').replace(/\0.*$/, '')
    const sizeField = header.subarray(124, 136).toString('utf8').replace(/\0.*$/, '').trim()
    const size = parseInt(sizeField || '0', 8)
    const type = String.fromCharCode(header[156] || 0x30)
    offset += 512
    if (type === '0' || type === '\0') {
      entries.set(name, data.subarray(offset, offset + size))
      order.push(name)
    }
    offset += Math.ceil(size / 512) * 512
  }
  return { entries, order }
}

/** at resolves a package-relative name, accepting an absolute directory. */
function at(directory, name = '') {
  const root = path.isAbsolute(directory) ? directory : path.join(repoDir, directory)
  return name ? path.join(root, name) : root
}

function workingTreeFiles(directory) {
  const result = []
  const walk = (relative) => {
    for (const entry of readdirSync(at(directory, relative), { withFileTypes: true })) {
      const next = relative ? path.join(relative, entry.name) : entry.name
      if (entry.isDirectory()) {
        walk(next)
      } else if (entry.isFile()) {
        result.push(next.split(path.sep).join('/'))
      }
    }
  }
  walk('')
  return result.sort()
}

/** check compares one tarball with the package directory and returns failures. */
export function check(tarball, directory = PACKAGE_DIR) {
  const { entries } = readTar(readFileSync(tarball))
  const failures = []
  const packed = [...entries.keys()]
    .filter((name) => name.startsWith(PREFIX))
    .map((name) => name.slice(PREFIX.length))
    .sort()

  const manifest = JSON.parse(readFileSync(at(directory, 'package.json'), 'utf8'))
  // npm ships exactly the `files` allowlist plus package.json and the files it
  // always includes (README*, LICENSE). A comparator that expected every file in
  // the directory would report verify.mjs and scripts/ as "missing" - which is
  // the tool's first bug, found by its first run against the real package.
  const allowlist = Array.isArray(manifest.files) ? manifest.files : []
  const ships = (name) => {
    if (name === 'package.json') return true
    if (name === 'LICENSE' || /^README(\..*)?\.md$/i.test(name) || /^LICENSE(\..*)?$/i.test(name)) return true
    return allowlist.some((entry) => {
      const trimmed = String(entry).replace(/\/$/, '')
      return name === trimmed || name.startsWith(trimmed + '/')
    })
  }
  const tree = workingTreeFiles(directory).filter((name) => !name.endsWith('.tgz'))

  const missing = tree.filter((name) => ships(name) && !packed.includes(name))
  const extra = packed.filter((name) => !tree.includes(name))
  for (const name of missing) failures.push(`missing from the tarball: ${name}`)
  for (const name of extra) failures.push(`in the tarball but not in ${directory}: ${name}`)

  for (const name of tree) {
    if (name === 'package.json') continue // npm rewrites fields on publish; checked below
    if (!ships(name)) continue
    const packedBytes = entries.get(PREFIX + name)
    if (!packedBytes) continue
    const treeBytes = readFileSync(at(directory, name))
    if (!packedBytes.equals(treeBytes)) {
      failures.push(`differs from ${directory}/${name}: the tarball was packed from an older tree`)
    }
  }

  const packedManifest = entries.get(PREFIX + 'package.json')
  if (!packedManifest) {
    failures.push('the tarball has no package/package.json')
  } else {
    const packedVersion = JSON.parse(packedManifest.toString('utf8')).version
    if (packedVersion !== manifest.version) {
      failures.push(`package.json says ${packedVersion}, the tree says ${manifest.version}`)
    }
  }
  return failures
}

function newestTarball(directory) {
  const root = at(directory)
  const candidates = existsSync(root) ? readdirSync(root).filter((name) => name.endsWith('.tgz')).sort() : []
  if (candidates.length === 0) return null
  return path.join(root, candidates[candidates.length - 1])
}

/**
 * parseArgs accepts one optional tarball and an optional --package directory,
 * because there are two npm packages in this repository now and the rule is the
 * same for both: compare the artifact, not the working tree.
 */
function parseArgs(argv) {
  let tarball
  let directory = PACKAGE_DIR
  for (let index = 0; index < argv.length; index++) {
    const argument = argv[index]
    if (argument === '--package') {
      directory = argv[++index]
    } else if (argument.startsWith('--package=')) {
      directory = argument.slice('--package='.length)
    } else {
      tarball = argument
    }
  }
  return { tarball, directory }
}

function main() {
  const args = parseArgs(process.argv.slice(2))
  const directory = args.directory
  const tarball = args.tarball ? path.resolve(args.tarball) : newestTarball(directory)
  if (!tarball) {
    console.log(`tarball: no .tgz in ${directory} - run "npm pack" there first`)
    process.exit(2)
  }
  let failures
  try {
    failures = check(tarball, directory)
  } catch (error) {
    console.log(`tarball: cannot read ${tarball}: ${error.message}`)
    process.exit(2)
  }
  if (failures.length > 0) {
    console.log(`tarball: ${failures.length} difference(s) between ${path.basename(tarball)} and ${directory}`)
    for (const failure of failures) console.log('  ' + failure)
    console.log('       Pack after the version bump, never before: the artifact is what ships.')
    process.exit(1)
  }
  console.log(`tarball: ${path.basename(tarball)} matches ${directory} file for file`)
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main()
}
