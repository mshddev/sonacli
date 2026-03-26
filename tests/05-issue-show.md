# 05 Issue Show Command

This case file defines end-to-end coverage for the `sonacli issue show` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Ensure local SonarQube is `UP`.
- Ensure the local sample project is already seeded before running the cases.
- Save a valid auth config inside the isolated `HOME` before success cases.
- Write the final Markdown report to `tests/results/05-issue-show-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### ISSUE-SHOW-01 Show Issue By Key

Run the issue show command with a live issue key from the seeded sample project.

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
cat >"$CONFIG_FILE" <<EOF
server_url: "${SONACLI_TEST_SONARQUBE_URL}"
token: "${SONACLI_TEST_SONARQUBE_TOKEN}"
EOF
chmod 600 "$CONFIG_FILE"

ISSUE_KEY="$(
  curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
    "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&rules=go:S1135&ps=1" |
    jq -r '.issues[0].key'
)"
test -n "$ISSUE_KEY"
test "$ISSUE_KEY" != "null"

curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?issues=${ISSUE_KEY}&additionalFields=_all&ps=1" |
  jq -c '.issues[0]' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue show "$ISSUE_KEY" >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-SHOW-02 Show Issue By URL In Pretty JSON

Run the issue show command with a SonarQube issue URL and `--pretty`.

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
cat >"$CONFIG_FILE" <<EOF
server_url: "${SONACLI_TEST_SONARQUBE_URL}"
token: "${SONACLI_TEST_SONARQUBE_TOKEN}"
EOF
chmod 600 "$CONFIG_FILE"

ISSUE_KEY="$(
  curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
    "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&rules=go:S1135&ps=1" |
    jq -r '.issues[0].key'
)"
test -n "$ISSUE_KEY"
test "$ISSUE_KEY" != "null"

ISSUE_URL="${SONACLI_TEST_SONARQUBE_URL}/project/issues?id=sonacli-sample-basic&issues=${ISSUE_KEY}&open=${ISSUE_KEY}"

curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?issues=${ISSUE_KEY}&additionalFields=_all&ps=1" |
  jq '.issues[0]' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue show "$ISSUE_URL" --pretty >"$STDOUT_FILE" 2>"$STDERR_FILE"
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
