# 13 Star Command

This case file defines end-to-end coverage for the `sonacli star` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Use a fake browser opener executable earlier on `PATH` so the test does not open a real browser.
- Verify that the star command does not create any files or directories inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/13-star-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### STAR-01 Open The Repository URL With The Platform Browser Command

Run the star command with a fake `open` or `xdg-open` executable on `PATH`.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
EXPECTED_URL_FILE="$(mktemp)"
ACTUAL_URL_FILE="$(mktemp)"
FAKE_BIN_DIR="$(mktemp -d)"
OPEN_LOG_FILE="$(mktemp)"

case "$(uname -s)" in
  Linux) OPENER_NAME="xdg-open" ;;
  Darwin) OPENER_NAME="open" ;;
  *) echo "unsupported OS" >&2; exit 1 ;;
esac

cat >"$FAKE_BIN_DIR/$OPENER_NAME" <<EOF
#!/bin/sh
printf '%s\n' "\$1" >"$OPEN_LOG_FILE"
exit 0
EOF

chmod +x "$FAKE_BIN_DIR/$OPENER_NAME"

cat >"$EXPECTED_STDOUT_FILE" <<'EOF'
Opened https://github.com/mshddev/sonacli in your browser.
EOF

: >"$EXPECTED_STDERR_FILE"
printf '%s\n' 'https://github.com/mshddev/sonacli' >"$EXPECTED_URL_FILE"

PATH="$FAKE_BIN_DIR:/usr/bin:/bin:/usr/sbin:/sbin" ./tests/bin/sonacli star >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
cp "$OPEN_LOG_FILE" "$ACTUAL_URL_FILE"
```

#### Verify

```sh
test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
cmp -s "$ACTUAL_URL_FILE" "$EXPECTED_URL_FILE"
test -z "$(find "$HOME" -mindepth 1 -print -quit)"
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
diff -u "$EXPECTED_URL_FILE" "$ACTUAL_URL_FILE"
find "$HOME" -mindepth 1 -maxdepth 3 -print
```

Remove case-specific files after verification:

```sh
rm -rf "$FAKE_BIN_DIR"
rm -f "$OPEN_LOG_FILE" "$ACTUAL_URL_FILE" "$EXPECTED_URL_FILE"
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```
