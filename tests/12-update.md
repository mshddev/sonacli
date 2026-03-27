# 12 Update Command

This case file defines end-to-end coverage for the `sonacli update` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Use a copied binary instead of mutating `./tests/bin/sonacli` directly.
- Verify that the update command does not create any files or directories inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/12-update-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### UPDATE-01 Replace The Current Binary From A Local Fixture Server

Run the update command against a copied binary and a local release fixture server.

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
UPDATE_FIXTURE_ROOT="$(mktemp -d)"
UPDATE_BIN_DIR="$(mktemp -d)"
UPDATE_BINARY="$UPDATE_BIN_DIR/sonacli"
UPDATE_BINARY_REAL="$(cd "$UPDATE_BIN_DIR" && pwd -P)/sonacli"
TEST_VERSION="v9.8.7"

cp ./tests/bin/sonacli "$UPDATE_BINARY"
chmod 755 "$UPDATE_BINARY"

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

mkdir -p "$UPDATE_FIXTURE_ROOT/repos/mshddev/sonacli/releases"
mkdir -p "$UPDATE_FIXTURE_ROOT/download/$TEST_VERSION"
mkdir -p "$UPDATE_FIXTURE_ROOT/$TEST_ASSET_BASENAME"

cat >"$UPDATE_FIXTURE_ROOT/repos/mshddev/sonacli/releases/latest" <<EOF
{"tag_name":"$TEST_VERSION"}
EOF

cat >"$UPDATE_FIXTURE_ROOT/$TEST_ASSET_BASENAME/sonacli" <<EOF
#!/bin/sh
printf '%s\n' "fake sonacli $TEST_VERSION"
EOF

chmod 755 "$UPDATE_FIXTURE_ROOT/$TEST_ASSET_BASENAME/sonacli"

tar -C "$UPDATE_FIXTURE_ROOT" -czf \
  "$UPDATE_FIXTURE_ROOT/download/$TEST_VERSION/$TEST_ASSET_BASENAME.tar.gz" \
  "$TEST_ASSET_BASENAME"

if command -v shasum >/dev/null 2>&1; then
  (
    cd "$UPDATE_FIXTURE_ROOT/download/$TEST_VERSION" &&
      shasum -a 256 "$TEST_ASSET_BASENAME.tar.gz" > checksums.txt
  )
else
  (
    cd "$UPDATE_FIXTURE_ROOT/download/$TEST_VERSION" &&
      sha256sum "$TEST_ASSET_BASENAME.tar.gz" > checksums.txt
  )
fi

cat >"$EXPECTED_STDOUT_FILE" <<EOF
Updated sonacli from dev to $TEST_VERSION.
Path: $UPDATE_BINARY_REAL
EOF

: >"$EXPECTED_STDERR_FILE"
printf '%s\n' "fake sonacli $TEST_VERSION" >"$EXPECTED_VERIFY_STDOUT_FILE"
printf '%s\n' "$HOME" >"$EXPECTED_FILE_LIST"

UPDATE_SERVER_PORT_FILE="$UPDATE_FIXTURE_ROOT/server.port"
UPDATE_SERVER_LOG="$UPDATE_FIXTURE_ROOT/server.log"

python3 - <<'PY' "$UPDATE_FIXTURE_ROOT" "$UPDATE_SERVER_PORT_FILE" >"$UPDATE_SERVER_LOG" 2>&1 &
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

UPDATE_SERVER_PID=$!

while [ ! -s "$UPDATE_SERVER_PORT_FILE" ]; do
  sleep 1
done

UPDATE_SERVER_PORT="$(cat "$UPDATE_SERVER_PORT_FILE")"

SONACLI_INSTALL_API_BASE="http://127.0.0.1:$UPDATE_SERVER_PORT" \
SONACLI_INSTALL_DOWNLOAD_BASE="http://127.0.0.1:$UPDATE_SERVER_PORT/download" \
"$UPDATE_BINARY" update >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?

"$UPDATE_BINARY" >"$VERIFY_STDOUT_FILE" 2>"$VERIFY_STDERR_FILE"
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
printf 'update_exit=%s\n' "$EXIT_CODE"
printf 'verify_exit=%s\n' "$VERIFY_EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
diff -u "$EXPECTED_VERIFY_STDOUT_FILE" "$VERIFY_STDOUT_FILE"
cat "$VERIFY_STDERR_FILE"
diff -u "$EXPECTED_FILE_LIST" "$ACTUAL_FILE_LIST"
```

Remove case-specific files after verification:

```sh
kill "$UPDATE_SERVER_PID"
wait "$UPDATE_SERVER_PID" 2>/dev/null || true
rm -rf "$UPDATE_FIXTURE_ROOT" "$UPDATE_BIN_DIR"
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$VERIFY_STDOUT_FILE" "$VERIFY_STDERR_FILE"
rm -f "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE" "$EXPECTED_VERIFY_STDOUT_FILE"
rm -f "$ACTUAL_FILE_LIST" "$EXPECTED_FILE_LIST"
```
