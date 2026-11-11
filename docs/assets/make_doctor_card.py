"""Render the terminal output a healthy install produces, as a real capture.

This one is not a mockup: `--doctor` is deterministic text, so the image is the
program's own stdout, captured and typeset. The only thing added is syntax
colouring (`[ok]` in green) and the window chrome around it, and a caption that
says the command.

Usage, from the repository root:

    & "$env:LOCALAPPDATA\\AI-Terminal-Pet\\AI-Terminal-Pet.exe" --doctor > doctor.txt
    python docs/assets/make_doctor_card.py doctor.txt
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")

try:
    from PIL import Image, ImageDraw, ImageFont
except ImportError:
    print("缺少 Pillow 依赖。\nMissing Pillow: pip install pillow", file=sys.stderr)
    sys.exit(2)

ASSETS = Path(__file__).resolve().parent
BG = (0x14, 0x16, 0x1D)
BAR = (0x22, 0x25, 0x2E)
EDGE = (0x30, 0x34, 0x3F)
OK_GREEN = (0x8F, 0xD6, 0x9A)
WARN_AMBER = (0xFF, 0xC8, 0x6A)
FAIL_RED = (0xE0, 0x6C, 0x60)
INK = (0xC8, 0xD0, 0xDC)
MUTED = (0x6B, 0x73, 0x82)
ACCENT = (0xFF, 0xC8, 0x6A)


def mono(size: int) -> ImageFont.FreeTypeFont:
    for name in ("consola.ttf", "cour.ttf"):
        try:
            return ImageFont.truetype(str(Path("C:/Windows/Fonts") / name), size)
        except OSError:
            continue
    return ImageFont.load_default()


def ui(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    for name in (("segoeuib.ttf" if bold else "segoeui.ttf"),):
        try:
            return ImageFont.truetype(str(Path("C:/Windows/Fonts") / name), size)
        except OSError:
            continue
    return ImageFont.load_default()


def colour(line: str):
    if line.startswith("[ok]"):
        return OK_GREEN, None
    if line.startswith("[warn]"):
        return WARN_AMBER, None
    if line.startswith("[fail]"):
        return FAIL_RED, None
    if line.startswith("doctor:"):
        return ACCENT, True
    return INK, None


def read_report(source: Path) -> list[str]:
    """Read the report whatever encoding the shell that captured it used.

    `>` in Windows PowerShell 5.1 writes UTF-16LE, `>` in PowerShell 7 writes
    UTF-8, and a copy pasted into a file can be either. Guessing wrong turns a
    real capture into a mojibake card, so both are tried in order.
    """
    raw = source.read_bytes()
    for encoding in ("utf-8", "utf-16", "utf-16-le"):
        try:
            text = raw.decode(encoding)
            break
        except UnicodeDecodeError:
            continue
    else:
        fail("cannot decode %s as UTF-8 or UTF-16" % source)
    # A BOM that survived the utf-8 attempt shows up as \ufeff at the start.
    text = text.lstrip("\ufeff")
    return [line.rstrip("\r") for line in text.splitlines() if line.strip()]


def redact(lines: list[str], home: str) -> list[str]:
    """Replace the capturing machine's user directory with <you>.

    This image goes on the public internet, and the report is full of
    `C:\\Users\\<name>\\...` - three occurrences per line on several lines. The
    report is otherwise unchanged, including every number: the point of showing
    it is that these are real numbers from a real run, so nothing but the
    account name is touched.
    """
    if not home:
        return lines
    replacements = [home, home.replace("\\", "/"), home.lower(), home.lower().replace("\\", "/")]
    out = []
    for line in lines:
        for value in replacements:
            line = line.replace(value, "<you>")
        out.append(line)
    return out


def fail(message: str) -> None:
    print(message, file=sys.stderr)
    sys.exit(2)


def build(source: Path, home: str = "") -> Path:
    lines = redact(read_report(source), home)
    if not lines:
        print("input file is empty", file=sys.stderr)
        sys.exit(2)

    body_font = mono(15)
    width = 300
    for line in lines:
        width = max(width, int(body_font.getlength(line)))
    cmd = '& "$env:LOCALAPPDATA\\AI-Terminal-Pet\\AI-Terminal-Pet.exe" --doctor'
    cmd_font = mono(14)
    width = max(width, int(cmd_font.getlength(cmd)))
    inner = width
    pad_x, pad_y = 22, 18
    W = inner + pad_x * 2 + 2
    H = 44 + pad_y * 2 + len(lines) * 22 + 58

    base = Image.new("RGB", (W, H), BG)
    draw = ImageDraw.Draw(base)
    draw.rounded_rectangle([0, 0, W - 1, H - 1], radius=10, outline=EDGE)
    draw.rounded_rectangle([1, 1, W - 2, 42], radius=10, fill=BAR)
    draw.rectangle([1, 30, W - 2, 42], fill=BAR)
    for index in range(3):
        draw.ellipse([16 + index * 16, 15, 24 + index * 16, 23], fill=(0x5A, 0x5F, 0x6B))
    draw.text((74, 11), "Windows PowerShell  -  AI-Terminal-Pet --doctor", font=ui(14), fill=(0xB4, 0xBC, 0xCA))

    y = 42 + pad_y
    draw.text((pad_x + 1, y), "PS> ", font=cmd_font, fill=(0x8F, 0xD6, 0x9A))
    draw.text((pad_x + 1 + int(cmd_font.getlength("PS> ")), y), cmd, font=cmd_font, fill=INK)
    y += 30

    for line in lines:
        colour_value, bold = colour(line)
        draw.text((pad_x + 1, y), line, font=body_font, fill=colour_value)
        y += 22

    # A closing line, in the same voice as the report: what the numbers mean.
    y += 8
    draw.text((pad_x + 1, y), "the overlay, the pipe and the install are healthy -",
          font=ui(13), fill=MUTED)
    y += 18
    draw.text((pad_x + 1, y), "if the panel is still empty, the terminal needs its reporter.",
          font=ui(13), fill=MUTED)

    target = ASSETS / "doctor-green.png"
    base.save(target)
    return target


def main() -> None:
    if len(sys.argv) < 2:
        print("usage: make_doctor_card.py <doctor-output.txt> [<user-home-to-redact>]", file=sys.stderr)
        sys.exit(2)
    home = sys.argv[2] if len(sys.argv) > 2 else ""
    target = build(Path(sys.argv[1]), home)
    print("wrote: %s (%d bytes)" % (target.name, target.stat().st_size))
    sys.exit(0)


if __name__ == "__main__":
    main()
