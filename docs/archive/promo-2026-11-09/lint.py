#!/usr/bin/env python3
# Fact-check linter for drafts. Read-only: prints findings, changes nothing.
import re, os, sys, glob

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

BANNED = ["revolutionary","blazingly","blazing","supercharge","game-changer","game changer",
          "10x","seamless","effortless","cutting-edge","next-generation","next generation",
          "unlock","powerful","delightful","magical","simply","just works","robust",
          "comprehensive","elegant",
          "革命性","颠覆","极致","丝滑","无与伦比","强大","优雅","赋能","闭眼入","神器",
          "黑科技","遥遥领先","一键搞定","体验拉满","效率翻倍"]

NUMBERS_OK = {
    "0.9.0","1.23","140","3","256","1 MiB","1MiB","32","5 s","5s","5 秒","15 s","30 s",
    "45","2","65","160","56x156","56×156","200","6","4,178,432","3,644,416","10,201",
    "3,831","1,373","1,742","44.1","16-bit","85","100","60","1,060","10 min","10 分钟",
    "14","11","0.19","0.562","0.5%","22","38","8","16","4.05","4,067,840","3,977,216",
    "1742","10201","3831","1373","4178432","3644416","7","3.98","3.48","1–6","1-6",
    "120","96%","150","504","306","450","2–3","2-3","2026","11","24","4 min","+4",
    # verified extras (fact sheet / docs/ROADMAP.md L2 / README structure)
    "1","4","5","9.0","10","30","40","1:1","3.7","2.5","5 MiB","0.9","v0.9.0",
    "235","213","217","221","196","11 s","11 分钟","十四分钟","56","156","200 px",
}

def check(path):
    text = open(path, encoding="utf-8").read()
    lines = text.splitlines()
    out = []
    name = os.path.basename(path)
    # 1. images exist
    for m in re.finditer(r"!\[[^\]]*\]\(([^)]+)\)", text):
        p = m.group(1).split(" ")[0].strip('"\'')
        if p.startswith("http"): continue
        if not os.path.exists(os.path.join(ROOT, p)):
            out.append("MISSING IMAGE: %s" % p)
    # 2. banned words
    low = text.lower()
    for w in BANNED:
        for m in re.finditer(re.escape(w), low):
            ln = low.count("\n", 0, m.start()) + 1
            out.append("BANNED '%s' at %s:%d" % (w, name, ln))
    # 3. exclamation marks outside code fences
    in_code = False
    for i, line in enumerate(lines, 1):
        if line.startswith("```"):
            in_code = not in_code
            continue
        if in_code: continue
        if "!" in line and not line.startswith("!["):
            out.append("EXCLAMATION at %s:%d: %s" % (name, i, line.strip()[:90]))
    # 4. emoji count
    emoji = re.findall("[\U0001F300-\U0001FAFF\u2600-\u27BF]", text)
    out.append("EMOJI count: %d" % len(emoji))
    # 5. link targets
    for m in re.finditer(r"\[([^\]]+)\]\(([^)]+)\)", text):
        t = m.group(2).split(" ")[0]
        if t.startswith("http") or t.startswith("#") or t.startswith("mailto:"): continue
        if not os.path.exists(os.path.join(ROOT, t.split("#")[0])):
            out.append("BROKEN LINK '%s' -> %s" % (m.group(1), t))
    # 6. numeric claims
    for m in re.finditer(r"\b\d[\d,.]*\b", text):
        tok = m.group(0)
        key = tok.strip(".,")
        if key in NUMBERS_OK: continue
        ln = text.count("\n", 0, m.start()) + 1
        ctx = lines[ln-1].strip()[:100]
        out.append("NUM? %s:%d '%s' | %s" % (name, ln, key, ctx))
    out.append("LINES: %d  WORDS: %d  CHARS: %d" % (len(lines), len(re.findall(r"\S+", text)), len(text)))
    return out

for f in sys.argv[1:] or sorted(glob.glob(os.path.join(ROOT, "drafts", "README*.md")) + glob.glob(os.path.join(ROOT, "drafts", "SOCIAL.md"))):
    print("=== %s ===" % os.path.basename(f))
    for line in check(f):
        print(" ", line)
