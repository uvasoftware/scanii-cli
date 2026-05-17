# Changelog

## [1.7.0]

### Added

- Homebrew install path. Goreleaser now publishes the formula to `scanii/homebrew-tap` on every release: `brew install scanii/tap/scanii-cli`.
- Shell installer at `install.sh`. POSIX-compatible script that detects OS/arch, fetches the matching release archive, verifies the SHA-256, and installs `sc` to `~/.local/bin` (override via `SCANII_CLI_BIN_DIR`, pin a version via `SCANII_CLI_VERSION`): `curl -fsSL https://raw.githubusercontent.com/scanii/scanii-cli/main/install.sh | sh`.

### Changed

- `.goreleaser.yaml` now pins the build matrix explicitly to `amd64`+`arm64` across all OSes, so the shell installer can rely on a stable archive-name contract. The `386` archives (linux-386, windows-386) are no longer produced — no GitHub-hosted runner uses them and they had no known consumers.

## [1.6.0]

### Added

- CORS support on the mock server, matching `api.scanii.com` in production (`Allow-Origin: *`, `Allow-Methods: GET, POST, HEAD, OPTIONS, DELETE`, `Allow-Headers: Authorization, User-Agent`, `Max-Age: 300`). OPTIONS preflight short-circuits with 200, no credentials required. Lets browser-based clients call the mock from a different origin.

## [1.5.0]

### Added

- `sc files trace <id>` command — wraps the `GET /v2.2/files/{id}/trace` endpoint and prints the events as a `timestamp / message` table.
- `Client.RetrieveTrace(ctx, id)` in `internal/client`.

## [1.4.0]

### Added

- `GET /v2.2/files/{id}/trace` mock-server endpoint.
- `location` field support for `POST /v2.2/files`.

## [1.3.1]

### Added

- `/healthcheck` route used by the Docker container health check.

### Changed

- Improved terminal output formatting and warning labels.

## [1.3.0]

### Changed

- Docker image improvements.

## [1.2.0]

### Removed

- Dependabot config.

### Changed

- Docker image namespace updated in the Goreleaser config.

## [1.1.1]

### Changed

- Expanded README with usage examples and a CI guide.

## [1.1.0]

### Added

- Callback delivery support in the local server.
- Embedded test asset fixtures.

### Fixed

- Panic when `/tmp` does not exist inside the Docker container.

## [1.0.0]

First stable release published under the `scanii/scanii-cli` repo.

### Added

- Multiple profile support — named profiles via `sc profile create [name]` and `-p, --profile` global flag.

## [0.1.x and earlier]

Pre-1.0 releases lived under `uvasoftware/scanii-cli`. See the [GitHub releases page](https://github.com/scanii/scanii-cli/releases?q=v0.) for details.
