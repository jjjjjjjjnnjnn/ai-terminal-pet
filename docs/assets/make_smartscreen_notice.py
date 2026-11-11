"""Render the SmartScreen warning an unsigned build produces, as a labelled mockup.

Why a mockup and not a screenshot: the dialog only appears on a machine that did
not build this project (Defender flags files that arrived with a zone marker),
and a genuine capture carries somebody's desktop, machine name and Windows
build. What the post actually needs to convey is "you will see this, click
here" - so the dialog is rebuilt from the documented text, and the image says
on its face that it is a reconstruction. Presenting it as a capture would be a
lie about provenance for no benefit; the dialog itself is a screenshot of the
same OS whether it is 26100 or 22631.

Usage, from the repository root:

    python docs/assets/make_smartscreen_notice.py
"""
from __future__ import annotations

import sys
from pathlib import Path

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")

try:
    from PIL import Image, ImageDraw, ImageFont
except ImportError:
    print("缺少 Pillow 依赖，无法生成图片。\nMissing Pillow: pip install pillow", file=sys.stderr)
    sys.exit(2)

ASSETS = Path(__file__).resolve().parent

# Windows 11 light-theme dialog colours, sampled from the shipped dialog rather
# than guessed: #FFFFFF card on a #F3F3F3 scrim, #1A1A1A title text, #5D5D5D body.
CARD = (0xFF, 0xFF, 0xFF)
SCRIM = (0xF3, 0xF3, 0xF3)
INK = (0x1A, 0x1A, 0x1A)
BODY = (0x5D, 0x5D, 0x5D)
MUTED = (0x77, 0x77, 0x77)
ACCENT = (0x00, 0x67, 0xC0)
WARN = (0xC4, 0x2B, 0x1C)
BUTTON = (0xF5, 0xF5, 0xF5)
BUTTON_EDGE = (0xD6, 0xD6, 0xD6)


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    for name in (("segoeuib.ttf" if bold else "segoeui.ttf"), "arial%s.ttf" % ("bd" if bold else "")):
        try:
            return ImageFont.truetype(str(Path("C:/Windows/Fonts") / name), size)
        except OSError:
            continue
    return ImageFont.load_default()


def shield(draw: ImageDraw.ImageDraw, x: int, y: int, size: int, colour) -> None:
    """The blue-and-red shield Defender uses, drawn as a filled polygon.

    A rasterized logo would mean shipping Microsoft artwork inside this
    repository; a shield shape in the right colours reads the same at the size
    this appears in a post and stays ours to license.
    """
    h = size
    w = int(size * 0.84)
    top = y
    mid = y + int(h * 0.52)
    bottom = y + h
    draw.polygon(
        [(x + w // 2, top), (x + w, top + int(h * 0.12)), (x + w, mid),
         (x + w // 2, bottom), (x, mid), (x, top + int(h * 0.12))],
        fill=(0x00, 0x5F, 0xB8),
    )
    inner_w = int(w * 0.62)
    inner_x = x + (w - inner_w) // 2
    draw.polygon(
        [(x + w // 2, top + int(h * 0.16)), (x + w // 2 + inner_w // 2, mid - int(h * 0.08)),
         (x + w // 2, mid + int(h * 0.14)), (x + w // 2 - inner_w // 2, mid - int(h * 0.08))],
        fill=(0xFF, 0xFF, 0xFF),
    )
    draw.line([(x + w // 2, mid + int(h * 0.02)), (x + w // 2, mid + int(h * 0.34))],
              fill=WARN, width=max(2, size // 22))
    draw.ellipse([x + w // 2 - size // 20, mid + int(h * 0.40),
                  x + w // 2 + size // 20, mid + int(h * 0.40) + size // 10], fill=WARN)


def dialog(draw: ImageDraw.ImageDraw, origin: tuple[int, int], expanded_button: bool) -> None:
    """One SmartScreen card, in the two states a user actually walks through."""
    x0, y0 = origin
    w, h = 560, 348
    draw.rounded_rectangle([x0, y0, x0 + w, y0 + h], radius=8, fill=CARD, outline=BUTTON_EDGE)

    shield(draw, x0 + 32, y0 + 32, 34, (0x00, 0x5F, 0xB8))

    draw.text((x0 + 84, y0 + 26), "Windows protected your PC",
              font=font(22, True), fill=INK)
    body = font(13)
    lines = [
        "Microsoft Defender SmartScreen prevented an unrecognized",
        "app from starting. Running this app might put your PC at",
        "risk.",
    ]
    for index, line in enumerate(lines):
        draw.text((x0 + 84, y0 + 60 + index * 19), line, font=body, fill=BODY)

    draw.text((x0 + 84, y0 + 132), "App:", font=body, fill=MUTED)
    draw.text((x0 + 124, y0 + 132), "install.cmd", font=body, fill=BODY)
    draw.text((x0 + 84, y0 + 152), "Publisher:", font=body, fill=MUTED)
    draw.text((x0 + 148, y0 + 152), "Unknown publisher", font=body, fill=BODY)

    if expanded_button:
        draw.text((x0 + 84, y0 + 182), "More info", font=body, fill=ACCENT)
        draw.line([(x0 + 84, y0 + 199), (x0 + 143, y0 + 199)], fill=ACCENT, width=1)

    # The buttons sit bottom-right, More info left of Don't run, exactly as the
    # shipped dialog lays them out. A hairline above that row is what makes the
    # card read as a dialog rather than as a picture of buttons.
    by = y0 + h - 58
    draw.line([(x0 + 1, y0 + h - 70), (x0 + w - 1, y0 + h - 70)],
              fill=(0xE6, 0xE6, 0xE6), width=1)
    if expanded_button:
        draw.rounded_rectangle([x0 + 84, by + 26, x0 + 248, by + 58], radius=4,
                               fill=BUTTON, outline=BUTTON_EDGE)
        draw.text((x0 + 100, by + 34), "Run anyway", font=body, fill=BODY)
    draw.rounded_rectangle([x0 + 352, by + 26, x0 + 458, by + 58], radius=4,
                           fill=(0x00, 0x5F, 0xB8), outline=(0x00, 0x5F, 0xB8))
    draw.text((x0 + 376, by + 34), "Don't run", font=body, fill=(0xFF, 0xFF, 0xFF))
    draw.rounded_rectangle([x0 + 468, by + 26, x0 + 548, by + 58], radius=4,
                           fill=BUTTON, outline=BUTTON_EDGE)
    draw.text((x0 + 480, by + 34), "More info", font=body, fill=BODY)


def build() -> Path:
    W, H = 1320, 560
    base = Image.new("RGB", (W, H), SCRIM)
    draw = ImageDraw.Draw(base)

    # Two states side by side, because the whole point is the second click.
    dialog(draw, (40, 60), expanded_button=False)
    dialog(draw, (700, 60), expanded_button=True)

    arrow_y = 250
    draw.line([(610, arrow_y), (684, arrow_y)], fill=(0x8A, 0x8A, 0x8A), width=3)
    draw.polygon([(684, arrow_y), (672, arrow_y - 7), (672, arrow_y + 7)], fill=(0x8A, 0x8A, 0x8A))
    draw.text((604, arrow_y + 12), "click", font=font(12), fill=MUTED)
    draw.text((596, arrow_y + 28), "More info", font=font(12, True), fill=MUTED)

    # Captions under each card.
    draw.text((48, 470), "1. Windows has never seen this publisher.",
              font=font(15, True), fill=INK)
    draw.text((708, 470), "2. More info reveals the override, then Run anyway.",
              font=font(15, True), fill=INK)

    # The honest line. It is on the image itself, not only in the README, so the
    # picture cannot be lifted out of context and read as a real capture.
    draw.text((48, 512),
              "MOCKUP: dialog rebuilt from the documented SmartScreen text, not a screenshot of a real machine.",
              font=font(12), fill=(0x99, 0x99, 0x99))

    target = ASSETS / "smartscreen-unsigned.png"
    base.save(target)
    return target


def main() -> None:
    target = build()
    print("wrote: %s (%d bytes)" % (target.name, target.stat().st_size))
    sys.exit(0)


if __name__ == "__main__":
    main()
