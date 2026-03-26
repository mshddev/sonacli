# 08 Skill Uninstall Command

This case file defines end-to-end coverage for the `sonacli skill uninstall` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Write the final Markdown report to `tests/results/08-skill-uninstall-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### SKILL-UNINSTALL-01 Default Uninstall Removes Both Managed Skills

Run the uninstall command after installing both managed skills.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
CODEX_SKILL_DIR="$HOME/.codex/skills/sonacli"
CLAUDE_SKILL_DIR="$HOME/.claude/skills/sonacli"

./tests/bin/sonacli skill install --codex --claude >/dev/null

cat >"$EXPECTED_STDOUT_FILE" <<EOF
Removed sonacli skill for codex.
Path: ${CODEX_SKILL_DIR}
Removed sonacli skill for claude.
Path: ${CLAUDE_SKILL_DIR}
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli skill uninstall >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
test ! -e "$CODEX_SKILL_DIR"
test ! -e "$CLAUDE_SKILL_DIR"
test "$FILE_COUNT" -eq 0
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
find "$HOME" -mindepth 1 -maxdepth 6 -print
```

Remove case-specific files after verification:

```sh
rmdir "$HOME/.codex/skills" 2>/dev/null || true
rmdir "$HOME/.codex" 2>/dev/null || true
rmdir "$HOME/.claude/skills" 2>/dev/null || true
rmdir "$HOME/.claude" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```

### SKILL-UNINSTALL-02 Default Uninstall Reports Missing Skills

Run the uninstall command when no managed skills are installed.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
CODEX_SKILL_DIR="$HOME/.codex/skills/sonacli"
CLAUDE_SKILL_DIR="$HOME/.claude/skills/sonacli"

cat >"$EXPECTED_STDOUT_FILE" <<EOF
sonacli skill is not installed for codex.
Path: ${CODEX_SKILL_DIR}
sonacli skill is not installed for claude.
Path: ${CLAUDE_SKILL_DIR}
EOF

: >"$EXPECTED_STDERR_FILE"

./tests/bin/sonacli skill uninstall >"$STDOUT_FILE" 2>"$STDERR_FILE"
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
find "$HOME" -mindepth 1 -maxdepth 6 -print
```

Remove case-specific files after verification:

```sh
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```
