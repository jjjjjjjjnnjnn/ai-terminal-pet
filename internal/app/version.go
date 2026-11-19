package app

// The version constant lives in version.go so that every platform, and every
// command that prints it, agrees without a build tag. See docs/UPSTREAM.md for the
// release checklist that bumps it; tools/check_versions.mjs parses this file.
//
// This file has no build tag: a platform without a window layer still has a version.
const Version = "1.0.24"
