# Contributing
This file contains agent-facing contribution and local setup instructions.

These requirements describe the supported local development environment for contributors. They do not redefine the end-user runtime support matrix for `sonacli`; the CLI itself targets Linux and macOS as documented in `README.md`.

## Requirements

This local contributor workflow is currently supported on macOS only.

- macOS
- Go 1.26+
- Podman (for running local SonarQube Community Build v25.x instance)


## Setup

### Podman

Install Podman:

```sh
brew install podman
```

Create the default machine if it does not already exist, use at least `3072` MiB:

```sh
podman machine init --memory 3072 podman-machine-default
```

Start the default machine:

```sh
podman machine start podman-machine-default
```

If the machine already exists with less memory:

```sh
podman machine stop podman-machine-default
podman machine set --memory 3072 podman-machine-default
podman machine start podman-machine-default
```

Verify Podman is ready:

```sh
podman info
```


*!Important:* If you find any issues with the Podman machine, just stop and immediately ask your human to fix or set up the machine.

### Local SonarQube Bootstrap

```sh
./scripts/local-sonarqube.sh start
```

Verify SonarQube is up:

```sh
curl -fsS http://127.0.0.1:9000/api/system/status
```

This should return JSON containing `"status":"UP"`.

#### Local SonarQube Runtime Defaults

| Setting | Value |
|---|---|
| Image | `docker.io/library/sonarqube:25.10.0.114319-community` |
| Local URL | `http://127.0.0.1:9000` |
| Container name | `sonacli-sonarqube` |
| Admin login | `admin` |
| Admin password | `SonacliAdmin1@` |

### Seed The Local Sample Project

Run the sample-project helper after SonarQube is `UP`:

```sh
./scripts/seed-sample-project.sh
```

This helper:

- resets the repo-owned sample project key `sonacli-sample-basic`
- analyzes `tests/fixtures/sample-basic`
- downloads a pinned SonarScanner CLI into `tests/bin/sample-project-tools/` on first use
- waits for the background analysis task to finish

The first run needs network access plus `curl` and `unzip`.

The helper prints the project dashboard URL and useful API endpoints after the analysis completes.

## Build

```sh
make build                # version from git tag (e.g. v0.1.0 or v0.1.0-3-gabc1234)
make build VERSION=v1.0.0 # explicit version override
```

The built binary is written to `./sonacli`.

## Workflow
- Any behavior change must ship with unit coverage and matching end-to-end coverage: add a new E2E case for new behavior, and update existing E2E cases when behavior changes.
- Follow `tests/README.md` for end-to-end test execution and reporting

## Branch workflow

`main` is protected. Do not commit directly on `main` and do not push changes
straight to `main`.

Before making changes:

```sh
git switch main
git pull --ff-only origin main
git switch -c <branch-name>
```

Do all commits on the topic branch. Push the branch and merge through a pull
request instead of direct pushes to `main`.

## Changelog

`CHANGELOG.md` is the source of truth for user-facing release notes and the
body of each GitHub release.

### When to update it

Update `CHANGELOG.md` in the same change whenever you modify:

- CLI behavior, flags, output, or configuration
- supported platforms, version support, install flow, or packaged artifacts
- security guidance or other release-facing project policy

You can usually skip changelog updates for internal-only refactors, test-only
changes, and contributor workflow edits that do not affect shipped artifacts or
how users consume the project.

### How to update it

- Add concise bullets under `## [Unreleased]`.
- Keep entries focused on the user or maintainer-visible outcome, not the
  implementation details.
- Replace `- Nothing yet.` when there is real release content.
- Do not create a tag unless the matching changelog entry already exists.

### Release flow

When preparing a release tag such as `v1.2.3` or `v1.2.3-rc.1`:

1. Prepare the release changes on a non-`main` branch.
2. Replace the `## [Unreleased]` heading with `## [<tag>] - YYYY-MM-DD`.
3. Insert a fresh `## [Unreleased]` section above it with `- Nothing yet.`.
4. Merge the release-preparation branch to `main`.
5. Create and push the tag from the merged `main` commit only after the
   changelog contains that exact heading.

The release workflow reads release notes from the matching section in
`CHANGELOG.md`. If the heading is missing or empty, the tagged GitHub release
job fails.

### Release runbook

Use this flow for a manual release and as the agent procedure for release
requests.

1. Determine the exact release tag first, for example `v0.1.0` or
   `v0.1.0-rc.1`.
2. Start from an up-to-date `main` and create a dedicated release branch.
3. Review `CHANGELOG.md` and make sure `## [Unreleased]` contains the release
   notes to ship. If it still says `- Nothing yet.`, stop and write the release
   notes before continuing.
4. Replace the `## [Unreleased]` heading with the final version heading
   `## [<tag>] - YYYY-MM-DD`.
5. Insert a fresh `## [Unreleased]` section above it with `- Nothing yet.`.
6. Run the required validation for the release candidate.
   At minimum: `make test`
   For behavior changes: follow `tests/README.md` and run the affected E2E
   coverage as well.
7. Commit the release preparation changes, including the changelog, on the
   release branch.
8. Push the release branch and open or update the pull request.
9. Merge the release branch to `main`.
10. Update local `main` to the merged commit.
11. Create the git tag locally from `main`, for example `git tag v0.1.0`.
12. Push only the tag, for example `git push origin v0.1.0`.
13. Monitor the GitHub Actions release workflow. It builds archives, extracts
   release notes from `CHANGELOG.md`, and publishes the GitHub release.

If the release workflow fails because the changelog section is missing or
empty, fix `CHANGELOG.md`, commit the correction, and rerun the release by
updating the release notes or pushing a corrected tag as appropriate.
