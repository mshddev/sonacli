---
name: sonacli
description: Use when working with SonarQube analysis reports via the sonacli CLI — setting up auth, listing projects, listing project issues, or showing a specific issue by key or URL. Triggers on "sonarqube issues", "sonar report", "list sonarqube issues for project", "show sonarqube issue", "list sonarqube projects", "sonacli", or when the user references a SonarQube issue URL.
---

# Sonacli

## Preconditions

- Confirm the CLI is available with `command -v sonacli`.
- Check saved credentials with `sonacli auth status`.
- If authentication is missing, use `sonacli auth setup -s <url> -t <token>`.

## Recommended flow

1. Run `sonacli auth status`.
2. If needed, run `sonacli auth setup -s <url> -t <token>`.
3. Run `sonacli project list` to discover project keys.
4. Run `sonacli issue list <project-key>` to inspect the raw search response for that project.
5. Run `sonacli issue show <issue-key-or-url>` when you need one issue object.

## Command behavior

- `sonacli project list` returns the raw JSON payload from SonarQube project search.
- `sonacli issue list <project-key>` returns the raw JSON payload from SonarQube issue search.
- `sonacli issue show <issue-id-or-url>` returns one issue object as JSON.
- `sonacli issue show` accepts plain keys or SonarQube URLs with `?issue=`, `?issues=`, or `?open=` query parameters.
- JSON commands default to compact JSON. Add `--pretty` for readable output when the user wants it.

## Good defaults

- Prefer `project list` before guessing a project key.
- Prefer passing a full issue URL directly to `issue show` when the user already provided one.
- Preserve compact JSON unless the user asked for pretty output or human-readable formatting.

## Working rules
- If the saved token or server is wrong, surface the error and ask for corrected credentials instead of retrying blindly.
