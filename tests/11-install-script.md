# 11 Install Script

This case file defines end-to-end coverage for the root `install.sh` installer.

## Additional Requirements

- Run from the repository root.
- Use the shared isolated `HOME` approach from [README.md](README.md).
- `python3`, `tar`, and either `shasum` or `sha256sum` must be available.
- Write the final Markdown report to `tests/results/11-install-script-testresults.md`.

## Cases

### INSTALL-SCRIPT-01 Install The Latest Release From A Local Fixture Server

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
VERIFY_STDOUT_FILE="$(mktemp)"
VERIFY_STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
EXPECTED_VERIFY_STDOUT_FILE="$(mktemp)"
ACTUAL_FILE_LIST="$(mktemp)"
EXPECTED_FILE_LIST="$(mktemp)"
INSTALL_FIXTURE_ROOT="$(mktemp -d)"
INSTALL_BIN_DIR="$HOME/.local/bin"
TEST_VERSION="v9.8.7"

case "$(uname -s)" in
  Linux) TEST_GOOS="linux" ;;
  Darwin) TEST_GOOS="darwin" ;;
  *) echo "unsupported OS" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) TEST_GOARCH="amd64" ;;
  arm64|aarch64) TEST_GOARCH="arm64" ;;
  *) echo "unsupported architecture" >&2; exit 1 ;;
esac

TEST_ASSET_BASENAME="sonacli_${TEST_VERSION#v}_${TEST_GOOS}_${TEST_GOARCH}"

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

cat >"$EXPECTED_STDOUT_FILE" <<EOF
installed sonacli $TEST_VERSION to $INSTALL_BIN_DIR/sonacli
add $INSTALL_BIN_DIR to your PATH, for example: export PATH="$INSTALL_BIN_DIR:\$PATH"
EOF

: >"$EXPECTED_STDERR_FILE"
printf '%s\n' "fake sonacli $TEST_VERSION" >"$EXPECTED_VERIFY_STDOUT_FILE"

cat >"$EXPECTED_FILE_LIST" <<EOF
$HOME
$HOME/.local
$HOME/.local/bin
$HOME/.local/bin/sonacli
EOF

INSTALL_SERVER_PORT_FILE="$INSTALL_FIXTURE_ROOT/server.port"
INSTALL_SERVER_LOG="$INSTALL_FIXTURE_ROOT/server.log"

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

INSTALL_SERVER_PID=$!

while [ ! -s "$INSTALL_SERVER_PORT_FILE" ]; do
  sleep 1
done

INSTALL_SERVER_PORT="$(cat "$INSTALL_SERVER_PORT_FILE")"

PATH="/usr/bin:/bin:/usr/sbin:/sbin" \
SONACLI_INSTALL_API_BASE="http://127.0.0.1:$INSTALL_SERVER_PORT" \
SONACLI_INSTALL_DOWNLOAD_BASE="http://127.0.0.1:$INSTALL_SERVER_PORT/download" \
sh ./install.sh --install-dir "$INSTALL_BIN_DIR" >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?

"$INSTALL_BIN_DIR/sonacli" >"$VERIFY_STDOUT_FILE" 2>"$VERIFY_STDERR_FILE"
VERIFY_EXIT_CODE=$?

find "$HOME" -print | LC_ALL=C sort >"$ACTUAL_FILE_LIST"
```

#### Verify

```sh
test "$EXIT_CODE" -eq 0
test "$VERIFY_EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
cmp -s "$VERIFY_STDOUT_FILE" "$EXPECTED_VERIFY_STDOUT_FILE"
test ! -s "$VERIFY_STDERR_FILE"
cmp -s "$ACTUAL_FILE_LIST" "$EXPECTED_FILE_LIST"
```

If any assertion fails, capture the mismatch:

```sh
printf 'install_exit=%s\n' "$EXIT_CODE"
printf 'verify_exit=%s\n' "$VERIFY_EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
diff -u "$EXPECTED_VERIFY_STDOUT_FILE" "$VERIFY_STDOUT_FILE"
cat "$VERIFY_STDERR_FILE"
diff -u "$EXPECTED_FILE_LIST" "$ACTUAL_FILE_LIST"
```

Remove case-specific files after verification:

```sh
kill "$INSTALL_SERVER_PID"
wait "$INSTALL_SERVER_PID" 2>/dev/null || true
rm -rf "$INSTALL_FIXTURE_ROOT"
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$VERIFY_STDOUT_FILE" "$VERIFY_STDERR_FILE"
rm -f "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE" "$EXPECTED_VERIFY_STDOUT_FILE"
rm -f "$ACTUAL_FILE_LIST" "$EXPECTED_FILE_LIST"
```
