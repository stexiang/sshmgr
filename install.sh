#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="sshmgr"
REPO="${SSHMGR_REPO:-stexiang/sshmgr}"
VERSION="${SSHMGR_VERSION:-latest}"
DEFAULT_INSTALL_DIR="/usr/local/bin"
INSTALL_DIR="${SSHMGR_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
USE_SUDO=""
TMP_DIR=""

log() {
  printf '[sshmgr-install] %s\n' "$*"
}

warn() {
  printf '[sshmgr-install] WARN: %s\n' "$*" >&2
}

fatal() {
  printf '[sshmgr-install] ERROR: %s\n' "$*" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fatal "Required command not found: $1"
}

cleanup() {
  if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
    rm -rf "$TMP_DIR"
  fi
}
trap cleanup EXIT

detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$OS" in
    darwin)
      OS_VARIANTS=(darwin macos)
      ;;
    *)
      fatal "Unsupported OS: $OS (sshmgr currently supports macOS only)"
      ;;
  esac

  ARCH_RAW="$(uname -m)"
  case "$ARCH_RAW" in
    arm64|aarch64)
      ARCH="arm64"
      ARCH_VARIANTS=(arm64 aarch64)
      ;;
    x86_64|amd64)
      ARCH="amd64"
      ARCH_VARIANTS=(amd64 x86_64)
      ;;
    *)
      fatal "Unsupported architecture: $ARCH_RAW"
      ;;
  esac
}

pick_install_dir() {
  if [ -n "${SSHMGR_INSTALL_DIR:-}" ]; then
    if mkdir -p "$INSTALL_DIR" 2>/dev/null; then
      :
    elif command -v sudo >/dev/null 2>&1; then
      USE_SUDO="sudo"
      "$USE_SUDO" mkdir -p "$INSTALL_DIR" || fatal "Cannot create install dir: $INSTALL_DIR"
    else
      fatal "Cannot create install dir without sudo: $INSTALL_DIR"
    fi

    if [ -w "$INSTALL_DIR" ]; then
      return
    fi

    if command -v sudo >/dev/null 2>&1; then
      USE_SUDO="sudo"
      return
    fi

    fatal "Install dir is not writable and sudo is unavailable: $INSTALL_DIR"
  fi

  INSTALL_DIR="$DEFAULT_INSTALL_DIR"
  if [ -d "$INSTALL_DIR" ] && [ -w "$INSTALL_DIR" ]; then
    return
  fi

  if command -v sudo >/dev/null 2>&1; then
    USE_SUDO="sudo"
    return
  fi

  INSTALL_DIR="$HOME/.local/bin"
  mkdir -p "$INSTALL_DIR"
}

release_base_url() {
  if [ "$VERSION" = "latest" ]; then
    printf 'https://github.com/%s/releases/latest/download' "$REPO"
  else
    printf 'https://github.com/%s/releases/download/%s' "$REPO" "$VERSION"
  fi
}

build_asset_candidates() {
  ASSET_CANDIDATES=()
  local osv archv
  for osv in "${OS_VARIANTS[@]}"; do
    for archv in "${ARCH_VARIANTS[@]}"; do
      ASSET_CANDIDATES+=(
        "${BINARY_NAME}_${osv}_${archv}.tar.gz"
        "${BINARY_NAME}-${osv}-${archv}.tar.gz"
        "${BINARY_NAME}_${osv}_${archv}.zip"
        "${BINARY_NAME}-${osv}-${archv}.zip"
        "${BINARY_NAME}_${osv}_${archv}"
        "${BINARY_NAME}-${osv}-${archv}"
      )
    done
  done
}

find_executable() {
  local src_dir="$1"
  local candidate
  while IFS= read -r candidate; do
    if [ -f "$candidate" ]; then
      printf '%s' "$candidate"
      return 0
    fi
  done < <(find "$src_dir" -type f -name "$BINARY_NAME")
  return 1
}

try_download_release() {
  local base_url asset url archive extract_dir found_bin
  base_url="$(release_base_url)"
  build_asset_candidates

  for asset in "${ASSET_CANDIDATES[@]}"; do
    url="${base_url}/${asset}"
    archive="${TMP_DIR}/${asset}"

    log "Trying release asset: ${asset}"
    if ! curl -fsSL --retry 2 --connect-timeout 10 -o "$archive" "$url"; then
      continue
    fi

    extract_dir="${TMP_DIR}/extract"
    rm -rf "$extract_dir"
    mkdir -p "$extract_dir"

    case "$asset" in
      *.tar.gz)
        if ! tar -xzf "$archive" -C "$extract_dir" >/dev/null 2>&1; then
          warn "Failed to extract ${asset}"
          continue
        fi
        ;;
      *.zip)
        if ! unzip -q "$archive" -d "$extract_dir" >/dev/null 2>&1; then
          warn "Failed to extract ${asset}"
          continue
        fi
        ;;
      *)
        cp "$archive" "${extract_dir}/${BINARY_NAME}"
        ;;
    esac

    if ! found_bin="$(find_executable "$extract_dir")"; then
      warn "No ${BINARY_NAME} binary found in ${asset}"
      continue
    fi

    chmod +x "$found_bin"
    RELEASE_BIN="$found_bin"
    log "Downloaded release binary from ${asset}"
    return 0
  done

  return 1
}

build_from_source() {
  need_cmd git
  need_cmd go

  local src_dir="${TMP_DIR}/src"
  log "Building from source (release asset not found)..."
  git clone --depth 1 "https://github.com/${REPO}.git" "$src_dir" >/dev/null 2>&1 \
    || fatal "Failed to clone repository: ${REPO}"

  if [ "$VERSION" != "latest" ]; then
    (
      cd "$src_dir"
      git fetch --depth 1 origin "refs/tags/${VERSION}:refs/tags/${VERSION}" >/dev/null 2>&1
      git checkout -q "$VERSION"
    ) || fatal "Failed to checkout version: ${VERSION}"
  fi

  (
    cd "$src_dir"
    go build -o "${TMP_DIR}/${BINARY_NAME}" .
  ) || fatal "go build failed"

  SOURCE_BIN="${TMP_DIR}/${BINARY_NAME}"
  chmod +x "$SOURCE_BIN"
}

install_binary() {
  local src_bin="$1"
  local target="${INSTALL_DIR}/${BINARY_NAME}"

  if [ "$INSTALL_DIR" = "$DEFAULT_INSTALL_DIR" ] && [ ! -d "$INSTALL_DIR" ]; then
    if [ -n "$USE_SUDO" ]; then
      "$USE_SUDO" mkdir -p "$INSTALL_DIR"
    else
      mkdir -p "$INSTALL_DIR"
    fi
  fi

  if [ -n "$USE_SUDO" ]; then
    "$USE_SUDO" install -m 0755 "$src_bin" "$target"
  else
    install -m 0755 "$src_bin" "$target"
  fi

  log "Installed to ${target}"
}

post_install_hint() {
  if ! echo ":$PATH:" | grep -q ":${INSTALL_DIR}:"; then
    warn "${INSTALL_DIR} is not in your PATH. Add this line to your shell profile:"
    printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
  fi

  log "Done. Run: sshmgr --help"
}

main() {
  need_cmd curl
  detect_platform
  pick_install_dir

  TMP_DIR="$(mktemp -d)"

  if try_download_release; then
    install_binary "$RELEASE_BIN"
  else
    build_from_source
    install_binary "$SOURCE_BIN"
  fi

  post_install_hint
}

main "$@"
