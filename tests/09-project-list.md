# 09 Project List Command

This case file defines end-to-end coverage for the `sonacli project list` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Ensure local SonarQube is `UP`.
- Ensure the local sample project is already seeded before running the cases.
- Save a valid auth config inside the isolated `HOME` before success cases.
- Write the final Markdown report to `tests/results/09-project-list-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### PROJECT-LIST-01 List Projects

Run the project list command with default pagination.

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

curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
  "${SONACLI_TEST_SONARQUBE_URL}/api/projects/search?p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli project list >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### PROJECT-LIST-02 List Projects In Pretty JSON

Run the project list command with `--pretty`.

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

curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
  "${SONACLI_TEST_SONARQUBE_URL}/api/projects/search?p=1&ps=20" |
  jq '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli project list --pretty >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### PROJECT-LIST-03 List Projects With Explicit Pagination

Run the project list command with `--page` and `--page-size`.

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

curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
  "${SONACLI_TEST_SONARQUBE_URL}/api/projects/search?p=1&ps=1" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli project list --page 1 --page-size 1 >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### PROJECT-LIST-04 Reject Page Beyond The Last Page

Run the project list command with a page number that is beyond the last available page.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"

CONFIG_DIR="$HOME/.sonacli"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"
cat >"$CONFIG_FILE" <<EOF
server_url: "${SONACLI_TEST_SONARQUBE_URL}"
token: "${SONACLI_TEST_SONARQUBE_TOKEN}"
EOF
chmod 600 "$CONFIG_FILE"

TOTAL_PROJECTS="$(
  curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
    "${SONACLI_TEST_SONARQUBE_URL}/api/projects/search?p=1&ps=1" |
    jq -r '.paging.total'
)"
test "$TOTAL_PROJECTS" -ge 0

if [ "$TOTAL_PROJECTS" -eq 0 ]; then
  REQUESTED_PAGE=2
  EXPECTED_ERROR="Error: page 2 is out of range: there are no projects"
else
  REQUESTED_PAGE="$((TOTAL_PROJECTS + 1))"
  EXPECTED_ERROR="Error: page ${REQUESTED_PAGE} is out of range: last page is ${TOTAL_PROJECTS}"
fi

./tests/bin/sonacli project list --page "$REQUESTED_PAGE" --page-size 1 >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
test ! -s "$STDOUT_FILE"
grep -Fqx "$EXPECTED_ERROR" "$STDERR_FILE"
grep -Fqx "Usage:" "$STDERR_FILE"
grep -Fqx "  sonacli project list [flags]" "$STDERR_FILE"
test -f "$CONFIG_FILE"
test "$FILE_COUNT" -eq 1
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
printf 'stdout:\n'
cat "$STDOUT_FILE"
printf '\nstderr:\n'
cat "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 5 -print
```

Remove case-specific files after verification:

```sh
rm -f "$CONFIG_FILE"
rmdir "$CONFIG_DIR" 2>/dev/null || true
rmdir "$(dirname "$CONFIG_DIR")" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE"
unset TOTAL_PROJECTS REQUESTED_PAGE EXPECTED_ERROR
```
