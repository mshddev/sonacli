# 10 Group Commands

This case file defines end-to-end coverage for grouped commands that should reject unknown subcommands.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Verify that these commands do not create any files or directories inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/10-group-commands-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### GROUP-01 Unknown Auth Subcommand

Run the auth command with an unknown subcommand.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown command "unknown" for "sonacli auth"

Manage authentication settings to SonarQube

Usage:
  sonacli auth <command> [flags]

Available Commands:
  setup       Setup the SonarQube server URL and token for sonacli
  status      Show sonacli authentication status to the SonarQube server
Flags:
  -h, --help   help for auth
EOF

./tests/bin/sonacli auth unknown >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
test "$EXIT_CODE" -eq 1
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

### GROUP-02 Unknown Issue Subcommand

Run the issue command with an unknown subcommand.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown command "unknown" for "sonacli issue"

Read SonarQube issues

Usage:
  sonacli issue <command> [flags]

Available Commands:
  list        List SonarQube issues for a project as JSON
  show        Show a SonarQube issue as JSON
Flags:
  -h, --help   help for issue
EOF

./tests/bin/sonacli issue unknown >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
test "$EXIT_CODE" -eq 1
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

### GROUP-03 Unknown Project Subcommand

Run the project command with an unknown subcommand.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown command "unknown" for "sonacli project"

Read SonarQube projects

Usage:
  sonacli project <command> [flags]

Available Commands:
  list        List SonarQube projects as JSON
Flags:
  -h, --help   help for project
EOF

./tests/bin/sonacli project unknown >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
test "$EXIT_CODE" -eq 1
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

### GROUP-04 Unknown Skill Subcommand

Run the skill command with an unknown subcommand.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown command "unknown" for "sonacli skill"

Manage agent skills for sonacli

Usage:
  sonacli skill <command> [flags]

Available Commands:
  install     Install the sonacli skill for supported agents
  uninstall   Remove the installed sonacli skill from supported agents
Flags:
  -h, --help   help for skill
EOF

./tests/bin/sonacli skill unknown >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
test "$EXIT_CODE" -eq 1
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
