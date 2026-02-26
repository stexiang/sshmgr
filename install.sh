#!/usr/bin/env bash
set -euo pipefail

BINARY_NAME="sshmgr"
REPO="${SSHMGR_REPO:-stexiang/sshmgr}"
VERSION="${SSHMGR_VERSION:-latest}"
DEFAULT_INSTALL_DIR="/usr/local/bin"
INSTALL_DIR="${SSHMGR_INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"
USE_SUDO=""
TMP_DIR=""
CURRENT_STEP=0
TOTAL_STEPS=7

C_RESET=""
C_INFO=""
C_WARN=""
C_ERROR=""
C_SUCCESS=""
C_DIM=""

setup_ui() {
  if [ ! -t 1 ]; then
    return
  fi

  if ! command -v tput >/dev/null 2>&1; then
    return
  fi

  local colors
  colors="$(tput colors 2>/dev/null || echo 0)"
  if [ "$colors" -lt 8 ]; then
    return
  fi

  C_RESET="$(tput sgr0)"
  C_INFO="$(tput setaf 6)"
  C_WARN="$(tput setaf 3)"
  C_ERROR="$(tput setaf 1)"
  C_SUCCESS="$(tput setaf 2)"
  C_DIM="$(tput dim)"
}

log() {
  printf '%b[sshmgr-install]%b %s\n' "$C_INFO" "$C_RESET" "$*"
}

warn() {
  printf '%b[sshmgr-install] WARN:%b %s\n' "$C_WARN" "$C_RESET" "$*" >&2
}

success() {
  printf '%b[sshmgr-install] OK:%b %s\n' "$C_SUCCESS" "$C_RESET" "$*"
}

fatal() {
  printf '%b[sshmgr-install] ERROR:%b %s\n' "$C_ERROR" "$C_RESET" "$*" >&2
  exit 1
}

render_step_bar() {
  local width=24
  local filled=$((CURRENT_STEP * width / TOTAL_STEPS))
  local empty=$((width - filled))
  local bar_filled bar_empty

  bar_filled="$(printf '%*s' "$filled" '' | tr ' ' '#')"
  bar_empty="$(printf '%*s' "$empty" '' | tr ' ' '-')"

  printf '[%s%s] %d/%d' "$bar_filled" "$bar_empty" "$CURRENT_STEP" "$TOTAL_STEPS"
}

step() {
  CURRENT_STEP=$((CURRENT_STEP + 1))
  log "$(render_step_bar) $*"
}

run_with_spinner() {
  local message="$1"
  shift

  if [ ! -t 1 ] || [ "${SSHMGR_NO_SPINNER:-0}" = "1" ]; then
    local status
    if "$@"; then
      return 0
    else
      status=$?
      return "$status"
    fi
  fi

  local spin='|/-\\'
  local i=0
  local output_file="${TMP_DIR}/cmd-${RANDOM}.log"

  "$@" >"$output_file" 2>&1 &
  local pid=$!

  while kill -0 "$pid" 2>/dev/null; do
    i=$(((i + 1) % 4))
    printf '\r%b[sshmgr-install]%b %s %s' "$C_DIM" "$C_RESET" "$message" "${spin:$i:1}"
    sleep 0.12
  done

  local status
  if wait "$pid"; then
    status=0
  else
    status=$?
  fi

  if [ "$status" -eq 0 ]; then
    printf '\r\033[K'
    log "${message} done."
    return 0
  fi

  printf '\r\033[K' >&2
  warn "${message} failed."
  if [ -s "$output_file" ]; then
    tail -n 30 "$output_file" >&2
  fi
  return "$status"
}

show_banner() {
  log "----------------------------------------"
  log "sshmgr installer"
  log "Repo: ${REPO}"
  log "Version: ${VERSION}"
  log "----------------------------------------"
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
  if [ "${SSHMGR_SKIP_RELEASE:-0}" = "1" ]; then
    log "Skipping release lookup (SSHMGR_SKIP_RELEASE=1)."
    return 1
  fi

  local base_url asset url archive extract_dir found_bin
  base_url="$(release_base_url)"
  build_asset_candidates

  log "Checking GitHub release assets for ${OS}/${ARCH}..."
  for asset in "${ASSET_CANDIDATES[@]}"; do
    url="${base_url}/${asset}"
    archive="${TMP_DIR}/${asset}"

    if ! curl -fsL --retry 2 --connect-timeout 10 -o "$archive" "$url" >/dev/null 2>&1; then
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
    success "Using prebuilt release asset: ${asset}"
    return 0
  done

  warn "No matching release asset found; switching to source build."
  return 1
}

build_from_source() {
  need_cmd git
  need_cmd go

  local src_dir="${TMP_DIR}/src"
  log "Building from source. This may take a little while..."

  run_with_spinner "Cloning ${REPO}" \
    git clone --depth 1 "https://github.com/${REPO}.git" "$src_dir" \
    || fatal "Failed to clone repository: ${REPO}"

  if [ "$VERSION" != "latest" ]; then
    log "Checking out ${VERSION}..."
    (
      cd "$src_dir"
      git fetch --depth 1 origin "refs/tags/${VERSION}:refs/tags/${VERSION}" >/dev/null 2>&1
      git checkout -q "$VERSION"
    ) || fatal "Failed to checkout version: ${VERSION}"
  fi

  (
    cd "$src_dir"
    run_with_spinner "Compiling ${BINARY_NAME}" \
      go build -o "${TMP_DIR}/${BINARY_NAME}" .
  ) || fatal "go build failed"

  SOURCE_BIN="${TMP_DIR}/${BINARY_NAME}"
  chmod +x "$SOURCE_BIN"
  success "Source build completed."
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

  success "Installed to ${target}"
}

post_install_hint() {
  if ! echo ":$PATH:" | grep -q ":${INSTALL_DIR}:"; then
    warn "${INSTALL_DIR} is not in your PATH. Add this line to your shell profile:"
    printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
  fi

  log "Quick start:"
  printf '  %s\n' "sshmgr --help"
  printf '  %s\n' "sshmgr add myhost --user <user> --host <host>.local"
  printf '  %s\n' "sshmgr ssh myhost"
}

main() {
  setup_ui
  show_banner

  step "Checking prerequisites"
  need_cmd curl

  step "Detecting platform"
  detect_platform
  log "Detected: ${OS}/${ARCH_RAW}"

  step "Selecting install directory"
  pick_install_dir
  log "Install dir: ${INSTALL_DIR}"
  if [ -n "$USE_SUDO" ]; then
    warn "You may be prompted for your sudo password."
  fi

  step "Preparing temporary workspace"
  TMP_DIR="$(mktemp -d)"

  step "Acquiring binary"
  if try_download_release; then
    INSTALL_SOURCE_BIN="$RELEASE_BIN"
  else
    build_from_source
    INSTALL_SOURCE_BIN="$SOURCE_BIN"
  fi

  step "Installing sshmgr"
  install_binary "$INSTALL_SOURCE_BIN"

  step "Finalizing"
  post_install_hint
  success "Installation completed."
}

main "$@"
