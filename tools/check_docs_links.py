"""Verify every relative link, in-document anchor and local image in the docs we wrote.

A launch week is the wrong time to discover that `#connect-your-terminal` does
not exist, so this checks what markdown renders without warning: a relative file
path that is not there, a fragment that names no heading, and - since 2026-11-16 -
a local `<img src>`, `![](path)` or `<a href>` that points at a file nobody
committed. The image check exists because the README's showcase table pointed at
four PNGs that had been renamed, and the page rendered four broken images for
weeks because this checker only looked at `[links](...)`. The `<a href>` form
exists because the HTML tables in the same documents are a blind spot of exactly
that kind: they render fine in GitHub's preview and 404 for a reader.

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
        "docs/UPSTREAM.md", "docs/PROMOTION.md",
        "docs/assets/README.md", "packages/dsh-ai-terminal-pet/README.md",
        "packages/opencode-ai-terminal-pet/README.md"]

LINK = re.compile(r"\[[^\]]*\]\(([^)\s]+)")
HEADING = re.compile(r"^#{1,6} ")
CODE_FENCE = re.compile(r"```")
# <a id="..."> and id=/name= on any HTML tag. GitHub prefixes author ids with
# "user-content-" and resolves #fragment against that prefix, the same machinery
# its own heading anchors use (a heading renders as <h2>...</h2><a
# id="user-content-install" href="#install">), so an explicit anchor is a real
# destination. Verified against the rendered README, not assumed: docs/ROADMAP.md
# carried the opposite claim for one round (HANDOVER §69.1).
HTML_ANCHOR = re.compile(r"<(?:a|[a-z0-9]+)\b[^>]*\b(?:id|name)=[\"']([^\"']+)[\"']", re.I)
# <img src="..."> and markdown images ![alt](path). Images are the one asset a
# reader notices immediately when it is broken and a link checker never sees.
IMG_TAG = re.compile(r"<img\b[^>]*\bsrc=\"([^\"]+)\"", re.I)
IMG_MD = re.compile(r"!\[[^\]]*\]\(([^)\s]+)")
# <a href="..."> inside an HTML table or a <details> block. Markdown link syntax
# is covered by LINK; this is the other half of what GitHub renders.
HREF_TAG = re.compile(r"<a\b[^>]*\bhref=[\"']([^\"']+)[\"']", re.I)


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
    """Every fragment target a reader can land on: markdown headings and HTML ids.

    Fenced blocks are skipped for headings but not for ids: an `<a id>` inside a
    fence is code, so it is skipped there too by the same in_fence flag.
    """
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
        for match in HTML_ANCHOR.finditer(line):
            found.add(match.group(1))
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
        # Inline code is prose about links, not a link: `<img src="...">` in a
        # sentence explaining this checker must not be reported as a broken
        # image. Blank the spans, keeping the surrounding text.
        stripped = re.sub(r"`[^`\n]*`", "``", stripped)
        # A markdown image is not also a link: strip the whole construct first,
        # or every `![alt](path)` would be reported twice.
        without_images = IMG_MD.sub("", stripped)
        for pattern, kind, body in (
            (IMG_TAG, "image", stripped),
            (IMG_MD, "image", stripped),
            (HREF_TAG, "link", without_images),
            (LINK, "link", without_images),
        ):
            for raw in pattern.findall(body):
                if raw.startswith(("http://", "https://", "mailto:", "data:")):
                    continue
                target, _, fragment = raw.partition("#")
                if not target:
                    # Same-file anchor.
                    if fragment and fragment not in anchors(path):
                        problems.append(f"{name}: #{fragment} names no heading in this file")
                    continue
                resolved = (path.parent / target).resolve()
                if not resolved.exists():
                    problems.append(f"{name}: {kind} {target} does not exist")
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
