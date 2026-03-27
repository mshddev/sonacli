# Install Script End-to-End Case

This case file defines end-to-end coverage for the root `install.sh` installer.

## Preconditions

- Run from the repository root.
- Keep one isolated temporary `HOME` for the full run.
- `python3`, `tar`, and either `shasum` or `sha256sum` must be available.

## Shared Setup

```sh
export ORIGINAL_HOME="${HOME:-}"
export TEST_HOME="$(mktemp -d)"
export HOME="$TEST_HOME"
export INSTALL_FIXTURE_ROOT="$(mktemp -d)"
export INSTALL_STDOUT_FILE="$(mktemp)"
export INSTALL_STDERR_FILE="$(mktemp)"
export INSTALL_BIN_DIR="$HOME/.local/bin"
export TEST_VERSION="v9.8.7"

case "$(uname -s)" in
  Linux) export TEST_GOOS="linux" ;;
  Darwin) export TEST_GOOS="darwin" ;;
  *) echo "unsupported OS" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) export TEST_GOARCH="amd64" ;;
  arm64|aarch64) export TEST_GOARCH="arm64" ;;
  *) echo "unsupported architecture" >&2; exit 1 ;;
esac

export TEST_ASSET_BASENAME="sonacli_${TEST_VERSION#v}_${TEST_GOOS}_${TEST_GOARCH}"

mkdir -p "$INSTALL_FIXTURE_ROOT/repos/mshddev/sonacli/releases"
mkdir -p "$INSTALL_FIXTURE_ROOT/download/$TEST_VERSION"
mkdir -p "$INSTALL_FIXTURE_ROOT/$TEST_ASSET_BASENAME"

cat >"$INSTALL_FIXTURE_ROOT/repos/mshddev/sonacli/releases/latest" <<EOF
{"tag_name":"$TEST_VERSION"}
EOF

cat >"$INSTALL_FIXTURE_ROOT/$TEST_ASSET_BASENAME/sonacli" <<EOF
#!/bin/sh
printf '%s\n' "fake sonacli $TEST_VERSION"
EOF

chmod 755 "$INSTALL_FIXTURE_ROOT/$TEST_ASSET_BASENAME/sonacli"

tar -C "$INSTALL_FIXTURE_ROOT" -czf \
  "$INSTALL_FIXTURE_ROOT/download/$TEST_VERSION/$TEST_ASSET_BASENAME.tar.gz" \
  "$TEST_ASSET_BASENAME"

if command -v shasum >/dev/null 2>&1; then
  (
    cd "$INSTALL_FIXTURE_ROOT/download/$TEST_VERSION" &&
      shasum -a 256 "$TEST_ASSET_BASENAME.tar.gz" > checksums.txt
  )
else
  (
    cd "$INSTALL_FIXTURE_ROOT/download/$TEST_VERSION" &&
      sha256sum "$TEST_ASSET_BASENAME.tar.gz" > checksums.txt
  )
fi

export INSTALL_SERVER_PORT_FILE="$INSTALL_FIXTURE_ROOT/server.port"
export INSTALL_SERVER_LOG="$INSTALL_FIXTURE_ROOT/server.log"

python3 - <<'PY' "$INSTALL_FIXTURE_ROOT" "$INSTALL_SERVER_PORT_FILE" >"$INSTALL_SERVER_LOG" 2>&1 &
import functools
import http.server
import pathlib
import socketserver
import sys

root = pathlib.Path(sys.argv[1])
port_file = pathlib.Path(sys.argv[2])
handler = functools.partial(http.server.SimpleHTTPRequestHandler, directory=str(root))

with socketserver.TCPServer(("127.0.0.1", 0), handler) as httpd:
    port_file.write_text(str(httpd.server_address[1]), encoding="utf-8")
    httpd.serve_forever()
PY

export INSTALL_SERVER_PID=$!

while [ ! -s "$INSTALL_SERVER_PORT_FILE" ]; do
  sleep 1
done

export INSTALL_SERVER_PORT="$(cat "$INSTALL_SERVER_PORT_FILE")"
```

## Case INSTALL-SCRIPT-01

Install the latest release from the local fixture server.

```sh
PATH="/usr/bin:/bin:/usr/sbin:/sbin" \
SONACLI_INSTALL_API_BASE="http://127.0.0.1:$INSTALL_SERVER_PORT" \
SONACLI_INSTALL_DOWNLOAD_BASE="http://127.0.0.1:$INSTALL_SERVER_PORT/download" \
sh ./install.sh --install-dir "$INSTALL_BIN_DIR" >"$INSTALL_STDOUT_FILE" 2>"$INSTALL_STDERR_FILE"
```

Expected:

- Exit code `0`
- `"$INSTALL_BIN_DIR/sonacli"` exists and is executable
- stdout reports the installed path
- stderr is empty

Verify the installed binary:

```sh
"$INSTALL_BIN_DIR/sonacli"
```

Expected stdout:

```text
fake sonacli v9.8.7
```

## Cleanup

```sh
kill "$INSTALL_SERVER_PID"
wait "$INSTALL_SERVER_PID" 2>/dev/null || true
rm -f "$INSTALL_STDOUT_FILE" "$INSTALL_STDERR_FILE"
rm -rf "$INSTALL_FIXTURE_ROOT" "$TEST_HOME"
if [ -n "${ORIGINAL_HOME:-}" ]; then
  export HOME="$ORIGINAL_HOME"
else
  unset HOME
fi
unset ORIGINAL_HOME TEST_HOME INSTALL_FIXTURE_ROOT INSTALL_STDOUT_FILE INSTALL_STDERR_FILE INSTALL_BIN_DIR
unset TEST_VERSION TEST_GOOS TEST_GOARCH TEST_ASSET_BASENAME INSTALL_SERVER_PORT_FILE INSTALL_SERVER_LOG
unset INSTALL_SERVER_PID INSTALL_SERVER_PORT
