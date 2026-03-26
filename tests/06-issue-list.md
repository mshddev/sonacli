# 06 Issue List Command

This case file defines end-to-end coverage for the `sonacli issue list` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Ensure local SonarQube is `UP`.
- Ensure the local sample project is already seeded before running the cases.
- Save a valid auth config inside the isolated `HOME` before success cases.
- Write the final Markdown report to `tests/results/06-issue-list-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### ISSUE-LIST-01 List Issues By Project Key

Run the issue list command with the seeded sample project key.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-02 List Issues In Pretty JSON

Run the issue list command with `--pretty`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --pretty >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-03 List Issues With Explicit Pagination

Run the issue list command with `--page` and `--page-size` to fetch a non-first page.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&issueStatuses=OPEN%2CCONFIRMED&p=2&ps=2" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --page 2 --page-size 2 >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-04 Reject Unknown Project Key

Run the issue list command with a project key that does not exist.

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

./tests/bin/sonacli issue list sonacli-sample-does-not-exist >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
test ! -s "$STDOUT_FILE"
grep -Fqx "Error: project not found: sonacli-sample-does-not-exist" "$STDERR_FILE"
grep -Fqx "Usage:" "$STDERR_FILE"
grep -Fqx "  sonacli issue list <project-key> [flags]" "$STDERR_FILE"
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
```

### ISSUE-LIST-05 Reject Page Beyond The Last Page

Run the issue list command with a page number that is beyond the last available page.

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

TOTAL_ISSUES="$(
  curl -fsS -H "Authorization: Bearer ${SONACLI_TEST_SONARQUBE_TOKEN}" \
    "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=1" |
    jq -r '.paging.total'
)"
test "$TOTAL_ISSUES" -ge 0

if [ "$TOTAL_ISSUES" -eq 0 ]; then
  REQUESTED_PAGE=2
  EXPECTED_ERROR="Error: page 2 is out of range: there are no issues for this project"
else
  REQUESTED_PAGE="$((TOTAL_ISSUES + 1))"
  EXPECTED_ERROR="Error: page ${REQUESTED_PAGE} is out of range: last page is ${TOTAL_ISSUES}"
fi

./tests/bin/sonacli issue list sonacli-sample-basic --page "$REQUESTED_PAGE" --page-size 1 >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
test ! -s "$STDOUT_FILE"
grep -Fqx "$EXPECTED_ERROR" "$STDERR_FILE"
grep -Fqx "Usage:" "$STDERR_FILE"
grep -Fqx "  sonacli issue list <project-key> [flags]" "$STDERR_FILE"
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
unset TOTAL_ISSUES REQUESTED_PAGE EXPECTED_ERROR
```

### ISSUE-LIST-06 List Issues With Explicit Status Filter

Run the issue list command with `--status OPEN` to override the default filter.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&issueStatuses=OPEN&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --status OPEN >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-07 List Issues With Severity Filter

Run the issue list command with `--severity INFO`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&impactSeverities=INFO&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --severity INFO >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-08 List Issues With Qualities Filter

Run the issue list command with `--qualities MAINTAINABILITY`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&impactSoftwareQualities=MAINTAINABILITY&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --qualities MAINTAINABILITY >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-09 List Issues With Assigned Filter

Run the issue list command with `--assigned true`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&assigned=true&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --assigned true >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-10 List Issues With Assignees Filter

Run the issue list command with `--assignees __me__`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&assignees=__me__&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --assignees __me__ >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ISSUE-LIST-11 List Issues With Me Flag

Run the issue list command with `--me`, which is shorthand for `--assignees __me__`.

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
  "${SONACLI_TEST_SONARQUBE_URL}/api/issues/search?componentKeys=sonacli-sample-basic&additionalFields=_all&assignees=__me__&issueStatuses=OPEN%2CCONFIRMED&p=1&ps=20" |
  jq -c '.' >"$EXPECTED_STDOUT_FILE"
printf '\n' >>"$EXPECTED_STDOUT_FILE"

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli issue list sonacli-sample-basic --me >"$STDOUT_FILE" 2>"$STDERR_FILE"
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
