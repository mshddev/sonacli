# 02 Version Command

This case file defines end-to-end coverage for the `sonacli version` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Verify that the version command does not create any files or directories inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/02-version-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### VERSION-01 Print Default Version

Run the version command with no additional arguments.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

cat >"$EXPECTED_STDOUT_FILE" <<'EOF'
sonacli dev
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli version >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test -z "$(find "$HOME" -mindepth 1 -print -quit)"
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 3 -print
```

Remove case-specific files after verification:

```sh
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```
