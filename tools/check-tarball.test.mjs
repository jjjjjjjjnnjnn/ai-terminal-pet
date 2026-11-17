import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import zlib from 'node:zlib'

import { check, readTar } from './check-tarball.mjs'

// writeTar builds a minimal ustar archive so the parser is tested against bytes
// it did not produce. Building it here (instead of shelling out to tar) keeps the
// test dependency-free and identical on every platform.
function writeTar(files) {
  const blocks = []
  for (const [name, content] of Object.entries(files)) {
    const body = Buffer.from(content)
    const header = Buffer.alloc(512)
    header.write(name, 0, 100, 'utf8')
    header.write('0000644', 100, 8, 'utf8')
    header.write('0000000', 108, 8, 'utf8')
    header.write('0000000', 116, 8, 'utf8')
    header.write(body.length.toString(8).padStart(11, '0') + ' ', 124, 12, 'utf8')
    header.write('00000000000 ', 136, 12, 'utf8')
    header.write('        ', 148, 8, 'utf8')
    header.write('0', 156, 2, 'utf8')
    header.write('ustar\0', 257, 6, 'utf8')
    header.write('00', 263, 2, 'utf8')
    let sum = 0
    for (const byte of header) sum += byte
    header.write(sum.toString(8).padStart(6, '0') + '\0 ', 148, 8, 'utf8')
    blocks.push(header, body)
    const padding = (512 - (body.length % 512)) % 512
    if (padding > 0) blocks.push(Buffer.alloc(padding))
  }
  blocks.push(Buffer.alloc(1024))
  return zlib.gzipSync(Buffer.concat(blocks))
}

function scratch(files) {
  const directory = mkdtempSync(path.join(os.tmpdir(), 'pet-tarball-'))
  for (const [name, content] of Object.entries(files)) {
    const target = path.join(directory, name)
    mkdirSync(path.dirname(target), { recursive: true })
    writeFileSync(target, content)
  }
  return directory
}

const manifest = (version, files) => JSON.stringify({ name: 'x', version, files })

test('readTar recovers every file and its bytes', () => {
  const gz = writeTar({ 'package/a.txt': 'hello', 'package/sub/b.txt': 'world' })
  const { entries, order } = readTar(gz)
  assert.deepEqual(order, ['package/a.txt', 'package/sub/b.txt'])
  assert.equal(entries.get('package/a.txt').toString(), 'hello')
  assert.equal(entries.get('package/sub/b.txt').toString(), 'world')
})

test('a tarball that matches its directory passes', () => {
  const directory = scratch({
    'README.md': '# readme\n',
    'lib/index.js': 'module.exports = 1\n',
    'package.json': manifest('1.0.0', ['lib'])
  })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({
    'package/README.md': '# readme\n',
    'package/lib/index.js': 'module.exports = 1\n',
    'package/package.json': JSON.stringify({ name: 'x', version: '1.0.0', files: ['lib'], _id: 'x@1.0.0' })
  }))
  assert.deepEqual(check(tarball, directory), [])
  rmSync(directory, { recursive: true, force: true })
})

// This is the shape of the 1.2.28 mistake: the tarball says 1.2.28 inside a
// README while the tree has moved on to 1.2.29.
test('a README packed before the version bump is reported', () => {
  const directory = scratch({
    'README.md': 'add dsh-plugin-v1.2.29\n',
    'package.json': manifest('1.2.29', [])
  })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({
    'package/README.md': 'add dsh-plugin-v1.2.28\n',
    'package/package.json': manifest('1.2.29', [])
  }))
  const failures = check(tarball, directory)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /README\.md: the tarball was packed from an older tree/)
  rmSync(directory, { recursive: true, force: true })
})

test('a version mismatch in package.json is reported', () => {
  const directory = scratch({ 'package.json': manifest('1.2.29', []) })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({ 'package/package.json': manifest('1.2.28', []) }))
  const failures = check(tarball, directory)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /package\.json says 1\.2\.28, the tree says 1\.2\.29/)
  rmSync(directory, { recursive: true, force: true })
})

test('a shipped file missing from the tarball is reported', () => {
  const directory = scratch({
    'a.txt': 'a',
    'b.txt': 'b',
    'package.json': manifest('1', ['a.txt', 'b.txt'])
  })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({ 'package/a.txt': 'a', 'package/package.json': manifest('1', ['a.txt', 'b.txt']) }))
  const failures = check(tarball, directory)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /missing from the tarball: b\.txt/)
  rmSync(directory, { recursive: true, force: true })
})

// The tool's own first bug: scripts/ and verify.mjs live in the package
// directory but are deliberately not shipped, so they must not be "missing".
test('files outside the files allowlist are not expected in the tarball', () => {
  const directory = scratch({
    'lib/index.js': 'x',
    'scripts/vendor.mjs': 'not shipped',
    'verify.mjs': 'not shipped',
    'package.json': manifest('1', ['lib'])
  })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({
    'package/lib/index.js': 'x',
    'package/package.json': manifest('1', ['lib'])
  }))
  assert.deepEqual(check(tarball, directory), [])
  rmSync(directory, { recursive: true, force: true })
})

test('a file in the tarball that is not in the tree is reported', () => {
  const directory = scratch({ 'a.txt': 'a', 'package.json': manifest('1', ['a.txt']) })
  const tarball = path.join(directory, 'x.tgz')
  writeFileSync(tarball, writeTar({
    'package/a.txt': 'a',
    'package/stray.txt': 'stray',
    'package/package.json': manifest('1', ['a.txt'])
  }))
  const failures = check(tarball, directory)
  assert.equal(failures.length, 1)
  assert.match(failures[0], /in the tarball but not in/)
  rmSync(directory, { recursive: true, force: true })
})

test('an unreadable tarball is a clear error rather than a crash', () => {
  assert.throws(() => readTar(Buffer.from('not a tarball')))
})
