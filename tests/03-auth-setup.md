# 03 Auth Setup Command

This case file defines end-to-end coverage for the `sonacli auth setup` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Ensure local SonarQube at `${SONACLI_TEST_SONARQUBE_URL:-http://127.0.0.1:9000}` is already `UP` before running the cases.
- Export `SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD` before running the cases. The default local testing password from `./scripts/local-sonarqube.sh start` is `SonacliAdmin1@`.
- Verify that success cases create exactly one config file inside the isolated `HOME`.
- Verify that failure cases do not create any config file inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/03-auth-setup-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### AUTH-SETUP-01 Save Server URL And Token

Run the auth setup command with a server URL and a real temporary SonarQube token.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
SONACLI_TEST_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL:-http://127.0.0.1:9000}"
EXPECTED_CONFIG_FILE="$(mktemp)"
NORMALIZED_SONACLI_TEST_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL%/}"
SONACLI_TEST_SONARQUBE_ADMIN_LOGIN="${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN:-admin}"
SONACLI_TEST_TOKEN_NAME="sonacli-auth-setup-$(date +%s)"

SONACLI_TEST_VALID_TOKEN="$(
  curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
    -X POST "${SONACLI_TEST_SONARQUBE_URL}/api/user_tokens/generate" \
    --data-urlencode "name=${SONACLI_TEST_TOKEN_NAME}" |
    sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
)"
test -n "$SONACLI_TEST_VALID_TOKEN"

CONFIG_DIR="$HOME/.sonacli"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

cat >"$EXPECTED_STDOUT_FILE" <<EOF
Saved SonarQube authentication settings.
Config file: ${CONFIG_FILE}
Server URL: ${NORMALIZED_SONACLI_TEST_SONARQUBE_URL}
Next:
  sonacli auth status
  sonacli project list
  sonacli issue list <project-key>
EOF

: >"$EXPECTED_STDERR_FILE"

cat >"$EXPECTED_CONFIG_FILE" <<EOF
server_url: "${NORMALIZED_SONACLI_TEST_SONARQUBE_URL}"
token: "${SONACLI_TEST_VALID_TOKEN}"
EOF

./tests/bin/sonacli auth setup -s "${SONACLI_TEST_SONARQUBE_URL}/" -t "$SONACLI_TEST_VALID_TOKEN" >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test -f "$CONFIG_FILE"
cmp -s "$CONFIG_FILE" "$EXPECTED_CONFIG_FILE"
test "$FILE_COUNT" -eq 1
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
diff -u "$EXPECTED_CONFIG_FILE" "$CONFIG_FILE"
find "$HOME" -mindepth 1 -maxdepth 5 -print
```

Remove case-specific files after verification:

```sh
curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
  -X POST "${SONACLI_TEST_SONARQUBE_URL}/api/user_tokens/revoke" \
  --data-urlencode "name=${SONACLI_TEST_TOKEN_NAME}" >/dev/null
rm -f "$CONFIG_FILE"
rmdir "$CONFIG_DIR" 2>/dev/null || true
rmdir "$(dirname "$CONFIG_DIR")" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE" "$EXPECTED_CONFIG_FILE"
```

### AUTH-SETUP-02 Reject Invalid Token

Run the auth setup command with an invalid token.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
SONACLI_TEST_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL:-http://127.0.0.1:9000}"

CONFIG_FILE="$HOME/.sonacli/config.yaml"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: token is not valid for the SonarQube server

Setup the SonarQube server URL and token for sonacli

Usage:
  sonacli auth setup [flags]

Flags:
  -h, --help                help for setup
  -s, --server-url string   SonarQube server URL
  -t, --token string        SonarQube user token
Examples:
  sonacli auth setup --server-url <server-url> --token <token>
  sonacli auth status
EOF

./tests/bin/sonacli auth setup -s "$SONACLI_TEST_SONARQUBE_URL" -t invalid-token >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test ! -e "$CONFIG_FILE"
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

### AUTH-SETUP-03 Missing Required Flags

Run the auth setup command with no flags.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

CONFIG_FILE="$HOME/.sonacli/config.yaml"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: required flag(s) "server-url", "token" not set

Setup the SonarQube server URL and token for sonacli

Usage:
  sonacli auth setup [flags]

Flags:
  -h, --help                help for setup
  -s, --server-url string   SonarQube server URL
  -t, --token string        SonarQube user token
Examples:
  sonacli auth setup --server-url <server-url> --token <token>
  sonacli auth status
EOF

./tests/bin/sonacli auth setup >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test ! -e "$CONFIG_FILE"
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

### AUTH-SETUP-04 Reject Invalid Server URL

Run the auth setup command with a malformed server URL.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

CONFIG_FILE="$HOME/.sonacli/config.yaml"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: server-url must use http or https

Setup the SonarQube server URL and token for sonacli

Usage:
  sonacli auth setup [flags]

Flags:
  -h, --help                help for setup
  -s, --server-url string   SonarQube server URL
  -t, --token string        SonarQube user token
Examples:
  sonacli auth setup --server-url <server-url> --token <token>
  sonacli auth status
EOF

./tests/bin/sonacli auth setup -s 127.0.0.1:9000 -t test-token >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test ! -e "$CONFIG_FILE"
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

### AUTH-SETUP-05 Surface Validation Errors From SonarQube

Run the auth setup command with a reachable URL that is not the SonarQube API root.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
SONACLI_TEST_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL:-http://127.0.0.1:9000}"
SONACLI_TEST_BROKEN_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL%/}/not-sonarqube"

CONFIG_FILE="$HOME/.sonacli/config.yaml"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: validate token with SonarQube: decode auth validation response: invalid character '<' looking for beginning of value

Setup the SonarQube server URL and token for sonacli

Usage:
  sonacli auth setup [flags]

Flags:
  -h, --help                help for setup
  -s, --server-url string   SonarQube server URL
  -t, --token string        SonarQube user token
Examples:
  sonacli auth setup --server-url <server-url> --token <token>
  sonacli auth status
EOF

./tests/bin/sonacli auth setup -s "$SONACLI_TEST_BROKEN_SONARQUBE_URL" -t test-token >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test ! -e "$CONFIG_FILE"
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
