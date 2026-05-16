#!/bin/sh
#
# Scanii CLI installer.
#
#   curl -fsSL https://raw.githubusercontent.com/scanii/scanii-cli/main/install.sh | sh
#
# Environment overrides:
#   SCANII_CLI_VERSION    Version to install (e.g. "1.6.0"). Defaults to the latest release.
#   SCANII_CLI_BIN_DIR    Directory to install `sc` into. Defaults to "$HOME/.local/bin".
#
set -eu

REPO="scanii/scanii-cli"
BIN_NAME="sc"
BIN_DIR="${SCANII_CLI_BIN_DIR:-$HOME/.local/bin}"
REQUESTED_VERSION="${SCANII_CLI_VERSION:-latest}"

log() {
  printf '%s\n' "$*" >&2
}

err() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"
}

need uname
need tar
need mkdir
need mv
need rm

# Prefer curl, fall back to wget.
if command -v curl >/dev/null 2>&1; then
  DL="curl"
elif command -v wget >/dev/null 2>&1; then
  DL="wget"
else
  err "need either curl or wget to download release artifacts"
fi

# OS detection.
case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)
    err "unsupported OS: $(uname -s). Download a binary from https://github.com/${REPO}/releases instead."
    ;;
esac

# Arch detection.
case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    err "unsupported arch: $(uname -m). Download a binary from https://github.com/${REPO}/releases instead."
    ;;
esac

# Checksum tool.
if command -v sha256sum >/dev/null 2>&1; then
  SHA_TOOL="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA_TOOL="shasum -a 256"
else
  err "need either sha256sum or shasum to verify the download"
fi

fetch() {
  # fetch <url> <output-path>
  if [ "$DL" = "curl" ]; then
    curl -fsSL --retry 3 -o "$2" "$1"
  else
    wget -q -O "$2" "$1"
  fi
}

# Resolve "latest" to a concrete version by following the GitHub redirect.
# /releases/latest redirects to /releases/tag/v<version>.
resolve_latest() {
  url="https://github.com/${REPO}/releases/latest"
  if [ "$DL" = "curl" ]; then
    location=$(curl -fsSLI -o /dev/null -w '%{url_effective}' "$url")
  else
    # wget prints the redirect chain on stderr with --max-redirect=0.
    location=$(wget -q --max-redirect=0 --server-response "$url" 2>&1 | awk '/^  Location:/ {print $2}' | tail -n 1)
  fi
  # location looks like https://github.com/scanii/scanii-cli/releases/tag/v1.6.0
  version=${location##*/tag/}
  # strip leading v if present
  case "$version" in
    v*) version=${version#v} ;;
  esac
  if [ -z "$version" ] || [ "$version" = "$location" ]; then
    err "could not determine the latest release version from $url"
  fi
  printf '%s' "$version"
}

if [ "$REQUESTED_VERSION" = "latest" ]; then
  VERSION=$(resolve_latest)
else
  # Allow callers to pass "v1.6.0" or "1.6.0" interchangeably.
  case "$REQUESTED_VERSION" in
    v*) VERSION=${REQUESTED_VERSION#v} ;;
    *)  VERSION=$REQUESTED_VERSION ;;
  esac
fi

ARCHIVE="scanii-cli-${VERSION}-${OS}-${ARCH}.tar.gz"
BASE_URL="https://github.com/${REPO}/releases/download/v${VERSION}"
ARCHIVE_URL="${BASE_URL}/${ARCHIVE}"
CHECKSUMS_URL="${BASE_URL}/checksums.txt"

TMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t scanii-cli)
trap 'rm -rf "$TMP_DIR"' EXIT INT TERM

log "Installing scanii-cli ${VERSION} for ${OS}/${ARCH}"
log "  source: ${ARCHIVE_URL}"

fetch "$ARCHIVE_URL"   "${TMP_DIR}/${ARCHIVE}"
fetch "$CHECKSUMS_URL" "${TMP_DIR}/checksums.txt"

# Verify checksum.
(
  cd "$TMP_DIR"
  expected=$(awk -v a="$ARCHIVE" '$2 == a {print $1}' checksums.txt)
  if [ -z "$expected" ]; then
    err "checksum for $ARCHIVE not found in checksums.txt"
  fi
  actual=$($SHA_TOOL "$ARCHIVE" | awk '{print $1}')
  if [ "$expected" != "$actual" ]; then
    err "checksum mismatch for $ARCHIVE (expected $expected, got $actual)"
  fi
)

# Extract. Archive wraps contents in a top-level directory.
tar -xzf "${TMP_DIR}/${ARCHIVE}" -C "$TMP_DIR"

EXTRACTED_BIN="${TMP_DIR}/scanii-cli-${VERSION}-${OS}-${ARCH}/${BIN_NAME}"
if [ ! -f "$EXTRACTED_BIN" ]; then
  err "expected binary not found in archive at $EXTRACTED_BIN"
fi

mkdir -p "$BIN_DIR"
mv "$EXTRACTED_BIN" "${BIN_DIR}/${BIN_NAME}"
chmod +x "${BIN_DIR}/${BIN_NAME}"

# Best-effort: clear the macOS quarantine attribute set by Gatekeeper.
if [ "$OS" = "darwin" ] && command -v xattr >/dev/null 2>&1; then
  xattr -d com.apple.quarantine "${BIN_DIR}/${BIN_NAME}" 2>/dev/null || true
fi

log ""
log "Installed ${BIN_NAME} to ${BIN_DIR}/${BIN_NAME}"

case ":${PATH}:" in
  *":${BIN_DIR}:"*)
    log "Run '${BIN_NAME} --help' to get started."
    ;;
  *)
    log "Add ${BIN_DIR} to your PATH to run '${BIN_NAME}' from anywhere, e.g.:"
    log "  echo 'export PATH=\"${BIN_DIR}:\$PATH\"' >> ~/.profile"
    ;;
esac
