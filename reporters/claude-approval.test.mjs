import test from "node:test";
import assert from "node:assert/strict";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { claudeApprovalAnswer } = require("./protocol.js");

// Claude Code 2.1.312 ignored the legacy {"decision":"allow"} shape on
// PermissionRequest (measured 2026-11-18 against a local Messages API stub).
// These expectations are the contract the Go core pins in
// internal/reporter/claude_decision_test.go; keep the two in step.

test("allow emits hookSpecificOutput.decision.behavior", () => {
  assert.equal(
    claudeApprovalAnswer("allow", "PermissionRequest"),
    '{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"allow"}}}'
  );
});

test("an empty event name defaults to PermissionRequest", () => {
  assert.equal(
    claudeApprovalAnswer("allow"),
    '{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"allow"}}}'
  );
});

test("deny carries the reason the terminal shows", () => {
  assert.equal(
    claudeApprovalAnswer("deny", "PermissionRequest"),
    '{"hookSpecificOutput":{"hookEventName":"PermissionRequest","decision":{"behavior":"deny","message":"Denied from the AI Terminal Pet panel."}}}'
  );
});

test("the legacy decision dialect is never emitted", () => {
  assert.notEqual(claudeApprovalAnswer("allow"), '{"decision":"allow"}');
  assert.notEqual(claudeApprovalAnswer("deny"), '{"decision":"deny"}');
});
