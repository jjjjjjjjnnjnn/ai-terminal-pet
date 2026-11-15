#!/usr/bin/env sh
# install.sh - install, verify, or uninstall ai-terminal-pet on Linux and macOS.
#
# Usage:
#   ./scripts/install.sh              Install into the per-user data directory
#   ./scripts/install.sh --verify-only Check installed files against source/manifest
#   ./scripts/install.sh --uninstall   Remove installed binaries (keeps config.json)
#   ./scripts/install.sh --uninstall --purge  Also remove config.json/state.json/runtime
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)"

VERIFY_ONLY=0
UNINSTALL=0
PURGE=0
LINK_BIN=1

for arg in "$@"; do
  case "$arg" in
    --verify-only|-VerifyOnly)
      VERIFY_ONLY=1
      ;;
    --uninstall|-Uninstall)
      UNINSTALL=1
      ;;
    --purge)
      PURGE=1
      ;;
    --no-link)
      LINK_BIN=0
      ;;
    --help|-h)
      cat <<'EOF'
Usage: install.sh [--verify-only] [--uninstall [--purge]] [--no-link]
EOF
      exit 0
      ;;
    *)
      echo "install.sh: unknown option: $arg" >&2
      exit 2
      ;;
  esac
done

resolve_data_dir() {
  if [ -z "${HOME:-}" ]; then
    echo "install.sh: HOME is not set" >&2
    exit 1
  fi
  os_name="$(uname -s)"
  case "$os_name" in
    Darwin)
      printf '%s/Library/Application Support/AI-Terminal-Pet\n' "${HOME%/}"
      ;;
    *)
      case "${XDG_DATA_HOME:-}" in
        /*)
          printf '%s/ai-terminal-pet\n' "${XDG_DATA_HOME%/}"
          ;;
        *)
          printf '%s/.local/share/ai-terminal-pet\n' "${HOME%/}"
          ;;
      esac
      ;;
  esac
}

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print tolower($1)}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print tolower($1)}'
  else
    echo "install.sh: neither sha256sum nor shasum found" >&2
    exit 1
  fi
}

INSTALL_DIR="$(resolve_data_dir)"
BIN_LINK="${HOME%/}/.local/bin/ai-terminal-pet"

if [ "$UNINSTALL" -eq 1 ]; then
  if [ -x "$INSTALL_DIR/AI-Terminal-Pet" ]; then
    "$INSTALL_DIR/AI-Terminal-Pet" --autostart=off >/dev/null 2>&1 || true
    "$INSTALL_DIR/AI-Terminal-Pet" --quit >/dev/null 2>&1 || true
  fi
  for f in AI-Terminal-Pet pet-statusline pet-hook protocol.js; do
    rm -f "$INSTALL_DIR/$f"
  done
  if [ -L "$BIN_LINK" ]; then
    rm -f "$BIN_LINK"
  fi
  if [ "$PURGE" -eq 1 ]; then
    rm -rf "$INSTALL_DIR"
  fi
  echo "[ok] uninstalled from $INSTALL_DIR"
  exit 0
fi

FILES="bin/AI-Terminal-Pet:AI-Terminal-Pet bin/pet-statusline:pet-statusline bin/pet-hook:pet-hook reporters/protocol.js:protocol.js"

for pair in $FILES; do
  src="$REPO_ROOT/${pair%%:*}"
  if [ ! -f "$src" ]; then
    echo "[fail] missing source artifact: $src" >&2
    exit 1
  fi
done

if [ "$VERIFY_ONLY" -eq 1 ]; then
  failed=0
  for pair in $FILES; do
    src="$REPO_ROOT/${pair%%:*}"
    dst="$INSTALL_DIR/${pair#*:}"
    if [ ! -f "$dst" ]; then
      echo "[fail] missing installed file: $dst"
      failed=1
      continue
    fi
    h1="$(sha256_file "$src")"
    h2="$(sha256_file "$dst")"
    if [ "$h1" != "$h2" ]; then
      echo "[fail] hash mismatch for ${pair#*:}"
      failed=1
    else
      echo "[ok] ${pair#*:} ($h2)"
    fi
  done
  exit "$failed"
fi

if [ -x "$INSTALL_DIR/AI-Terminal-Pet" ]; then
  "$INSTALL_DIR/AI-Terminal-Pet" --quit >/dev/null 2>&1 || true
fi

mkdir -p "$INSTALL_DIR"
chmod 700 "$INSTALL_DIR" 2>/dev/null || true

for pair in $FILES; do
  src="$REPO_ROOT/${pair%%:*}"
  dst="$INSTALL_DIR/${pair#*:}"
  tmp="$dst.tmp.$$"
  cp "$src" "$tmp"
  case "${pair#*:}" in
    protocol.js)
      chmod 0644 "$tmp"
      ;;
    *)
      chmod 0755 "$tmp"
      ;;
  esac
  mv -f "$tmp" "$dst"
  echo "[ok] installed $dst"
done

if [ "$LINK_BIN" -eq 1 ]; then
  mkdir -p "$(dirname "$BIN_LINK")"
  ln -sf "$INSTALL_DIR/AI-Terminal-Pet" "$BIN_LINK"
  echo "[ok] symlinked $BIN_LINK -> $INSTALL_DIR/AI-Terminal-Pet"
fi

echo
echo "Installed to: $INSTALL_DIR"
echo "Run diagnostic: $INSTALL_DIR/AI-Terminal-Pet --doctor"
echo "Print terminal snippets: $INSTALL_DIR/AI-Terminal-Pet --connect"
