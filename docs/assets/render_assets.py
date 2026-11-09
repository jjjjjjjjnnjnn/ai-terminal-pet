"""Build README hero + screenshot assets from the real Win32 rasterizer output.

Screenshots come from `go test ./internal/ui -run TestRenderPreview` with
PET_PREVIEW_DIR set, so they can never drift from the shipped layout.
"""
from __future__ import annotations

import sys
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

PREVIEW = Path(r"C:\Users\rongj\AppData\Local\Temp\petpreview2")
ASSETS = Path(r"C:\Users\rongj\Desktop\项目开发\ai-terminal-pet\docs\assets")
PANEL_BG = (0x20, 0x22, 0x2b)


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    path = "C:/Windows/Fonts/segoeuib.ttf" if bold else "C:/Windows/Fonts/segoeui.ttf"
    try:
        return ImageFont.truetype(path, size)
    except OSError:
        return ImageFont.truetype("C:/Windows/Fonts/arial%s.ttf" % ("bd" if bold else ""), size)


def trim(img: Image.Image, bg: tuple[int, int, int], pad: int = 16) -> Image.Image:
    """Crop the flat backdrop around a rendered panel so it sits flush."""
    rgb = img.convert("RGB")
    mask = Image.new("L", rgb.size, 0)
    data = mask.load()
    src = rgb.load()
    w, h = rgb.size
    for y in range(h):
        for x in range(w):
            if src[x, y] != bg:
                data[x, y] = 255
    box = mask.getbbox()
    if not box:
        return img
    return img.crop((max(0, box[0] - pad), max(0, box[1] - pad),
                     min(w, box[2] + pad), min(h, box[3] + pad)))


def flat_panels() -> dict[str, Image.Image]:
    return {
        "compact": trim(Image.open(PREVIEW / "compact-waiting.png"), PANEL_BG),
        "compact-hover": trim(Image.open(PREVIEW / "compact-waiting-hover.png"), PANEL_BG),
        "compact-en": trim(Image.open(PREVIEW / "compact-waiting-en.png"), PANEL_BG),
        "expanded": trim(Image.open(PREVIEW / "expanded-waiting.png"), PANEL_BG),
        "expanded-en": trim(Image.open(PREVIEW / "expanded-waiting-en.png"), PANEL_BG),
        "expanded-many": trim(Image.open(PREVIEW / "expanded-many.png"), PANEL_BG),
        "auto-approve": trim(Image.open(PREVIEW / "expanded-approve-zh.png"), PANEL_BG),
    }


def hero(panels: dict[str, Image.Image]) -> Image.Image:
    W, H = 1600, 640
    base = Image.new("RGB", (W, H), (0x14, 0x16, 0x1d))
    draw = ImageDraw.Draw(base)

    # Vertical gradient plus a warm glow behind the panels.
    for y in range(H):
        shade = int(5 * (y / H))
        draw.line([(0, y), (W, y)], fill=(0x14 + shade, 0x16 + shade, 0x1d + shade))
    glow = Image.new("RGB", (W, H), (0, 0, 0))
    ImageDraw.Draw(glow).ellipse([880, -320, 1900, 760], fill=(0x2b, 0x1e, 0x09))
    base = Image.blend(base, glow, 0.30)
    draw = ImageDraw.Draw(base)

    ink = (0xf2, 0xf4, 0xf8)
    accent = (0xff, 0xc8, 0x6a)
    muted = (0x9a, 0xa2, 0xb2)

    draw.ellipse([82, 96, 104, 118], fill=accent)
    draw.text((120, 88), "AI TERMINAL PET", font=font(28, True), fill=(0xd8, 0xdd, 0xe6))

    draw.text((80, 176), "One glance tells you", font=font(58, True), fill=ink)
    draw.text((80, 246), "which AI terminal", font=font(58, True), fill=accent)
    draw.text((80, 316), "is waiting on you.", font=font(58, True), fill=ink)
    draw.text((84, 412), "Windows overlay · named-pipe session state",
              font=font(24), fill=muted)
    draw.text((84, 448), "Go + Win32 · zero runtime dependencies",
              font=font(24), fill=muted)

    draw.text((84, H - 78), "OpenCode · Claude Code · Antigravity CLI · DeepSeek Harness",
              font=font(21), fill=(0x6f, 0x78, 0x88))

    def fit(img: Image.Image, max_h: int) -> Image.Image:
        scale = max_h / img.height
        return img.resize((int(img.width * scale), max_h), Image.LANCZOS)

    compact = fit(panels["compact-en"], 380)
    expanded = fit(panels["expanded-en"], 400)

    compact_x, expanded_x = 950, 1190
    bottom = H - 96
    base.paste(compact, (compact_x, bottom - compact.height))
    base.paste(expanded, (expanded_x, bottom - expanded.height))
    draw.text((compact_x + 4, bottom - compact.height - 26), "compact",
              font=font(19), fill=muted)
    draw.text((expanded_x + 4, bottom - expanded.height - 26), "expanded",
              font=font(19), fill=muted)

    # Hairline frame so the sheet reads as a banner, not a screenshot.
    draw.rectangle([20, 20, W - 21, H - 21], outline=(0x2a, 0x2d, 0x36))
    return base


def main() -> None:
    ASSETS.mkdir(parents=True, exist_ok=True)
    panels = flat_panels()
    for name, img in panels.items():
        img.save(ASSETS / f"overlay-{name}.png")
    hero(panels).save(ASSETS / "hero.png")
    print("wrote:", ", ".join(sorted(p.name for p in ASSETS.glob("*.png"))))
    sys.exit(0)


if __name__ == "__main__":
    main()
