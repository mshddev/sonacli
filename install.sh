#!/bin/sh

set -eu

REPO="${SONACLI_INSTALL_REPO:-mshddev/sonacli}"
API_BASE="${SONACLI_INSTALL_API_BASE:-https://api.github.com}"
DOWNLOAD_BASE="${SONACLI_INSTALL_DOWNLOAD_BASE:-https://github.com/${REPO}/releases/download}"
VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-}"
TMPDIR_ROOT="${TMPDIR:-/tmp}"
tmpdir=""

usage() {
  cat <<'EOF'
Install sonacli from a GitHub release.

Usage:
  install.sh [--version <tag>] [--install-dir <dir>]

Options:
  --version <tag>      Install a specific release tag such as v0.1.0
  --install-dir <dir>  Install into a specific directory
  -h, --help           Show this help

Environment:
  VERSION              Same as --version
  INSTALL_DIR          Same as --install-dir
EOF
}

log() {
  printf '%s\n' "$*"
}

fail() {
  printf 'install.sh: %s\n' "$*" >&2
  exit 1
}

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

cleanup() {
  if [ -n "${tmpdir:-}" ] && [ -d "${tmpdir}" ]; then
    rm -rf "${tmpdir}"
  fi
}

expand_home() {
  value=$1
  case "$value" in
    "~/"*)
      [ -n "${HOME:-}" ] || fail "HOME is not set, cannot expand ${value}"
      printf '%s/%s\n' "$HOME" "${value#~/}"
      ;;
    *)
      printf '%s\n' "$value"
      ;;
  esac
}

fetch_text() {
  url=$1
  if have_cmd curl; then
    curl -fsSL "$url"
    return
  fi
  if have_cmd wget; then
    wget -qO- "$url"
    return
  fi
  fail "curl or wget is required"
}

download_to_file() {
  url=$1
  destination=$2
  if have_cmd curl; then
    curl -fsSL "$url" -o "$destination"
    return
  fi
  if have_cmd wget; then
    wget -qO "$destination" "$url"
    return
  fi
  fail "curl or wget is required"
}

fetch_release_tag() {
  url=$1
  json=$(fetch_text "$url") || return 1
  printf '%s' "$json" | tr '\n' ' ' | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p'
}

resolve_version() {
  if [ -n "${VERSION}" ]; then
    printf '%s\n' "${VERSION}"
    return
  fi

  latest_url="${API_BASE%/}/repos/${REPO}/releases/latest"
  if tag=$(fetch_release_tag "$latest_url" 2>/dev/null); then
    if [ -n "$tag" ]; then
      printf '%s\n' "$tag"
      return
    fi
  fi

  releases_url="${API_BASE%/}/repos/${REPO}/releases?per_page=1"
  if tag=$(fetch_release_tag "$releases_url" 2>/dev/null); then
    if [ -n "$tag" ]; then
      printf '%s\n' "$tag"
      return
    fi
  fi

  fail "could not determine a release tag from ${REPO}"
}

resolve_os() {
  case "$(uname -s)" in
    Linux)
      printf '%s\n' "linux"
      ;;
    Darwin)
      printf '%s\n' "darwin"
      ;;
    *)
      fail "unsupported operating system: $(uname -s)"
      ;;
  esac
}

resolve_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      printf '%s\n' "amd64"
      ;;
    arm64|aarch64)
      printf '%s\n' "arm64"
      ;;
    *)
      fail "unsupported architecture: $(uname -m)"
      ;;
  esac
}

resolve_install_dir() {
  if [ -n "${INSTALL_DIR}" ]; then
    expand_home "${INSTALL_DIR}"
    return
  fi

  current_binary=$(command -v sonacli 2>/dev/null || true)
  if [ -n "$current_binary" ]; then
    current_dir=$(dirname "$current_binary")
    if [ -d "$current_dir" ] && [ -w "$current_dir" ]; then
      printf '%s\n' "$current_dir"
      return
    fi
  fi

  if [ -d "/usr/local/bin" ] && [ -w "/usr/local/bin" ]; then
    printf '%s\n' "/usr/local/bin"
    return
  fi

  if [ -n "${HOME:-}" ]; then
    printf '%s\n' "${HOME}/.local/bin"
    return
  fi

  fail "set INSTALL_DIR to a writable directory"
}

make_tmpdir() {
  if dir=$(mktemp -d "${TMPDIR_ROOT%/}/sonacli-install.XXXXXX" 2>/dev/null); then
    printf '%s\n' "$dir"
    return
  fi
  if dir=$(mktemp -d 2>/dev/null); then
    printf '%s\n' "$dir"
    return
  fi
  fail "could not create a temporary directory"
}

sha256_file() {
  file=$1
  if have_cmd sha256sum; then
    sha256sum "$file" | awk '{print $1}'
    return
  fi
  if have_cmd shasum; then
    shasum -a 256 "$file" | awk '{print $1}'
    return
  fi
  if have_cmd openssl; then
    openssl dgst -sha256 "$file" | sed 's/^.*= //'
    return
  fi
  fail "sha256sum, shasum, or openssl is required"
}

verify_checksum() {
  asset_name=$1
  archive_path=$2
  checksums_path=$3

  expected=$(awk -v asset="$asset_name" '$2 == asset { print $1; exit }' "$checksums_path")
  [ -n "$expected" ] || fail "could not find ${asset_name} in checksums.txt"

  actual=$(sha256_file "$archive_path")
  [ "$expected" = "$actual" ] || fail "checksum verification failed for ${asset_name}"
}

path_contains() {
  dir=$1
  case ":${PATH:-}:" in
    *:"$dir":*)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      --version)
        [ "$#" -ge 2 ] || fail "--version requires a value"
        VERSION=$2
        shift 2
        ;;
      --install-dir)
        [ "$#" -ge 2 ] || fail "--install-dir requires a value"
        INSTALL_DIR=$2
        shift 2
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "unknown argument: $1"
        ;;
    esac
  done
}

main() {
  parse_args "$@"

  have_cmd tar || fail "tar is required"
  have_cmd mktemp || fail "mktemp is required"
  have_cmd uname || fail "uname is required"

  version=$(resolve_version)
  os=$(resolve_os)
  arch=$(resolve_arch)
  install_dir=$(resolve_install_dir)
  version_no_v=${version#v}
  asset_basename="sonacli_${version_no_v}_${os}_${arch}"
  archive_name="${asset_basename}.tar.gz"

  tmpdir=$(make_tmpdir)
  trap cleanup EXIT HUP INT TERM

  archive_path="${tmpdir}/${archive_name}"
  checksums_path="${tmpdir}/checksums.txt"
  package_dir="${tmpdir}/${asset_basename}"

  download_to_file "${DOWNLOAD_BASE%/}/${version}/${archive_name}" "$archive_path"
  download_to_file "${DOWNLOAD_BASE%/}/${version}/checksums.txt" "$checksums_path"
  verify_checksum "$archive_name" "$archive_path" "$checksums_path"

  mkdir -p "$install_dir"
  tar -xzf "$archive_path" -C "$tmpdir"
  [ -f "${package_dir}/sonacli" ] || fail "release archive did not contain ${asset_basename}/sonacli"

  cp "${package_dir}/sonacli" "${install_dir}/sonacli"
  chmod 755 "${install_dir}/sonacli"

  log "installed sonacli ${version} to ${install_dir}/sonacli"
  if ! path_contains "$install_dir"; then
    log "add ${install_dir} to your PATH, for example: export PATH=\"${install_dir}:\$PATH\""
  fi
}

main "$@"
