# End-to-End Test Result Template

Use this template for every Markdown report written under `tests/results/`.

## How To Use This Template

- Run `./tests/init-report.sh <case-file>` to generate a report skeleton that follows this structure
- Replace every placeholder wrapped in `<...>`
- Keep the section headings and bullet labels stable
- Keep one detailed subsection per executed case
- Show actual `stdout` and `stderr` even when the case passes
- If a stream is empty or no filesystem side effects occurred, state that explicitly
- If a case fails, keep the same structure and add the exact mismatch details

---

# <NN Case Name> Test Results

## Run Summary

- Run date: <YYYY-MM-DD>
- Working directory: `<absolute repo path>`
- Case file executed: [<case-file>.md](../<case-file>.md)
- Selected case IDs: `<CASE-01>`, `<CASE-02>`
- Overall result: `PASS` or `FAIL`

## Shared Setup Used

- Built the real CLI binary at `./tests/bin/sonacli`
- Used one isolated temporary `HOME` for the full selected run
- Ran all commands from the repository root
- <additional shared setup or infrastructure used>
- Cleaned up per-case temporary capture files after each case
- Cleaned up the isolated `HOME` and removed `./tests/bin/sonacli` after the selected run

## Case Results

| Case ID | Result | Command | Exit Code | Stdout | Stderr | Filesystem |
| --- | --- | --- | --- | --- | --- | --- |
| `<CASE-01>` | `PASS` or `FAIL` | `<exact command>` | expected `<n>`, actual `<n>` | matched or mismatch | matched or mismatch | <filesystem summary> |
| `<CASE-02>` | `PASS` or `FAIL` | `<exact command>` | expected `<n>`, actual `<n>` | matched or mismatch | matched or mismatch | <filesystem summary> |

## Detailed Results

### <CASE-01> <Case Title>

- Command run: `<exact command>`
- Status: `PASS` or `FAIL`
- Exit code: expected `<n>`, actual `<n>`
- Stdout: exact match or mismatch
- Stderr: exact match or mismatch
- Filesystem: <observed filesystem result>
- Unexpected output or exit-code mismatches: none or described below

Actual `stdout`:

```text
<actual stdout, or leave this block empty if stdout was empty>
```

Actual `stderr`:

```text
<actual stderr, or leave this block empty if stderr was empty>
```

Actual filesystem state:

```text
<paths created or modified under the isolated HOME, or leave this block empty if none>
```

If the case failed, include mismatch details:

```diff
<diff between expected and actual output, if applicable>
```

### <CASE-02> <Case Title>

- Command run: `<exact command>`
- Status: `PASS` or `FAIL`
- Exit code: expected `<n>`, actual `<n>`
- Stdout: exact match or mismatch
- Stderr: exact match or mismatch
- Filesystem: <observed filesystem result>
- Unexpected output or exit-code mismatches: none or described below

Actual `stdout`:

```text
<actual stdout, or leave this block empty if stdout was empty>
```

Actual `stderr`:

```text
<actual stderr, or leave this block empty if stderr was empty>
```

Actual filesystem state:

```text
<paths created or modified under the isolated HOME, or leave this block empty if none>
```

If the case failed, include mismatch details:

```diff
<diff between expected and actual output, if applicable>
```

## Cleanup Verification

- `./tests/bin/sonacli` removed after the run: yes or no
- `./tests/bin` directory left in repo state after cleanup: yes or no
- <case-specific cleanup verification, such as token revocation, if applicable>
