# Overview

This is `sonacli`, CLI for consuming SonarQube analysis reports.

The project currently targets:

- SonarQube Community Build v25.x+ only
- Linux and macOS only

# Workflow

- Follow `CONTRIBUTING.md` for repository workflow, local environment setup, and testing policy.
- Follow `tests/README.md` for end-to-end test execution and reporting.
- `main` is protected. Create or switch to a non-`main` branch before making edits or commits.
- Do not push changes directly to `main`; use a branch and merge through the normal review flow.

# Release Requests

- If the user asks to release and does not provide a version, ask for the exact release tag first, for example `v0.1.0` or `v0.1.0-rc.1`.
- After the version is known, follow the release runbook in `CONTRIBUTING.md` instead of inventing a separate flow.
- Treat `CHANGELOG.md` as the source of truth for release notes. Do not create or push a release tag unless the matching changelog section exists and is ready.
- Prepare release changes on a non-`main` branch, merge that branch to `main`, then create and push the tag from the merged `main` commit.
