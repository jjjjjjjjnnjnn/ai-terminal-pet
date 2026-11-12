"""Build the two launch assets that the README screenshots cannot cover.

`render_assets.py` produces still panels from the real Win32 rasterizer. The two
things it cannot produce, and which every launch post needs, are:

  demo.gif            an animation proving the panel actually reacts - and the
                      shape of the interaction (breathe, expand, click a row)
  desktop-scene.png   the panel in context, because every existing asset is a
                      bare panel with no terminal around it

Both are built from the same rasters: PET_PREVIEW_DIR must already contain the
TestRenderPreview / TestRenderSettingsPreview output. Nothing here invents a
pixel of the panel; the terminal windows in the scene are drawn abstractions and
the file says so in its own caption, so nothing claims to be a screen recording
of somebody else's machine.

Usage, from the repository root (PowerShell):

    $env:PET_PREVIEW_DIR = "$env:TEMP\\petpreview"
    go test ./internal/ui -run TestRenderPreview
    go test ./internal/ui -run TestRenderSettingsPreview
    python docs/assets/make_launch_assets.py
"""
from __future__ import annotations

import math
import os
import sys
from pathlib import Path

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8")
    sys.stderr.reconfigure(encoding="utf-8")

try:
    from PIL import Image, ImageDraw, ImageFont
except ImportError:
    print(
        "缺少 Pillow 依赖，无法生成 GIF 与实景图。\n"
        "Missing Pillow: pip install pillow",
        file=sys.stderr,
    )
    sys.exit(2)

ASSETS = Path(__file__).resolve().parent
PREVIEW_ENV = "PET_PREVIEW_DIR"

# The overlay's own backdrop. Panels are rendered on this colour and composited
# back onto it, which is exactly what the layered window does at runtime: the
# rasterizer writes premultiplied pixels, the desktop shows through the alpha.
DESKTOP = (0x20, 0x22, 0x2B)
INK = (0xF2, 0xF4, 0xF8)
MUTED = (0x9A, 0xA2, 0xB2)
ACCENT = (0xFF, 0xC8, 0x6A)
TERMINAL_BG = (0x14, 0x16, 0x1D)
TERMINAL_BAR = (0x22, 0x25, 0x2E)


def fail(message: str) -> None:
    print(message, file=sys.stderr)
    sys.exit(2)


def preview_dir() -> Path:
    raw = os.environ.get(PREVIEW_ENV, "").strip()
    if not raw:
        fail(
            "尚未设置环境变量 PET_PREVIEW_DIR，没有面板可合成。\n"
            "PET_PREVIEW_DIR is not set, so there are no rendered panels.\n"
            '    PowerShell: $env:PET_PREVIEW_DIR = "$env:TEMP\\petpreview"\n'
            "    go test ./internal/ui -run TestRenderPreview\n"
            "    go test ./internal/ui -run TestRenderSettingsPreview"
        )
    path = Path(raw).expanduser()
    if not path.is_dir():
        fail("PET_PREVIEW_DIR 指向的目录不存在: %s" % raw)
    return path


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    name = "segoeuib.ttf" if bold else "segoeui.ttf"
    for candidate in (Path("C:/Windows/Fonts") / name, Path("C:/Windows/Fonts/arial%s.ttf" % ("bd" if bold else ""))):
        try:
            return ImageFont.truetype(str(candidate), size)
        except OSError:
            continue
    return ImageFont.load_default()


def mono(size: int) -> ImageFont.FreeTypeFont:
    for candidate in ("consola.ttf", "cour.ttf"):
        try:
            return ImageFont.truetype(str(Path("C:/Windows/Fonts") / candidate), size)
        except OSError:
            continue
    return font(size)


def load(preview: Path, name: str) -> Image.Image:
    path = preview / name
    if not path.is_file():
        fail("PET_PREVIEW_DIR 里缺少 %s，先跑 TestRenderPreview 与 TestRenderSettingsPreview。" % name)
    return Image.open(path).convert("RGBA")


# trim() is deliberately not used for the two launch assets.
#
# The preview PNG is exactly the panel's canvas - 168x468 for a 56x156 strip at
# scale 3 - so the rounded frame already touches all four edges. Trimming to the
# non-backdrop bounding box crops the antialiased corner pixels away and leaves
# a panel whose corners look sliced off, which is what the first version of
# desktop-scene.png looked like. The backdrop is only visible at the corners,
# and it is the overlay's own composite colour, so the honest fix is to place
# the whole frame on that same colour.
def trim(img: Image.Image, bg=DESKTOP, pad: int = 10) -> Image.Image:
    """Crop the flat backdrop so the panel can be pasted flush anywhere."""
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
    return img.crop((max(0, box[0] - pad), max(0, box[1] - pad), min(w, box[2] + pad), min(h, box[3] + pad)))


def at_height(img: Image.Image, height: int) -> Image.Image:
    scale = height / img.height
    return img.resize((max(1, int(img.width * scale)), height), Image.LANCZOS)


def corner_radius(panel: Image.Image, logical_width: int) -> int:
    """The real corner radius of the rendered panel, in the panel's pixels.

    The rasterizer draws `min(width/2, 24)` logical units
    (drawing_compact_windows.go:167), so the radius is 24 units for both the
    56-wide compact strip and the 200-wide expanded list - a constant that
    happens to look like a percentage of the width but is not one. Scaling it
    from the logical width is the only way to get the corners right after the
    panel has been resized for a GIF frame.
    """
    return max(2, int(round(24 * (panel.width / float(logical_width)))))


def breathe(panel: Image.Image, opacity: float) -> Image.Image:
    """Apply the whole-window opacity the pulsing state actually uses.

    At runtime a waiting_permission panel does not change what it draws - the
    layered window's alpha for the entire frame drifts between roughly 0.72 and
    1.0 of the configured opacity. Animating anything else here would be a
    prettier lie about what the product does.
    """
    out = panel.copy()
    alpha = out.getchannel("A").point(lambda a: int(a * opacity))
    out.putalpha(alpha)
    return out


def demo_gif(preview: Path) -> Path:
    compact = load(preview, "compact-waiting-en.png")
    hovered = load(preview, "compact-waiting-hover.png")
    expanded = load(preview, "expanded-waiting-en.png")

    # One logical pixel is 3 preview pixels (scale 3) for compact and 2 for
    # expanded, so normalising by panel height keeps the two sizes in the same
    # proportion the product uses on a 150% display. The compact strip gets
    # more headroom than it needs at 470 because the breathing state is drawn
    # with a glow behind the orb, and a frame that clips it looks like a bug.
    compact = at_height(compact, 440)
    hovered = at_height(hovered, 440)
    expanded = at_height(expanded, 452)

    width, height = 760, 660
    caption_h = 108
    panel_area = height - caption_h

    captions = [
        ("A session is waiting. The strip breathes.", MUTED),
        ("Hover: it unfolds into the session list.", MUTED),
        ("Which project, which state, how long.", INK),
        ("Click the arrow at the row end: that terminal comes forward.", ACCENT),
    ]

    def frame(panel: Image.Image, caption: str, colour, x: int, opacity: float = 1.0) -> Image.Image:
        # The backdrop is the overlay's own composite colour, flat, and for one
        # reason only: the preview raster is flattened onto that colour, so
        # anything else here draws attention to a seam that comes from how the
        # preview is produced rather than from anything the product does. The
        # panel is pasted through a rounded mask instead of its own alpha, so
        # the corners stay corners.
        base = Image.new("RGB", (width, height), DESKTOP)
        draw = ImageDraw.Draw(base)
        radius = corner_radius(panel, 56 if panel.height > panel.width else 200)
        mask = Image.new("L", panel.size, 0)
        ImageDraw.Draw(mask).rounded_rectangle(
            [0, 0, panel.width - 1, panel.height - 1], radius=radius, fill=255
        )
        if opacity < 1.0:
            mask = mask.point(lambda a: int(a * opacity))
        base.paste(panel.convert("RGB"), (x, (panel_area - panel.height) // 2), mask)

        draw = ImageDraw.Draw(base)
        draw.line([(0, panel_area), (width, panel_area)], fill=(0x2A, 0x2D, 0x36))
        draw.rectangle([0, panel_area + 1, width, height], fill=TERMINAL_BG)
        draw.ellipse([56, panel_area + 40, 68, panel_area + 52], fill=colour)
        draw.text((82, panel_area + 34), caption, font=font(26, True), fill=colour)
        return base

    frames: list[Image.Image] = []
    durations: list[int] = []

    # 1. Breathe. Twelve frames of 90 ms is 1.08 s per cycle, which is what
    #    pulseInterval (160 ms) produces at the shipped 6.25 fps breath.
    compact_x = (width - compact.width) // 2 + 90
    for step in range(12):
        phase = step / 12
        opacity = 0.72 + 0.28 * (0.5 - 0.5 * math.cos(2 * math.pi * phase))
        frames.append(frame(compact, captions[0][0], captions[0][1], compact_x, opacity))
        durations.append(90)

    # 2-3. Hover, then unfold, each held long enough to read. These are single
    #    frames with a long duration rather than many identical frames: a GIF
    #    encoder collapses identical consecutive frames and keeps only one of
    #    their durations, so ten copies of "unfolded" would flash past in 60 ms
    #    instead of holding for 600. Saying the hold in the duration is the only
    #    spelling that survives encoding.
    frames.append(frame(hovered, captions[1][0], captions[1][1], compact_x))
    durations.append(700)
    expanded_x = 40
    frames.append(frame(expanded, captions[2][0], captions[2][1], expanded_x))
    durations.append(1500)

    # 4. The click. A ring grows out of the waiting row's focus arrow, because
    #    a layered window has no cursor to show and the product answers with the
    #    jump itself - the ring is the closest honest stand-in in a still image.
    for step in range(10):
        base = frame(expanded, captions[3][0], captions[3][1], expanded_x)
        draw = ImageDraw.Draw(base)
        # Centred on the first row's jump arrow, measured off the rendered
        # frame rather than guessed: 0.88 of the panel width across, 0.43 of
        # its height down. Getting this wrong puts a "click" ring on the
        # close button, which is the one place a click means something else.
        cx = expanded_x + int(expanded.width * 0.88)
        cy = (panel_area - expanded.height) // 2 + int(expanded.height * 0.43)
        radius = 6 + step * 4
        shade = int(0xC8 - step * 8)
        draw.ellipse([cx - radius, cy - radius, cx + radius, cy + radius],
                     outline=(0xFF, max(0x60, shade), 0x6A))
        frames.append(base)
        durations.append(70)

    # Hold the last beat so the loop does not snap back mid-read.
    frames.append(frame(expanded, captions[3][0], captions[3][1], expanded_x))
    durations.append(1400)

    # Palette conversion per frame is what keeps the GIF small; PIL needs a
    # single palette for a global-table save, so adapt one frame and hand the
    # rest the same table.
    quantised = [f.quantize(colors=128, method=Image.MEDIANCUT, dither=Image.NONE) for f in frames]
    target = ASSETS / "demo.gif"
    quantised[0].save(
        target,
        save_all=True,
        append_images=quantised[1:],
        duration=durations,
        loop=0,
        optimize=True,
        disposal=1,
    )
    return target


def desktop_scene(preview: Path) -> Path:
    """Two terminals and the panel on the same screen.

    The terminals are drawn, the panel is the real rasterizer output. Drawing
    them is deliberate: a real capture of somebody else's desktop would include
    their project names, and a real capture of this machine has no two terminals
    open. The alt text in README says which half is an illustration.
    """
    width, height = 1500, 860
    base = Image.new("RGB", (width, height), (0x0E, 0x10, 0x16))
    draw = ImageDraw.Draw(base)
    for y in range(height):
        shade = int(10 * (y / height))
        draw.line([(0, y), (width, y)], fill=(0x0E + shade, 0x10 + shade, 0x16 + shade))
    compact = load(preview, "compact-waiting-en.png")
    compact = at_height(compact, 420)

    def terminal(x: int, y: int, w: int, h: int, title: str, lines: list[tuple[str, tuple[int, int, int]]], focused: bool) -> None:
        draw.rounded_rectangle([x, y, x + w, y + h], radius=10,
                               fill=TERMINAL_BG, outline=(0x36 if focused else 0x26, 0x39, 0x44))
        draw.rounded_rectangle([x, y, x + w, y + 40], radius=10, fill=TERMINAL_BAR)
        draw.rectangle([x, y + 28, x + w, y + 40], fill=TERMINAL_BAR)
        for index, dot in enumerate(((0x5A, 0x5F, 0x6B), (0x5A, 0x5F, 0x6B), (0x5A, 0x5F, 0x6B))):
            draw.ellipse([x + 16 + index * 16, y + 15, x + 24 + index * 16, y + 23], fill=dot)
        draw.text((x + 66, y + 11), title, font=mono(17), fill=(0xB4, 0xBC, 0xCA))
        body = mono(19)
        line_y = y + 62
        for text, colour in lines:
            draw.text((x + 22, line_y), text, font=body, fill=colour)
            line_y += 27

    dim = (0x6B, 0x73, 0x82)
    normal = (0xC8, 0xD0, 0xDC)
    green = (0x8F, 0xD6, 0x9A)
    amber = (0xFF, 0xC8, 0x6A)

    terminal(64, 72, 940, 372, "claude code - ai-terminal-pet", [
        ("> add a retry to the upload path", normal),
        ("  Bash(go test ./internal/store)          ", normal),
        ("  waiting for approval                  ", amber),
        ("  [ ] allow   [ ] deny                  ", dim),
    ], focused=False)

    terminal(64, 484, 940, 300, "opencode - webapp", [
        ("editing src/app.tsx                      ", normal),
        ("  tests 24 passed                        ", green),
        ("  watching for changes...                ", dim),
    ], focused=True)

    # The panel sits at the right edge, at the height the product would put it.
    # Pasting through a rounded mask, not the panel's own alpha: the preview PNG
    # is flattened onto the backdrop, so its alpha is opaque everywhere. The
    # mask reproduces what the layered window really does - everything outside
    # the rounded frame is fully transparent, and the desktop shows through -
    # and its radius is the rasterizer's own min(width/2, 24) logical units.
    radius = corner_radius(compact, 56)
    mask = Image.new("L", compact.size, 0)
    ImageDraw.Draw(mask).rounded_rectangle(
        [0, 0, compact.width - 1, compact.height - 1], radius=radius, fill=255
    )
    base.paste(compact, (width - compact.width - 56, 190), mask)

    draw.text((64, 812), "panel: rendered by the real Win32 rasterizer. terminals: illustration.",
              font=font(16), fill=(0x59, 0x60, 0x6E))

    target = ASSETS / "desktop-scene.png"
    base.save(target)
    return target


def main() -> None:
    preview = preview_dir()
    ASSETS.mkdir(parents=True, exist_ok=True)
    gif = demo_gif(preview)
    scene = desktop_scene(preview)
    for path in (gif, scene):
        size = path.stat().st_size
        print("wrote: %s (%d bytes)" % (path.name, size))
    sys.exit(0)


if __name__ == "__main__":
    main()
