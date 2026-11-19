#!/usr/bin/env node
/**
 * Print the published overlay version (latest `pet-vX.Y.Z` release tag).
 *
 *   node tools/promo-version.mjs
 *
 * The local `internal/app/version.go` can be ahead of the release (a bump that
 * has not shipped yet). Promotion must talk about what users can download, so
 * this reads the release tag and never the working tree. Exit 1 when the tag
 * cannot be read: the caller must record `unverified`, not guess.
 */

import { execFileSync } from 'node:child_process'

try {
  const tag = execFileSync('gh', [
    'release', 'list',
    '--repo', 'jjjjjjjjnnjnn/ai-terminal-pet',
    '--limit', '30',
    '--json', 'tagName,createdAt',
    '--jq', '[.[] | select(.tagName | startswith("pet-v"))] | sort_by(.createdAt) | last | .tagName'
  ], {
    encoding: 'utf8',
    timeout: 60000
  }).trim()
  const match = /^pet-v(\d+\.\d+\.\d+)$/.exec(tag)
  if (match) {
    process.stdout.write(`${match[1]}\n`)
    process.exit(0)
  }
} catch {
  // fall through
}
process.stderr.write('could not read the latest pet-vX.Y.Z release tag\n')
process.exit(1)
