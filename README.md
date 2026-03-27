# sonacli

`sonacli` is a CLI for consuming SonarQube analysis reports.

> SonarQube is an open-source platform for continuous inspection of code quality and security. It ships in several editions: **Community Build** (free, open-source, single-node), **Server** (commercial), **Data Center** (high-availability), and **Cloud** (SaaS, hosted by Sonar). `sonacli` targets the Community Build only.

## Supported:

- SonarQube Community Build `v25.x+`
- Linux and macOS

## Install

### Quick install

Install the latest GitHub release:

```sh
curl -fsSL https://raw.githubusercontent.com/mshddev/sonacli/main/install.sh | sh
```

Or with `wget`:

```sh
wget -qO- https://raw.githubusercontent.com/mshddev/sonacli/main/install.sh | sh
```

Install a specific release or choose a custom install directory:

```sh
curl -fsSL https://raw.githubusercontent.com/mshddev/sonacli/main/install.sh | sh -s -- --version v0.1.0-rc.2
curl -fsSL https://raw.githubusercontent.com/mshddev/sonacli/main/install.sh | sh -s -- --install-dir "$HOME/.local/bin"
```

The installer downloads the matching GitHub release archive for your OS and CPU, verifies `checksums.txt`, installs `sonacli`, and prints a `PATH` hint when needed.

After installing from a release, update the binary in place with:

```sh
sonacli update
sonacli update --version v0.1.0-rc.3
```

### Build from source

```sh
make build
./sonacli version
```

`make build` stamps the binary version from `git describe --tags --always --dirty`.

### Optional: Go install

Requires Go `1.26+`.

```sh
go install github.com/mshddev/sonacli@latest
```

Plain `go install` builds a working binary, but `sonacli version` reports `dev` unless the build passes version `ldflags`.

## Quick Start

```sh
sonacli auth setup --server-url http://127.0.0.1:9000 --token <token>
sonacli auth status
sonacli project list --pretty
sonacli issue list <project-key> --pretty
sonacli issue show AX1234567890 --pretty
sonacli issue show 'https://sonarqube.example.com/project/issues?id=my-project&issues=AX1234567890' --pretty
```


## Documentation

- [CHANGELOG.md](CHANGELOG.md): user-facing release notes and release history
- [docs/cli.md](docs/cli.md): source-derived command reference, config details, skill installation behavior, and versioning notes
- [CONTRIBUTING.md](CONTRIBUTING.md): local environment setup, SonarQube bootstrap, and repository workflow
- [tests/README.md](tests/README.md): end-to-end execution and reporting rules

## License

MIT. See [LICENSE](LICENSE).
