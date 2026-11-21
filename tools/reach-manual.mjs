/**
 * Which hand-written notes survive into the next docs/reach/latest.md.
 *
 * The generated report already ends with a "## Manual checks (daily, needs the
 * browser)" checklist. Searching latest.md for "\n## Manual check" matches that
 * boilerplate first, so every scheduled run used to append another copy of it
 * (and carried 2026-11-19's numbers forward forever). Only dated headings —
 * "## Manual check (2026-11-19 07:30 UTC)" — are notes a person actually wrote.
 */

export const MANUAL_KEEP_DAYS = 2;

const DATED_HEADING = /^## Manual check \((\d{4}-\d{2}-\d{2})/gm;

export function manualTail(previous, now = new Date()) {
  if (!previous) return "";

  const headings = [...previous.matchAll(DATED_HEADING)];
  if (headings.length === 0) return "";

  const tomorrow = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate() + 1));
  tomorrow.setUTCDate(tomorrow.getUTCDate() - MANUAL_KEEP_DAYS);
  const cutoff = tomorrow.toISOString().slice(0, 10);

  const kept = headings
    .map((match, index) => ({
      date: match[1],
      text: previous.slice(match.index, headings[index + 1]?.index ?? previous.length).trimEnd()
    }))
    .filter((block) => block.date >= cutoff)
    .map((block) => block.text);

  return kept.length ? "\n" + kept.join("\n\n") + "\n" : "";
}
