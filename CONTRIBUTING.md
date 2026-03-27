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
