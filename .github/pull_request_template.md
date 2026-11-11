## What changed

<!-- One or two sentences. If it fixes something a user would notice, say what they saw before this change. -->

## Why

<!-- The reason, not a restatement of the diff. Link the issue if there is one. -->

## Gates

<!-- CI runs exactly these. Paste the summary lines, not the whole log. -->

- [ ] `gofmt -l .` prints nothing
- [ ] `go vet ./...` prints nothing
- [ ] `go test -count=1 ./...` all packages ok
- [ ] each reporter passes its own `node --check` (one file at a time - the command only validates its first argument)
- [ ] `node integrations/deepseek-harness/verify.mjs`
- [ ] `node packages/dsh-ai-terminal-pet/scripts/vendor.mjs --check`
- [ ] `node packages/dsh-ai-terminal-pet/verify.mjs`
- [ ] `node tests/fresh-machine.mjs --no-overlay`
- [ ] `.\scripts\install.cmd -VerifyOnly` reports 11 files matching

## Does this touch a release-immutable file?

<!-- packages/dsh-ai-terminal-pet/README.md, its two package.json files, vendor/, or
     docs/market-index.json's versions[0]. If yes, this cannot ship without a new version
     number, a new tag and a new Release - say so here so the release is planned rather
     than discovered. -->

- [ ] No
- [ ] Yes - a version bump is required (describe which files)

## Notes for the reviewer

<!-- Anything non-obvious: a bug whose cause was not where the symptom was, a decision
     that looks arbitrary, a number that was measured rather than remembered. -->
