# 04 Auth Status Command

This case file defines end-to-end coverage for the `sonacli auth status` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Verify stdout, stderr, exit code, and config-file side effects for every case.
- Write the final Markdown report to `tests/results/04-auth-status-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### AUTH-STATUS-01 Report Saved Authentication

Run the auth status command after saving auth config into the isolated `HOME`.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

CONFIG_DIR="$HOME/.sonacli"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

cat >"$CONFIG_FILE" <<'EOF'
server_url: "http://127.0.0.1:9000"
token: "test-token"
EOF
chmod 600 "$CONFIG_FILE"

cat >"$EXPECTED_STDOUT_FILE" <<'EOF'
SonarQube authentication is configured.
EOF

cat >>"$EXPECTED_STDOUT_FILE" <<EOF
Config file: ${CONFIG_FILE}
Server URL: http://127.0.0.1:9000
Token: test-*****

Useful commands:
  sonacli project list
  sonacli issue list <project-key>
  sonacli issue show <issue-key-or-url>
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli auth status >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test -f "$CONFIG_FILE"
test "$FILE_COUNT" -eq 1
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 5 -print
```

Remove case-specific files after verification:

```sh
rm -f "$CONFIG_FILE"
rmdir "$CONFIG_DIR" 2>/dev/null || true
rmdir "$(dirname "$CONFIG_DIR")" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```

### AUTH-STATUS-02 Report Missing Authentication

Run the auth status command with no saved auth config.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

CONFIG_FILE="$HOME/.sonacli/config.yaml"

cat >"$EXPECTED_STDOUT_FILE" <<EOF
SonarQube authentication is not configured.
Config file: ${CONFIG_FILE}
Save credentials with:
  sonacli auth setup --server-url <url> --token <token>
Then verify with:
  sonacli auth status
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli auth status >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test "$FILE_COUNT" -eq 0
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 5 -print
```

Remove case-specific files after verification:

```sh
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```

### AUTH-STATUS-03 Reject Malformed Config

Run the auth status command with a malformed config file.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

CONFIG_DIR="$HOME/.sonacli"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

cat >"$CONFIG_FILE" <<'EOF'
server_url: http://127.0.0.1:9000
token: "test-token"
EOF
chmod 600 "$CONFIG_FILE"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<EOF
Error: load auth config "${CONFIG_FILE}": parse config file: line 1: decode "server_url" value: invalid syntax

Show sonacli authentication status to the SonarQube server

Usage:
  sonacli auth status [flags]

Flags:
  -h, --help   help for status
Examples:
sonacli auth status
  sonacli auth setup --server-url <server-url> --token <token>
EOF

./tests/bin/sonacli auth status >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test -f "$CONFIG_FILE"
test "$FILE_COUNT" -eq 1
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 5 -print
```

Remove case-specific files after verification:

```sh
rm -f "$CONFIG_FILE"
rmdir "$CONFIG_DIR" 2>/dev/null || true
rmdir "$(dirname "$CONFIG_DIR")" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```
