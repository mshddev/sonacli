# 01 Root Command

This case file defines end-to-end coverage for the `sonacli` root command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Verify that the root command does not create any files or directories inside the isolated `HOME`.
- Write the final Markdown report to `tests/results/01-root-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### ROOT-01 Bare Root Command

Run the root command with no arguments.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

cat >"$EXPECTED_STDOUT_FILE" <<'EOF'
CLI for consuming SonarQube reports

Usage:
  sonacli <command> <subcommand> [flags]

Available Commands:
  auth        Manage authentication settings to SonarQube
  issue       Read SonarQube issues
  project     Read SonarQube projects
  skill       Manage agent skills for sonacli
  star        Star the sonacli GitHub repository
  update      Update sonacli to the latest version
  version     Print the sonacli version
Flags:
  -h, --help   help for sonacli
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ROOT-02 Unknown Flag

Run the root command with an unknown flag.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown flag: --wat

CLI for consuming SonarQube reports

Usage:
  sonacli <command> <subcommand> [flags]

Available Commands:
  auth        Manage authentication settings to SonarQube
  issue       Read SonarQube issues
  project     Read SonarQube projects
  skill       Manage agent skills for sonacli
  star        Star the sonacli GitHub repository
  update      Update sonacli to the latest version
  version     Print the sonacli version
Flags:
  -h, --help   help for sonacli
EOF

./tests/bin/sonacli --wat >"$STDOUT_FILE" 2>"$STDERR_FILE"
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

### ROOT-03 Unknown Subcommand

Run the root command with an unknown subcommand.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: unknown command "scan" for "sonacli"

CLI for consuming SonarQube reports

Usage:
  sonacli <command> <subcommand> [flags]

Available Commands:
  auth        Manage authentication settings to SonarQube
  issue       Read SonarQube issues
  project     Read SonarQube projects
  skill       Manage agent skills for sonacli
  star        Star the sonacli GitHub repository
  update      Update sonacli to the latest version
  version     Print the sonacli version
Flags:
  -h, --help   help for sonacli
EOF

./tests/bin/sonacli scan >"$STDOUT_FILE" 2>"$STDERR_FILE"
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
