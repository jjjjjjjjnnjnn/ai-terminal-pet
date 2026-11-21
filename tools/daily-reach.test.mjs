import test from "node:test";
import assert from "node:assert/strict";
import { manualTail, MANUAL_KEEP_DAYS } from "./reach-manual.mjs";

const now = new Date("2026-11-21T16:00:00Z");
const BOILERPLATE = `## Manual checks (daily, needs the browser)

- X thread: https://x.com/JJjj1955714/status/2122943194904322144
- HN / V2EX: not posted.
`;
const NOTE = `## Manual check (2026-11-20 07:30 UTC)

- X: no external replies.
- Reddit: score 0, 0 comments.
`;
const OLD_NOTE = `## Manual check (2026-11-19 07:30 UTC)

- Stale numbers: views 74/27, DSH tarball 510.
`;

test("no previous latest.md means nothing to keep", () => {
  assert.equal(manualTail("", now), "");
});

test("the generated checklist is not a manual note", () => {
  // This was the bug: indexOf('\\n## Manual check') matched this heading and
  // appended another copy of the checklist on every single run.
  assert.equal(manualTail(`# Reach\n\n${BOILERPLATE}`, now), "");
});

test("a dated note is kept verbatim with one leading blank line", () => {
  assert.equal(manualTail(`# Reach\n\n${BOILERPLATE}${NOTE}`, now), `\n${NOTE}`);
});

test("a note older than the window is dropped, a current one is kept", () => {
  assert.equal(manualTail(`${OLD_NOTE}`, now), "");
  assert.equal(MANUAL_KEEP_DAYS, 2);
  assert.equal(manualTail(`${NOTE}`, now), `\n${NOTE}`);
});

test("mixed history keeps only the recent note, once", () => {
  const previous = `# Reach\n\n${BOILERPLATE}${BOILERPLATE}${OLD_NOTE}\n${NOTE}`;
  assert.equal(manualTail(previous, now), `\n${NOTE}`);
});

test("rewriting latest.md with the result is stable", () => {
  const before = `# Reach\n\n${BOILERPLATE}`;
  const once = manualTail(before + `\n${NOTE}`, now);
  const twice = manualTail(before + once, now);
  assert.equal(twice, once);
});
