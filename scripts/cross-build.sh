#!/bin/bash
# Local reproduction of CI's cross-platform gate, so the two cannot disagree.
#
# The gate itself lives in .github/workflows/ci.yml because that is where it runs; this
# script is the same logic for a developer (and for the agent that wrote it) to run
# before pushing. If they ever differ, CI is the one that decides.

set -e
# Repository root = the parent of this script's directory. `cd "$(dirname "$0")/../.."`
# was wrong when the script is invoked as "scripts/cross-build.sh" (that resolves one
# level too high, and the lists then appear missing).
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root" || exit 1

failed=0
while IFS= read -r pkg; do
  case "$pkg" in ''|'#'*) continue ;; esac
  for target in linux/amd64 darwin/arm64; do
    os="${target%%/*}"
    arch="${target##*/}"
    if ! GOOS="$os" GOARCH="$arch" GOFLAGS=-mod=readonly CGO_ENABLED=0 go build "./$pkg" >/tmp/cross.log 2>&1; then
      echo "cross: ./$pkg does not build for $target:"
      cat /tmp/cross.log
      failed=1
    fi
  done
done < scripts/portable-packages.txt
[ "$failed" -eq 0 ] && echo "cross: every listed package builds for linux/amd64 and darwin/arm64"

failed=0
while IFS= read -r pkg; do
  case "$pkg" in ''|'#'*) continue ;; esac
  for target in linux/amd64 darwin/arm64; do
    os="${target%%/*}"
    arch="${target##*/}"
    if GOOS="$os" GOARCH="$arch" GOFLAGS=-mod=readonly CGO_ENABLED=0 go build "./$pkg" >/dev/null 2>&1; then
      echo "cross: ./$pkg unexpectedly builds for $target; move it to scripts/portable-packages.txt if that is now true"
      failed=1
    fi
  done
done < scripts/windows-only-packages.txt
[ "$failed" -eq 0 ] && echo "cross: the windows-only list still matches reality"
exit "$failed"
