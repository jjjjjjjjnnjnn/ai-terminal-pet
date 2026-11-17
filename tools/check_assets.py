"""Check that docs/assets/README.md's table tells the truth about the files.

The table lists every generated asset with its pixel size, and it is the only
place a reader (or a release note) learns that hero.png is 1600x640 or that
demo.gif is 56 frames. Nothing read it, so it was a third table that could rot
quietly - the same class of drift that `tools/check_facts.mjs` pins for counts
and `tools/check_versions.mjs` pins for versions.

This reads the sizes out of the files themselves (PNG IHDR, GIF logical screen
descriptor and image-descriptor count), so the table has to agree with the bytes
that ship. Stdlib only: the repository's generators need Pillow, this checker
must not.

Usage: python tools/check_assets.py  (from the repository root)
Exit codes: 0 every row agrees, 1 at least one disagrees or is missing.
"""
from __future__ import annotations

import re
import struct
import sys
from pathlib import Path

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")

ROOT = Path(__file__).resolve().parent.parent
ASSETS = ROOT / "docs" / "assets"
README = ASSETS / "README.md"

# | `hero.png` | 1600x640 | README top | ... |
# | `demo.gif` | 760x600, 56 frames | launch posts | ... |
ROW = re.compile(r"^\|\s*`([^`]+\.(?:png|gif))`\s*\|\s*(\d+)x(\d+)(?:,\s*(\d+)\s*frames?)?\s*\|", re.I)


def png_size(path: Path) -> tuple[int, int]:
    header = path.read_bytes()[:24]
    if header[:8] != b"\x89PNG\r\n\x1a\n":
        raise ValueError("not a PNG")
    width, height = struct.unpack(">II", header[16:24])
    return width, height


def gif_info(path: Path) -> tuple[int, int, int]:
    """Logical screen size and how many image descriptors a GIF contains.

    Skipping the block structure properly matters: a naive count of 0x2C bytes
    also counts those bytes inside compressed data, which is how a "frame count"
    silently becomes a random number.
    """
    data = path.read_bytes()
    if data[:6] not in (b"GIF87a", b"GIF89a"):
        raise ValueError("not a GIF")
    width, height = struct.unpack("<HH", data[6:10])
    flags = data[10]
    offset = 13
    if flags & 0x80:  # global colour table
        offset += 3 * (2 ** ((flags & 0x07) + 1))
    frames = 0
    while offset < len(data):
        marker = data[offset]
        offset += 1
        if marker == 0x3B:  # trailer
            break
        if marker == 0x21:  # extension: label + sub-blocks
            offset += 1
            offset = skip_sub_blocks(data, offset)
        elif marker == 0x2C:  # image descriptor
            frames += 1
            if offset + 9 > len(data):
                break
            local_flags = data[offset + 8]
            offset += 9
            if local_flags & 0x80:
                offset += 3 * (2 ** ((local_flags & 0x07) + 1))
            offset = skip_sub_blocks(data, offset + 1)  # +1 = LZW minimum code size
        else:
            raise ValueError("unexpected block 0x%02X at %d" % (marker, offset - 1))
    return width, height, frames


def skip_sub_blocks(data: bytes, offset: int) -> int:
    while offset < len(data):
        length = data[offset]
        offset += 1
        if length == 0:
            return offset
        offset += length
    return offset


def main() -> int:
    if not README.is_file():
        print("assets: %s is missing" % README.relative_to(ROOT))
        return 1
    problems: list[str] = []
    checked = 0
    for line in README.read_text(encoding="utf-8").splitlines():
        match = ROW.match(line.strip())
        if not match:
            continue
        name, width, height = match.group(1), int(match.group(2)), int(match.group(3))
        frames = int(match.group(4)) if match.group(4) else None
        path = ASSETS / name
        if not path.is_file():
            problems.append("%s: listed in the table but not committed" % name)
            continue
        try:
            if path.suffix.lower() == ".png":
                actual_w, actual_h = png_size(path)
                if frames is not None:
                    problems.append("%s: the table claims %d frames but a PNG has none" % (name, frames))
            else:
                actual_w, actual_h, actual_frames = gif_info(path)
                if frames is None:
                    problems.append("%s: the table omits the frame count for a GIF" % name)
                elif actual_frames != frames:
                    problems.append("%s: the table says %d frames, the file has %d" % (name, frames, actual_frames))
        except (ValueError, struct.error) as error:
            problems.append("%s: cannot read the file: %s" % (name, error))
            continue
        if (actual_w, actual_h) != (width, height):
            problems.append("%s: the table says %dx%d, the file is %dx%d" % (name, width, height, actual_w, actual_h))
        checked += 1

    orphans = []
    listed = {ROW.match(line.strip()).group(1) for line in README.read_text(encoding="utf-8").splitlines() if ROW.match(line.strip())}
    for path in sorted(ASSETS.iterdir()):
        if path.suffix.lower() in (".png", ".gif") and path.name not in listed:
            orphans.append(path.name)

    if problems:
        print("assets: %d problem(s)" % len(problems))
        for item in problems:
            print("  " + item)
        return 1
    print("assets: %d table row(s) agree with the files" % checked)
    if orphans:
        # Not a failure: an asset kept for a locale check or a future post is
        # legitimate, but it must not appear in the table by accident either.
        print("assets: %d file(s) committed but not listed: %s" % (len(orphans), ", ".join(orphans)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
