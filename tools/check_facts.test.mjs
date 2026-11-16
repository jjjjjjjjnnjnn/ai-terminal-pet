import { test } from 'node:test'
import assert from 'node:assert/strict'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import {
  FACTS,
  checkFacts,
  countDoctorChecks,
  countFuncTests,
  countLines,
  listedPackages,
  measure,
  readDocuments,
  sectionStartLine,
  withinSection
} from './check_facts.mjs'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

// ------------------------------------------------------------- measurements

test('func Test counting counts declarations, not calls or comments', () => {
  const sources = [
    {
      file: 'a_test.go',
      text: [
        'package a',
        '',
        'func TestOne(t *testing.T) {}',
        'func TestMain(m *testing.M) {}',
        'func BenchmarkThing(b *testing.B) {}',
        '// func TestCommented(t *testing.T) {}',
        'func helperTest(t *testing.T) {}',
        'func TestTwo(t *testing.T) {}'
      ].join('\n')
    },
    { file: 'b_test.go', text: 'func TestThree(t *testing.T) {}\n' }
  ]
  // The rule is the `func Test` prefix, which is what every historical count
  // used - TestMain included, a commented line and a Benchmark excluded.
  assert.equal(countFuncTests(sources), 4)
})

test('countLines ignores the empty string after a trailing newline', () => {
  assert.equal(countLines(''), 0)
  assert.equal(countLines('a\n'), 1)
  assert.equal(countLines('a\nb\n'), 2)
  assert.equal(countLines('a\nb'), 2)
})

test('listedPackages skips comments and blank lines', () => {
  const text = ['# a comment', '', 'cmd/one', '  internal/two  ', '   ', '# trailing'].join('\n')
  assert.deepEqual(listedPackages(text), ['cmd/one', 'internal/two'])
})

test('doctor checks are the named runCheck entries, not the definition', () => {
  const source = [
    'func runCheck(name string, check func() DoctorCheck) (result DoctorCheck) {',
    '  return check()',
    '}',
    'func checks() []DoctorCheck {',
    '  return []DoctorCheck{',
    '    runCheck("environment", func() DoctorCheck { return DoctorCheck{} }),',
    '    runCheck("install files", func() DoctorCheck { return DoctorCheck{} }),',
    '    runCheck(computed, func() DoctorCheck { return DoctorCheck{} })',
    '  }',
    '}'
  ].join('\n')
  // Two string-literal names. The third call passes a variable, which is not a
  // check anybody can name in a document, so it must not inflate the count.
  assert.equal(countDoctorChecks(source), 2)
})

// ------------------------------------------------------------------ sections

const sectioned = [
  '## 0. baseline',
  '',
  '| tests | **428** 个 `func Test` |',
  '',
  '## 1. history',
  '',
  '| tests | **382** 个 `func Test` |'
].join('\n')

test('withinSection keeps only the named section', () => {
  const section = withinSection(sectioned, '## 0.')
  assert.match(section, /428/)
  assert.doesNotMatch(section, /382/)
})

test('sectionStartLine is the offset failures are reported against', () => {
  assert.equal(sectionStartLine(sectioned, '## 0.'), 0)
  assert.equal(sectionStartLine(sectioned, '## 1.'), 4)
  assert.equal(sectionStartLine(sectioned, '## 9.'), null)
  assert.equal(withinSection(sectioned, '## 9.'), null)
})

// -------------------------------------------------------------------- claims

const claim = {
  id: 'test claim',
  file: 'doc.md',
  pattern: /(\d+) checks — install files/g,
  expects: ['doctorChecks'],
  count: 1
}

test('a claim that agrees with the tree passes', () => {
  const documents = { 'doc.md': '13 checks — install files, version' }
  assert.deepEqual(checkFacts({ doctorChecks: 13 }, documents, [claim]), [])
})

test('a stale number fails, naming the file and the real line', () => {
  const documents = { 'doc.md': 'skip\nskip\n12 checks — install files, version' }
  const failures = checkFacts({ doctorChecks: 13 }, documents, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md:3 says 12, the tree says 13/)
})

test('a section-scoped claim reports the line number in the whole file', () => {
  const scoped = {
    id: 'section claim',
    file: 'roadmap.md',
    within: '## 1.',
    pattern: /\*\*(\d+)\*\* 个 `func Test`/g,
    expects: ['tests'],
    count: 1
  }
  const documents = { 'roadmap.md': sectioned }
  const failures = checkFacts({ tests: 463 }, documents, [scoped])
  assert.equal(failures.length, 1)
  // The number sits on line 7 of the file (1-based), not line 3 of the section.
  assert.match(failures[0], /roadmap\.md:7 says 382, the tree says 463/)
})

test('a reworded sentence is reported, not silently skipped', () => {
  const documents = { 'doc.md': 'thirteen checks - install files' }
  const failures = checkFacts({ doctorChecks: 13 }, documents, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /found 0 match\(es\).*update FACTS/)
})

test('a document that was not read is a failure, not a pass', () => {
  const failures = checkFacts({ doctorChecks: 13 }, {}, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /doc\.md was not read/)
})

test('a missing measured fact is a failure, not a pass', () => {
  const documents = { 'doc.md': '13 checks — install files' }
  const failures = checkFacts({}, documents, [claim])
  assert.equal(failures.length, 1)
  assert.match(failures[0], /no measured value for "doctorChecks"/)
})

// ------------------------------------------------------- the real repository

test('every claim shape still matches the documents it watches', () => {
  // This is the "new gate's first run" lesson applied: the patterns above are
  // prose, and a sentence can be reworded without any number going stale. If
  // this fails, either the sentence moved or the claim should be updated.
  const { documents, missing } = readDocuments(repoDir)
  assert.deepEqual(missing, [])
  const failures = checkFacts(measure(repoDir), documents)
  assert.deepEqual(failures, [])
})

test('the shipped claim shapes cover every document they name', () => {
  const files = new Set(FACTS.map((entry) => entry.file))
  assert.deepEqual([...files].sort(), ['AGENTS.md', 'README.md', 'docs/ROADMAP.md'])
})
