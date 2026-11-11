"""Verify every relative link and in-document anchor in the docs we wrote.

A launch week is the wrong time to discover that `#connect-your-terminal` does
not exist, so this checks the two things markdown renders without warning: a
relative file path that is not there, and a fragment that names no heading.

Usage: python tools/check_docs_links.py  (from the repository root)
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")

ROOT = Path(__file__).resolve().parent.parent
DOCS = ["README.md", "CONTRIBUTING.md", "CODE_OF_CONDUCT.md", "AGENTS.md", "CHANGELOG.md",
        "docs/HANDOVER.md", "docs/ROADMAP.md", "docs/PROTOCOL.md", "docs/LAUNCH.md",
        "docs/assets/README.md", "packages/dsh-ai-terminal-pet/README.md"]

LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)")
HEADING = re.compile(r"^#{1,6} ")
CODE_FENCE = re.compile(r"```")


def slug(text: str) -> str:
    """GitHub's heading anchor: lowercase, punctuation dropped, spaces to hyphens.

    Deliberately not a full implementation - it covers what our own headings
    contain (letters, digits, hyphens, slashes, backticks, CJK). Anything else
    would be a heading we should rename anyway.
    """
    text = text.strip().lower()
    text = re.sub(r"`", "", text)
    text = re.sub(r"[^\w\u4e00-\u9fff\- ]", "", text)
    return text.replace(" ", "-")


def anchors(path: Path) -> set[str]:
    found = set()
    in_fence = False
    for line in path.read_text(encoding="utf-8").splitlines():
        if CODE_FENCE.match(line.strip()):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if HEADING.match(line):
            found.add(slug(line.lstrip("#")))
    return found


def main() -> int:
    problems = []
    anchor_cache: dict[Path, set[str]] = {}

    for name in DOCS:
        path = ROOT / name
        if not path.is_file():
            problems.append(f"{name}: file is missing")
            continue
        text = path.read_text(encoding="utf-8")
        # Links inside fenced blocks are examples, not links.
        stripped = re.sub(r"```.*?```", "", text, flags=re.S)
        for raw in LINK.findall(stripped):
            if raw.startswith(("http://", "https://", "mailto:")):
                continue
            target, _, fragment = raw.partition("#")
            if not target:
                # Same-file anchor.
                if fragment and fragment not in anchors(path):
                    problems.append(f"{name}: #{fragment} names no heading in this file")
                continue
            resolved = (path.parent / target).resolve()
            if not resolved.exists():
                problems.append(f"{name}: {target} does not exist")
                continue
            if fragment:
                if resolved not in anchor_cache:
                    anchor_cache[resolved] = anchors(resolved)
                if fragment not in anchor_cache[resolved]:
                    problems.append(f"{name}: {target}#{fragment} names no heading in {target}")

    if problems:
        print("link check: %d problem(s)" % len(problems))
        for item in problems:
            print("  " + item)
        return 1
    print("link check: every relative path and anchor resolves (%d files)" % len(DOCS))
    return 0


if __name__ == "__main__":
    sys.exit(main())
