// Package main is the single-file installer.
//
// The payload lives in this directory as commit-placeholders only. See
// payload/README.md: the real files are staged by scripts/package.ps1 and are
// never committed, because they are the same 5 MB of binaries the portable zip
// carries and the repository already ignores those.
package main

import "embed"

// payloadFS embeds the staged release payload.
//
// `all:payload` rather than `payload`: without the `all:` prefix an embed
// pattern skips files whose names start with "." or "_", and the staging marker
// is exactly that kind of name. Reading the marker is how this program tells
// "staged by scripts/package.ps1" from "a plain go build over a checkout", so a
// pattern that silently dropped it would turn every release build into the
// refuse-to-run path.
//
// A checkout that never ran the packaging step therefore compiles to a working
// binary that refuses to install anything. That is on purpose: a setup exe with
// a half-empty payload would look like an installer and behave like a
// data-loss story. There is no fallback to the portable zip, because the zip is
// a build output and is not in the repository either.
//
//go:embed all:payload
var payloadFS embed.FS
