# 07 Skill Install Command

This case file defines end-to-end coverage for the `sonacli skill install` command.

## Additional Requirements

- Use the shared setup from [README.md](README.md).
- Use fresh capture files for each case.
- Use an isolated temporary directory with fake `codex` and `claude` executables when testing default detection.
- Write the final Markdown report to `tests/results/07-skill-install-testresults.md`.
- Do not remove `./tests/bin/sonacli` in this case file. The shared cleanup step in [README.md](README.md) removes the built binary after the selected run finishes.

## Cases

### SKILL-INSTALL-01 Default Install Detects Supported Agents

Run the skill install command with fake `codex` and `claude` executables on `PATH`.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
EXPECTED_FILE_LIST="$(mktemp)"
ACTUAL_FILE_LIST="$(mktemp)"
FAKE_BIN_DIR="$(mktemp -d)"
CODEX_SKILL_DIR="$HOME/.codex/skills/sonacli"
CLAUDE_SKILL_DIR="$HOME/.claude/skills/sonacli"

cat >"$FAKE_BIN_DIR/codex" <<'EOF'
#!/bin/sh
exit 0
EOF

cat >"$FAKE_BIN_DIR/claude" <<'EOF'
#!/bin/sh
exit 0
EOF

chmod +x "$FAKE_BIN_DIR/codex" "$FAKE_BIN_DIR/claude"

cat >"$EXPECTED_STDOUT_FILE" <<EOF
Installed sonacli skill for codex.
Path: ${CODEX_SKILL_DIR}
Installed sonacli skill for claude.
Path: ${CLAUDE_SKILL_DIR}
EOF

: >"$EXPECTED_STDERR_FILE"

cat >"$EXPECTED_FILE_LIST" <<EOF
$CLAUDE_SKILL_DIR/.sonacli-managed
$CLAUDE_SKILL_DIR/SKILL.md
$CODEX_SKILL_DIR/.sonacli-managed
$CODEX_SKILL_DIR/SKILL.md
$CODEX_SKILL_DIR/agents/openai.yaml
EOF

PATH="$FAKE_BIN_DIR:/usr/bin:/bin:/usr/sbin:/sbin" ./tests/bin/sonacli skill install >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
find "$HOME" -type f | LC_ALL=C sort >"$ACTUAL_FILE_LIST"
```

#### Verify

```sh
test "$EXIT_CODE" -eq 0
cmp -s "$STDOUT_FILE" "$EXPECTED_STDOUT_FILE"
cmp -s "$STDERR_FILE" "$EXPECTED_STDERR_FILE"
cmp -s "$ACTUAL_FILE_LIST" "$EXPECTED_FILE_LIST"
```

If any assertion fails, capture the mismatch:

```sh
printf 'exit=%s\n' "$EXIT_CODE"
diff -u "$EXPECTED_STDOUT_FILE" "$STDOUT_FILE"
diff -u "$EXPECTED_STDERR_FILE" "$STDERR_FILE"
diff -u "$EXPECTED_FILE_LIST" "$ACTUAL_FILE_LIST"
find "$HOME" -mindepth 1 -maxdepth 6 -print
```

Remove case-specific files after verification:

```sh
rm -rf "$CODEX_SKILL_DIR" "$CLAUDE_SKILL_DIR" "$FAKE_BIN_DIR"
rmdir "$HOME/.codex/skills" 2>/dev/null || true
rmdir "$HOME/.codex" 2>/dev/null || true
rmdir "$HOME/.claude/skills" 2>/dev/null || true
rmdir "$HOME/.claude" 2>/dev/null || true
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE" "$EXPECTED_FILE_LIST" "$ACTUAL_FILE_LIST"
```

### SKILL-INSTALL-02 Default Install Fails Without Supported Agents

Run the skill install command without `codex` or `claude` on `PATH`.

#### Execute

```sh
STDOUT_FILE="$(mktemp)"
STDERR_FILE="$(mktemp)"
EXPECTED_STDOUT_FILE="$(mktemp)"
EXPECTED_STDERR_FILE="$(mktemp)"
EMPTY_BIN_DIR="$(mktemp -d)"

: >"$EXPECTED_STDOUT_FILE"

cat >"$EXPECTED_STDERR_FILE" <<'EOF'
Error: no supported agent CLI detected on PATH: codex, claude; pass --codex or --claude to install explicitly

Install the sonacli skill for supported agents

Install the managed sonacli skill into the detected agent skill directories. Pass --codex or --claude to skip PATH detection and target a specific agent explicitly.

Usage:
  sonacli skill install [flags]

Flags:
      --claude   Install the skill for Claude Code
      --codex    Install the skill for Codex
  -h, --help     help for install
Examples:
  sonacli skill install
  sonacli skill install --codex
  sonacli skill install --claude
EOF

PATH="$EMPTY_BIN_DIR" ./tests/bin/sonacli skill install >"$STDOUT_FILE" 2>"$STDERR_FILE"
EXIT_CODE=$?
```

#### Verify

```sh
FILE_COUNT="$(find "$HOME" -type f | wc -l | tr -d '[:space:]')"

test "$EXIT_CODE" -eq 1
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
rm -rf "$EMPTY_BIN_DIR"
rm -f "$STDOUT_FILE" "$STDERR_FILE" "$EXPECTED_STDOUT_FILE" "$EXPECTED_STDERR_FILE"
```
