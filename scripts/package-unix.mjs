#!/usr/bin/env node
// package-unix.mjs - build and package portable tar.gz archives for Linux and macOS.
//
// Zero npm dependencies. Cross-compiles Go binaries with CGO_ENABLED=0
// (-trimpath -buildvcs=false -ldflags "-s -w") and emits deterministic POSIX
// USTAR tar.gz archives with 0755 permissions preserved on executables even
// when packaged from Windows.
//
// Usage:
//   node scripts/package-unix.mjs
//   node scripts/package-unix.mjs --target=linux/amd64,darwin/arm64
//   node scripts/package-unix.mjs --output=dist --version=1.0.13

import fs from "node:fs";
import path from "node:path";
import crypto from "node:crypto";
import zlib from "node:zlib";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..");

const DEFAULT_TARGETS = [
  "linux/amd64",
  "linux/arm64",
  "darwin/arm64",
  "darwin/amd64",
];

function parseArgs(argv) {
  let outputDir = path.join(repoRoot, "dist");
  let version = "";
  let targets = DEFAULT_TARGETS;
  for (const arg of argv) {
    if (arg.startsWith("--output=")) {
      outputDir = path.resolve(arg.slice("--output=".length));
    } else if (arg.startsWith("--version=")) {
      version = arg.slice("--version=".length).trim();
    } else if (arg.startsWith("--target=")) {
      targets = arg
        .slice("--target=".length)
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean);
    } else if (arg === "--help" || arg === "-h") {
      console.log(
        "Usage: node scripts/package-unix.mjs [--target=os/arch,...] [--output=dir] [--version=X.Y.Z]"
      );
      process.exit(0);
    } else {
      console.error(`package-unix: unknown argument ${arg}`);
      process.exit(2);
    }
  }
  return { outputDir, version, targets };
}

function readOverlayVersion() {
  const versionFile = path.join(repoRoot, "internal", "app", "version.go");
  const raw = fs.readFileSync(versionFile, "utf8");
  const match = raw.match(/const\s+Version\s*=\s*"([^"]+)"/);
  if (!match) {
    throw new Error(`cannot find Version in ${versionFile}`);
  }
  return match[1];
}

function goBuild(goos, goarch, pkg, outFile) {
  const res = spawnSync(
    "go",
    ["build", "-trimpath", "-buildvcs=false", "-ldflags", "-s -w", "-o", outFile, pkg],
    {
      cwd: repoRoot,
      env: {
        ...process.env,
        CGO_ENABLED: "0",
        GOOS: goos,
        GOARCH: goarch,
        GO111MODULE: "on",
      },
      stdio: "inherit",
    }
  );
  if (res.status !== 0) {
    throw new Error(`go build failed for ${goos}/${goarch} ${pkg} (exit ${res.status})`);
  }
}

function encodeOctal(value, width) {
  const str = value.toString(8).padStart(width - 1, "0") + "\0";
  return Buffer.from(str.slice(0, width), "ascii");
}

function makeTarHeader(name, size, mode, mtimeSec, typeflag = "0") {
  const buf = Buffer.alloc(512, 0);
  let prefix = "";
  let fileName = name;
  if (Buffer.byteLength(name) > 100) {
    const slash = name.lastIndexOf("/", 154);
    if (slash > 0 && Buffer.byteLength(name.slice(slash + 1)) <= 100) {
      prefix = name.slice(0, slash);
      fileName = name.slice(slash + 1);
    } else {
      throw new Error(`tar entry path too long: ${name}`);
    }
  }
  buf.write(fileName, 0, 100, "utf8");
  encodeOctal(mode, 8).copy(buf, 100);
  encodeOctal(0, 8).copy(buf, 108); // uid
  encodeOctal(0, 8).copy(buf, 116); // gid
  encodeOctal(size, 12).copy(buf, 124);
  encodeOctal(mtimeSec, 12).copy(buf, 136);
  buf.fill(0x20, 148, 156); // checksum field spaces during computation
  buf.write(typeflag, 156, 1, "ascii");
  buf.write("ustar\0", 257, 6, "ascii");
  buf.write("00", 263, 2, "ascii");
  if (prefix) {
    buf.write(prefix, 345, 155, "utf8");
  }
  let sum = 0;
  for (let i = 0; i < 512; i++) {
    sum += buf[i];
  }
  buf.write(sum.toString(8).padStart(6, "0") + "\0 ", 148, 8, "ascii");
  return buf;
}

function buildTarGz(entries, outFile) {
  const mtimeSec = Math.floor(Date.now() / 1000);
  const chunks = [];
  for (const entry of entries) {
    const header = makeTarHeader(entry.archivePath, entry.data.length, entry.mode, mtimeSec);
    chunks.push(header);
    chunks.push(entry.data);
    const rem = entry.data.length % 512;
    if (rem !== 0) {
      chunks.push(Buffer.alloc(512 - rem, 0));
    }
  }
  chunks.push(Buffer.alloc(1024, 0)); // end-of-archive zero blocks
  const tarBuf = Buffer.concat(chunks);
  const gzBuf = zlib.gzipSync(tarBuf, { level: 9 });
  fs.writeFileSync(outFile, gzBuf);
  return {
    bytes: gzBuf.length,
    sha256: crypto.createHash("sha256").update(gzBuf).digest("hex"),
  };
}

function installReadme(goos, goarch, version) {
  return (
    `ai-terminal-pet ${version} (${goos}/${goarch})\n` +
    `=========================================\n\n` +
    `1. Run ./install.sh to install binaries and protocol.js into your per-user\n` +
    `   data directory (~/.local/share/ai-terminal-pet on Linux, or\n` +
    `   ~/Library/Application Support/AI-Terminal-Pet on macOS) and symlink\n` +
    `   ~/.local/bin/ai-terminal-pet.\n\n` +
    `2. Verify the install:\n` +
    `   ai-terminal-pet --doctor\n\n` +
    `3. Print terminal wiring snippets (Claude Code, OpenCode, Antigravity):\n` +
    `   ai-terminal-pet --connect\n\n` +
    `4. Uninstall (preserves config.json):\n` +
    `   ./install.sh --uninstall\n`
  );
}

function packageTarget(version, outputDir, target) {
  const [goos, goarch] = target.split("/");
  if (!goos || !goarch) {
    throw new Error(`invalid target ${target} (expected os/arch)`);
  }
  const pkgName = `ai-terminal-pet-${version}-${goos}-${goarch}`;
  const stageDir = fs.mkdtempSync(path.join(outputDir, `.stage-${goos}-${goarch}-`));
  try {
    const binaries = [
      ["./cmd/ai-terminal-pet", "bin/AI-Terminal-Pet"],
      ["./cmd/pet-statusline", "bin/pet-statusline"],
      ["./cmd/pet-hook", "bin/pet-hook"],
    ];
    const items = [];
    for (const [pkg, relPath] of binaries) {
      const builtPath = path.join(stageDir, path.basename(relPath));
      goBuild(goos, goarch, pkg, builtPath);
      items.push({ relPath, data: fs.readFileSync(builtPath), mode: 0o755 });
    }

    const staticTextFiles = [
      ["reporters/protocol.js", "reporters/protocol.js", 0o644],
      ["reporters/opencode.js", "reporters/opencode.js", 0o644],
      ["reporters/claude-hook.js", "reporters/claude-hook.js", 0o644],
      ["reporters/antigravity-hook.js", "reporters/antigravity-hook.js", 0o644],
      ["reporters/antigravity-statusline.js", "reporters/antigravity-statusline.js", 0o644],
      ["reporters/generic.js", "reporters/generic.js", 0o644],
      ["scripts/install.sh", "scripts/install.sh", 0o755],
      ["LICENSE", "LICENSE", 0o644],
    ];
    for (const [srcRel, dstRel, mode] of staticTextFiles) {
      const data = fs.readFileSync(path.join(repoRoot, srcRel));
      items.push({ relPath: dstRel, data, mode });
    }

    items.push({
      relPath: "install.sh",
      data: Buffer.from('#!/usr/bin/env sh\nexec "$(dirname "$0")/scripts/install.sh" "$@"\n', "utf8"),
      mode: 0o755,
    });
    items.push({
      relPath: "README.install.txt",
      data: Buffer.from(installReadme(goos, goarch, version), "utf8"),
      mode: 0o644,
    });

    items.sort((a, b) => a.relPath.localeCompare(b.relPath));
    const sumLines = items.map(
      (item) => `${crypto.createHash("sha256").update(item.data).digest("hex")}  ${item.relPath}`
    );
    items.push({
      relPath: "SHA256SUMS.txt",
      data: Buffer.from(sumLines.join("\n") + "\n", "utf8"),
      mode: 0o644,
    });

    const tarEntries = items.map((item) => ({
      archivePath: `${pkgName}/${item.relPath}`,
      data: item.data,
      mode: item.mode,
    }));
    const archivePath = path.join(outputDir, `${pkgName}.tar.gz`);
    const info = buildTarGz(tarEntries, archivePath);
    console.log(`[ok] ${path.basename(archivePath)} (${info.bytes} bytes, sha256=${info.sha256})`);
    return { archivePath, ...info };
  } finally {
    fs.rmSync(stageDir, { recursive: true, force: true });
  }
}

function main() {
  const opts = parseArgs(process.argv.slice(2));
  const version = opts.version || readOverlayVersion();
  fs.mkdirSync(opts.outputDir, { recursive: true });
  console.log(`package-unix: overlay version ${version}`);
  for (const target of opts.targets) {
    packageTarget(version, opts.outputDir, target);
  }
}

main();
