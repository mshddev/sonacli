# Changelog

All notable user-facing changes to `sonacli` are documented in this file.

`CHANGELOG.md` is the source of truth for GitHub release notes. Follow
[CONTRIBUTING.md](CONTRIBUTING.md) for when and how to update it.

## [Unreleased]
- Added a root `install.sh` and simplified `README.md` so users can install `sonacli` directly from `curl` or `wget`.

## [v0.1.0-rc.2] - 2026-03-27

- Added a project changelog and documented release history for `sonacli`.
- GitHub releases now publish notes from `CHANGELOG.md` instead of generated
  release notes.
- Documented the protected-`main` branch workflow and the release runbook for
  contributors and agents.

## [v0.1.0-rc.1] - 2026-03-27

- Initial public pre-release of the `sonacli` CLI for SonarQube Community
  Build `v25.x+`.
- Authentication setup and status commands for storing and validating
  SonarQube access locally.
- Project and issue commands for listing projects, listing project issues, and
  showing a single issue by key or SonarQube URL.
- Skill install and uninstall commands for Codex and Claude Code.
- Source-based CLI reference, contributor setup docs, local SonarQube
  bootstrap helpers, and end-to-end test guidance.
- GitHub Actions CI, release automation, and release archives for Linux and
  macOS.
- Security policy for vulnerability reporting and supported version guidance.

[Unreleased]: https://github.com/mshddev/sonacli/compare/v0.1.0-rc.2...HEAD
[v0.1.0-rc.2]: https://github.com/mshddev/sonacli/releases/tag/v0.1.0-rc.2
[v0.1.0-rc.1]: https://github.com/mshddev/sonacli/releases/tag/v0.1.0-rc.1
