# End-to-End Testing Guide

This directory contains agent-facing instructions for end-to-end validation of the `sonacli` executable.

## Testing Approach

- All commands assume the current working directory is the repository root where `go.mod` lives.
- Test the built CLI executable, not `go run`.
- Verify real process behavior: stdout, stderr, exit code, and filesystem side effects.

## Preconditions

Complete the `Local SonarQube Bootstrap` from [CONTRIBUTING.md](../CONTRIBUTING.md) before running end-to-end cases. Confirm SonarQube is UP before any case file that requires live SonarQube access. For case files that depend on the repo-owned sample project, run `./scripts/seed-sample-project.sh` once as shared setup before executing the selected cases.

## Execution Rules

- Always use one isolated temporary `HOME` for the full run.
- Build the CLI once before running cases.
- Build before switching `HOME`, otherwise Go build cache or telemetry files can pollute the isolated test home.
- Execute case commands literally as written; do not reconstruct expected output or rewrite shell snippets.
- Treat stdout and stderr as separate contracts; create fresh capture files for each case.
- Write the final test report as a Markdown file under `tests/results/`.
- Clean up temporary files and any case-specific resources after the run.

## Common Setup

Use this setup once per run before executing cases.

### Working Directory

```sh
pwd
```

Expected shape:

```text
.../sonacli
```

### Build Binary

```sh
go build -o ./tests/bin/sonacli .
```

### Shared Sample Project Setup

For case files that depend on the repo-owned sample project, seed it once before switching `HOME` to the isolated test directory:

```sh
./scripts/seed-sample-project.sh
```

### Isolated Shell Setup

```sh
export ORIGINAL_HOME="${HOME:-}"
export TEST_HOME="$(mktemp -d)"
export HOME="$TEST_HOME"
```

### Live SonarQube Setup

For case files that exercise real SonarQube authentication, export these credentials before running. Defaults match the [local SonarQube runtime defaults](../CONTRIBUTING.md#local-sonarqube-runtime-defaults).

```sh
export SONACLI_TEST_SONARQUBE_URL="${SONACLI_TEST_SONARQUBE_URL:-http://127.0.0.1:9000}"
export SONACLI_TEST_SONARQUBE_ADMIN_LOGIN="${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN:-admin}"
export SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD="${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD:-SonacliAdmin1@}"
export SONACLI_TEST_SONARQUBE_TOKEN_NAME="sonacli-local-$(date +%s)"

export SONACLI_TEST_SONARQUBE_TOKEN="$(
  curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
    -X POST "${SONACLI_TEST_SONARQUBE_URL}/api/user_tokens/generate" \
    --data-urlencode "name=${SONACLI_TEST_SONARQUBE_TOKEN_NAME}" |
    sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
)"
test -n "$SONACLI_TEST_SONARQUBE_TOKEN"
```

If you no longer need the token, revoke it with:

```sh
curl -fsS -u "${SONACLI_TEST_SONARQUBE_ADMIN_LOGIN}:${SONACLI_TEST_SONARQUBE_ADMIN_PASSWORD}" \
  -X POST "${SONACLI_TEST_SONARQUBE_URL}/api/user_tokens/revoke" \
  --data-urlencode "name=${SONACLI_TEST_SONARQUBE_TOKEN_NAME}" >/dev/null
```

Case files that generate temporary tokens are responsible for revoking them during cleanup.

## Case Files

| Case file | Live SonarQube required | Notes |
| --- | --- | --- |
| [01-root.md](01-root.md) | No | CLI-only coverage |
| [02-version.md](02-version.md) | No | CLI-only coverage |
| [03-auth-setup.md](03-auth-setup.md) | Yes | Requires local SonarQube and temporary token setup |
| [04-auth-status.md](04-auth-status.md) | No | Uses saved or malformed config fixtures only |
| [05-issue-show.md](05-issue-show.md) | Yes | Requires saved auth config and the sample project seeded during shared setup |
| [06-issue-list.md](06-issue-list.md) | Yes | Requires saved auth config and the sample project seeded during shared setup |
| [07-skill-install.md](07-skill-install.md) | No | Uses isolated HOME and fake agent binaries on PATH |
| [08-skill-uninstall.md](08-skill-uninstall.md) | No | Uses isolated HOME and managed skill fixtures only |
| [09-project-list.md](09-project-list.md) | Yes | Requires saved auth config and the sample project seeded during shared setup |
| [10-group-commands.md](10-group-commands.md) | No | CLI-only coverage for grouped commands rejecting unknown subcommands |
| [11-install-script.md](11-install-script.md) | No | Uses a local HTTP fixture to validate the release installer script |
| [12-update.md](12-update.md) | No | Uses a local HTTP fixture and a copied binary to validate self-update behavior |
| [13-star.md](13-star.md) | No | Uses a fake browser opener on `PATH` to validate repository opening behavior |

Each case file may define additional requirements beyond the shared setup, read the case file for details.

## Reporting

Write the report to `tests/results/` using [RESULT_TEMPLATE.md](RESULT_TEMPLATE.md) as the structure. To initialize a report skeleton: `./tests/init-report.sh <case-file>`.

Naming convention: `tests/results/<case-file-basename>-testresults.md`

Each case in the report must include:

- Case ID and exact command run
- Pass/fail with expected vs actual exit code
- Actual stdout and stderr (state explicitly if empty)
- Filesystem side effects observed (state explicitly if none)
- Whether temporary resources (tokens, etc.) were cleaned up

Keep the report after the run — it is a test artifact, not a temporary file.

## Cleanup

```sh
rm -f "$STDOUT_FILE" "$STDERR_FILE"
rm -rf "$TEST_HOME"
rm -f ./tests/bin/sonacli
if [ -n "${ORIGINAL_HOME:-}" ]; then
  export HOME="$ORIGINAL_HOME"
else
  unset HOME
fi
unset ORIGINAL_HOME TEST_HOME
```
