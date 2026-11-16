#!/usr/bin/env node
/**
 * Fail when a document states a count that the code no longer produces.
 *
 * Why this exists
 * ---------------
 * tools/check_versions.mjs watches version strings and found four stale ones in
 * one afternoon. Round 68 fixed the README's numbers by hand after a reader
 * noticed four broken images and a 425/461/382 test count - and wrote down the
 * lesson "numbers come from a command, not from memory". One round later the
 * same class of drift was already back, in docs/ROADMAP.md's baseline table:
 *
 *   - "428 个 func Test" while the code had 463 (the same file said 463 one
 *     section lower);
 *   - "87 test files / 12,673 lines" and "136 source files / 26,526 lines"
 *     while the tree had 94 / 13,493 and 143 / 27,200;
 *   - a 4,702,208-byte overlay while bin/ held 4,717,568.
 *
 * A number that a person copies is a number that rots. This tool reads the
 * claims that CAN be measured without building anything, and requires each to
 * equal what the tree says today.
 *
 * What it pins (and why exactly these)
 * ------------------------------------
 *   - `func Test` declarations: what "463 func Test" means is a rule, not a
 *     count, so the rule lives here and the documents must agree with it. It
 *     counts lines starting with `func Test`, which includes `TestMain`; that
 *     is the definition every historical count used.
 *   - packages with tests: a package counts when its directory holds at least
 *     one `*_test.go`. Deliberately not `go list`: a Windows-only test file
 *     would make the same claim measure differently on ubuntu-latest, and this
 *     document is about the tree, not about one OS's build.
 *   - all Go packages: the non-comment lines of scripts/portable-packages.txt,
 *     which is also the cross-build gate's list.
 *   - `--doctor` checks: the `runCheck("...")` entries in internal/app/doctor.go,
 *     so adding a check without touching the README is red.
 *
 * What it deliberately does NOT pin: file/line counts and binary sizes. Those
 * change with every edit and with the toolchain (`go 1.23.0` is a floor, so CI
 * legitimately builds different bytes than the release machine), so a gate
 * would either be red constantly or would be taught to ignore itself. They are
 * printed by `--print` for whoever is writing a baseline row, with a date.
 *
 * Usage: node tools/check_facts.mjs           (verify; exit 1 on any mismatch)
 *        node tools/check_facts.mjs --print   (measure and print, never fails
 *                                              on a mismatch; exit 2 only when
 *                                              a file it must read is missing)
 */
import { readFileSync, readdirSync, statSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repoDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..')

const README = 'README.md'
const AGENTS = 'AGENTS.md'
const ROADMAP = 'docs/ROADMAP.md'
const DOCTOR = 'internal/app/doctor.go'
const PACKAGE_LIST = 'scripts/portable-packages.txt'

/** Directories that are not this repository's own Go sources. */
const SKIP_DIRS = new Set(['.git', 'node_modules', 'vendor', 'bin', 'dist', '.local'])

/**
 * Every known "this is how many" shape.
 *
 * `expects` names the measured facts, one per capture group, so a claim can
 * state two counts in one sentence ("463 func Test across 15 packages") without
 * needing two patterns that could drift apart.
 */
export const FACTS = [
  {
    id: 'readme en: func Test count and packages',
    file: README,
    pattern: /\*\*(\d+) `func Test`\*\* across (\d+) packages/g,
    expects: ['tests', 'testedPackages'],
    count: 1
  },
  {
    id: 'readme zh: func Test count and packages',
    file: README,
    pattern: /(\d+) 个包共 \*\*(\d+) 个 `func Test`\*\*/g,
    expects: ['testedPackages', 'tests'],
    count: 1
  },
  {
    id: 'readme en: all Go packages',
    file: README,
    pattern: /compiles all (\d+) Go packages/g,
    expects: ['packages'],
    count: 1
  },
  {
    id: 'readme zh: all Go packages',
    file: README,
    pattern: /通过 (\d+) 个 Go 包/g,
    expects: ['packages'],
    count: 1
  },
  {
    id: 'readme en: doctor checks',
    file: README,
    pattern: /(\d+) checks — install files/g,
    expects: ['doctorChecks'],
    count: 1
  },
  {
    id: 'readme zh: doctor checks',
    file: README,
    pattern: /共 \*\*(\d+) 项\*\*：安装文件/g,
    expects: ['doctorChecks'],
    count: 1
  },
  {
    id: 'agents: func Test count and packages',
    file: AGENTS,
    pattern: /\((\d+) func Test across (\d+) packages\)/g,
    expects: ['tests', 'testedPackages'],
    count: 1
  },
  {
    id: 'agents: cross-platform package count',
    file: AGENTS,
    pattern: /(\d+) packages must build for linux\/amd64 and darwin\/arm64/g,
    expects: ['packages'],
    count: 1
  },
  {
    id: 'roadmap baseline: func Test count in the code-volume row',
    file: ROADMAP,
    within: '## 0.',
    pattern: /（\*\*(\d+)\*\* 个 `func Test`）/g,
    expects: ['tests'],
    count: 1
  },
  {
    id: 'roadmap baseline: func Test count and packages in the gates row',
    file: ROADMAP,
    within: '## 0.',
    pattern: /\*\*(\d+) 个 `func Test`\*\* 分布在 (\d+) 个包/g,
    expects: ['tests', 'testedPackages'],
    count: 1
  },
  {
    id: 'roadmap baseline: doctor checks',
    file: ROADMAP,
    within: '## 0.',
    pattern: /\*\*(\d+) 项\*\*（environment/g,
    expects: ['doctorChecks'],
    count: 1
  }
]

// ---------------------------------------------------------------- pure parts

/**
 * withinSection returns one `## heading` section of a document, up to the next
 * `## ` heading. The ROADMAP keeps a dated baseline table above years of
 * history, and the history is full of numbers that were true when they were
 * written; only the baseline is a claim about today.
 */
export function withinSection(text, heading) {
  const lines = text.split('\n')
  const start = lines.findIndex((line) => line.startsWith(heading))
  if (start < 0) return null
  let end = lines.length
  for (let index = start + 1; index < lines.length; index += 1) {
    if (lines[index].startsWith('## ')) {
      end = index
      break
    }
  }
  return lines.slice(start, end).join('\n')
}

/**
 * sectionStartLine is withinSection's line offset, so a failure inside a
 * section still reports the line number a reader would find in the file.
 */
export function sectionStartLine(text, heading) {
  const index = text.split('\n').findIndex((line) => line.startsWith(heading))
  return index < 0 ? null : index
}

function lineOf(text, index) {
  let line = 1
  for (let cursor = 0; cursor < index; cursor += 1) {
    if (text.charCodeAt(cursor) === 10) line += 1
  }
  return line
}

/** countLines counts lines, ignoring the empty string after a trailing newline. */
export function countLines(text) {
  if (text === '') return 0
  const lines = text.split('\n')
  if (lines[lines.length - 1] === '') lines.pop()
  return lines.length
}

/**
 * countFuncTests counts `func Test...` declarations. sources is a list of
 * { file, text }, already filtered to *_test.go files, so the rule is visible
 * and testable instead of living inside a shell pipeline.
 */
export function countFuncTests(sources) {
  let count = 0
  for (const source of sources) {
    for (const line of source.text.split('\n')) {
      if (line.startsWith('func Test')) count += 1
    }
  }
  return count
}

/**
 * listedPackages reads scripts/portable-packages.txt: one package per line,
 * `#` comments and blank lines ignored.
 */
export function listedPackages(text) {
  return text
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line !== '' && !line.startsWith('#'))
}

/**
 * countDoctorChecks counts the named checks in the literal slice in
 * internal/app/doctor.go. Only `runCheck("name", ...)` is matched, so the
 * `func runCheck(` definition - which takes a variable name - is not counted.
 */
export function countDoctorChecks(goSource) {
  return [...goSource.matchAll(/\brunCheck\("([a-z0-9 ]+)",/g)].length
}

/**
 * checkFacts compares every claim against the measured facts. It returns
 * human-readable failures and never throws, so the CLI and the unit tests read
 * the same verdict.
 */
export function checkFacts(facts, documents, claims = FACTS) {
  const failures = []
  for (const claim of claims) {
    let text = documents[claim.file]
    if (text === undefined) {
      failures.push(`fail: ${claim.id}: ${claim.file} was not read`)
      continue
    }
    let lineOffset = 0
    if (claim.within) {
      const section = withinSection(text, claim.within)
      if (section === null) {
        failures.push(`fail: ${claim.id}: ${claim.file} has no "${claim.within}" section`)
        continue
      }
      lineOffset = sectionStartLine(text, claim.within)
      text = section
    }
    const matches = [...text.matchAll(claim.pattern)]
    if (matches.length !== claim.count) {
      failures.push(
        `fail: ${claim.id}: found ${matches.length} match(es) in ${claim.file}, expected ${claim.count}` +
          (matches.length === 0 ? ' (reworded? then update FACTS in tools/check_facts.mjs)' : '')
      )
      continue
    }
    for (const match of matches) {
      claim.expects.forEach((fact, group) => {
        const found = Number(match[group + 1])
        const want = facts[fact]
        if (want === undefined) {
          failures.push(`fail: ${claim.id}: no measured value for "${fact}"`)
        } else if (found !== want) {
          failures.push(
            `fail: ${claim.id}: ${claim.file}:${lineOffset + lineOf(text, match.index)} says ${found}, ` +
              `the tree says ${want}`
          )
        }
      })
    }
  }
  return failures
}

// -------------------------------------------------------------- measurement

function walk(dir, visit) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (SKIP_DIRS.has(entry.name)) continue
      walk(path.join(dir, entry.name), visit)
    } else if (entry.isFile()) {
      visit(path.join(dir, entry.name))
    }
  }
}

function measure(repoDirectory = repoDir) {
  const packages = listedPackages(readFileSync(path.join(repoDirectory, PACKAGE_LIST), 'utf8'))
  const goFiles = []
  walk(repoDirectory, (file) => {
    if (file.endsWith('.go')) goFiles.push(file)
  })

  const testSources = []
  let testLines = 0
  let sourceFiles = 0
  let sourceLines = 0
  for (const file of goFiles) {
    const text = readFileSync(file, 'utf8')
    const relative = path.relative(repoDirectory, file).split(path.sep).join('/')
    if (relative.endsWith('_test.go')) {
      testSources.push({ file: relative, text })
      testLines += countLines(text)
    } else {
      sourceFiles += 1
      sourceLines += countLines(text)
    }
  }

  const testedPackages = packages.filter((pkg) => {
    try {
      return readdirSync(path.join(repoDirectory, pkg)).some((name) => name.endsWith('_test.go'))
    } catch {
      return false
    }
  })

  return {
    tests: countFuncTests(testSources),
    testFiles: testSources.length,
    testLines,
    sourceFiles,
    sourceLines,
    packages: packages.length,
    testedPackages: testedPackages.length,
    doctorChecks: countDoctorChecks(readFileSync(path.join(repoDirectory, DOCTOR), 'utf8'))
  }
}

/** readDocuments reads every document a claim may point at, reporting the ones that are missing. */
export function readDocuments(repoDirectory = repoDir) {
  const documents = {}
  const missing = []
  for (const file of [README, AGENTS, ROADMAP]) {
    try {
      documents[file] = readFileSync(path.join(repoDirectory, file), 'utf8')
    } catch {
      missing.push(file)
    }
  }
  return { documents, missing }
}

export { measure }

function binarySizes() {
  const report = []
  for (const name of ['AI-Terminal-Pet.exe', 'pet-statusline.exe', 'pet-hook.exe']) {
    try {
      report.push(`${name}: ${statSync(path.join(repoDir, 'bin', name)).size} B`)
    } catch {
      report.push(`${name}: not built (run the build steps in AGENTS.md)`)
    }
  }
  return report
}

function main() {
  const printOnly = process.argv.includes('--print')
  const { documents, missing } = readDocuments()
  if (missing.length > 0) {
    console.log(`facts: cannot read ${missing.join(', ')} - refusing to guess`)
    process.exit(2)
  }

  let facts
  try {
    facts = measure()
  } catch (error) {
    console.log(`facts: cannot measure the tree: ${error.message}`)
    process.exit(2)
  }

  if (printOnly) {
    console.log(`facts measured on ${process.platform} (${new Date().toISOString().slice(0, 10)}):`)
    console.log(`  packages (scripts/portable-packages.txt): ${facts.packages}`)
    console.log(`  packages with tests: ${facts.testedPackages}`)
    console.log(`  func Test declarations: ${facts.tests}`)
    console.log(`  test files / lines: ${facts.testFiles} / ${facts.testLines.toLocaleString('en-US')}`)
    console.log(`  source files / lines: ${facts.sourceFiles} / ${facts.sourceLines.toLocaleString('en-US')}`)
    console.log(`  --doctor checks: ${facts.doctorChecks}`)
    for (const line of binarySizes()) console.log(`  ${line}`)
    console.log('\nWrite the counts with a date and the command that produced them, never from memory.')
    return
  }

  const failures = checkFacts(facts, documents)
  for (const failure of failures) console.log(failure)
  if (failures.length > 0) {
    console.log(`\nfacts: ${failures.length} claim(s) disagree with the tree`)
    console.log('       Fix the document, or update FACTS in tools/check_facts.mjs if the sentence was reworded.')
    process.exit(1)
  }
  console.log(
    `facts: ${FACTS.length} claim shapes agree - ${facts.tests} func Test across ${facts.testedPackages} ` +
      `packages, ${facts.packages} Go packages, ${facts.doctorChecks} --doctor checks`
  )
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main()
}
