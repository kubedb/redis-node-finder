# AGENTS.md - redis-node-finder

This file provides instructions for AI coding agents working in this KubeDB Go repository.

## Project Overview

`redis-node-finder` is a small Go CLI binary used by the KubeDB Redis operator stack. It runs inside Redis (or RedisSentinel) pod init containers to discover cluster topology before the redis-server process starts. It:

- Reads the `Redis` / `RedisSentinel` CRD via the Kubernetes API.
- Generates DNS / IP endpoints for every shard replica (or the sentinel replica count).
- Writes results to files under `/tmp/` that the init script consumes to decide whether to join or bootstrap the cluster.

It exists because `peer-finder` runs as a post-start hook and starts the server too early; `redis-node-finder` blocks the init script until topology is known.

## Build & Development Commands

```bash
# Build the binary for the host OS/ARCH (uses the golang-dev Docker image)
make build

# Build for all platforms (linux/amd64, linux/arm, linux/arm64, windows/amd64, darwin/amd64, darwin/arm64)
make all-build

# Format sources (runs goimports/gofmt inside the build image)
make fmt

# Run unit tests
make test
make unit-tests

# Run linter (golangci-lint via build image)
make lint

# License header tooling
make add-license
make check-license

# Verify go.mod/go.sum/vendor are tidy
make verify
make verify-modules

# Full CI bundle (verify + check-license + lint + build)
make ci

# Release builds (gated on APPSCODE_ENV=prod and a git tag)
make release
make qa            # staging multi-arch build

# Clean build artifacts
make clean
```

The default target `make all` runs `fmt build`. `make dev` runs `gen fmt build`.

### Versioning

`VERSION` is derived from `git describe`. Tag builds set `version_strategy=tag`; non-master/release branches use the branch name; otherwise the commit hash is used. Values are linked into `version.go` via `hack/build.sh`.

## Project Structure

```
main.go                                 # Entry point - wires cobra root command
version.go                              # Version vars set at build time via -ldflags
pkg/
  cmds/
    root.go                             # NewRootCmd: registers `version` and `run`
    run.go                              # NewCmdRun: flags + dispatches on --mode
  node-finder/
    redis-finder/redis-finder.go        # Cluster-mode discovery (Redis CRD)
    sentinel-finder/sentinel-finder.go  # Sentinel-mode discovery (RedisSentinel CRD)
hack/
  build.sh                              # Go build with version ldflags
  test.sh                               # `go test` runner
  fmt.sh                                # goimports/gofmt
  license/                              # License header template (ltag)
  gendocs/                              # Cobra docs generator
  scripts/                              # Misc CI helpers
.github/workflows/                      # ci.yml, release.yml, release-tracker.yml
vendor/                                 # Vendored dependencies (go mod vendor)
```

## Key Packages / APIs

- `pkg/cmds` - Cobra command tree. `NewRootCmd()` adds the KubeDB and AppCatalog schemes to `client-go`'s scheme registry on `PersistentPreRun`.
- `pkg/node-finder/redis-finder` - `RedisdNodeFinder` reads a `kubedb.dev/apimachinery/apis/kubedb/v1.Redis` object, waits for all shard pods to receive an IP, then writes:
  - `--master-file` (default `master.txt`) - shard count
  - `--slave-file` (default `slave.txt`) - replicas per shard minus 1
  - `--nodes-file` (default `db-nodes.txt`) - one line per pod: `<podName> <host> <port> <busPort> [<podIP>]`
  - `--initial-master-file` (default `initial-master-nodes.txt`) - pods whose ordinal is `0`
  - `--endpoint-type-file` (default `endpoint-type.txt`) - `ip` or `dns` from `spec.cluster.announce.type`
  - Honors `spec.cluster.announce.shards[].endpoints` (`host:port@busPort`) when present; otherwise falls back to in-cluster pod IP / pod FQDN `<pod>.<db>-pods.<ns>.svc` and default ports `6379` / `16379`.
- `pkg/node-finder/sentinel-finder` - `SentinelReplicaFinder` reads a `RedisSentinel` CRD and writes the replica count to `--sentinel-file` (default `sentinel-replicas.txt`).

### Runtime Inputs

The binary expects to run via `InClusterConfig()` and reads these env vars:

| Env Var | Used By | Purpose |
|---------|---------|---------|
| `NAMESPACE` | redis-finder | Namespace of the Redis CR |
| `DATABASE_NAME` | redis-finder | Name of the Redis CR |
| `DATABASE_GOVERNING_SERVICE` | redis-finder | Governing service name |
| `HOSTNAME` | redis-finder | Current pod name |
| `SENTINEL_NAMESPACE` | sentinel-finder | Namespace of the RedisSentinel CR |
| `SENTINEL_NAME` | sentinel-finder | Name of the RedisSentinel CR |

### CLI

```bash
redis-node-finder version
redis-node-finder run --mode=cluster   # default
redis-node-finder run --mode=sentinel
```

All output file flags are persistent on `run` and accept a bare filename (the code prepends `/tmp/`).

## Testing

```bash
make test          # alias for unit-tests
make unit-tests    # runs hack/test.sh over SRC_PKGS (pkg)
```

There are no `_test.go` files checked in at the time of writing; `make test` is wired but currently a no-op pass. Functional verification happens via the KubeDB Redis operator e2e suite (`hack/e2e.sh`).

## Dependencies

- **Go**: `1.25.5` (toolchain), build image `ghcr.io/appscode/golang-dev:1.25`.
- **CLI**: `github.com/spf13/cobra`, `gomodules.xyz/flags`, `gomodules.xyz/x/version`.
- **Kubernetes**: `k8s.io/client-go v0.34.3`, `k8s.io/apimachinery v0.34.3`, `k8s.io/klog/v2`.
- **KubeDB**: `kubedb.dev/apimachinery` (Redis / RedisSentinel CRDs and clientset).
- **Other**: `kmodules.xyz/client-go`, `kmodules.xyz/custom-resources` (AppCatalog scheme), `kubeops.dev/petset` (PetSet clientset, used to read container ports), `go.bytebuilders.dev/license-verifier`.
- **Vendoring**: dependencies are vendored under `vendor/`. `go.mod` pins a few replacements (`controller-runtime`, `apiserver`, `mergo`, `sprig`) to AppsCode forks.

## Code Conventions

- Apache 2.0 license header required on every Go source file (enforced by `make check-license`; template in `hack/license/`).
- Run `make fmt` before committing; CI runs `gofmt`, `goimports`, and `unparam` via `golangci-lint` (`ADDTL_LINTERS` in the Makefile).
- Vendor must stay in sync: `go mod tidy && go mod vendor` and commit; `make verify-modules` fails the build otherwise.
- Errors are typically handled with `klog.Fatalln` because the binary runs as a one-shot init step; failure should crash the init container and let Kubernetes retry.
- Package names use the hyphenated path with underscore aliases (`redis_finder "kubedb.dev/redis-node-finder/pkg/node-finder/redis-finder"`).
- Output files are always written under `/tmp/`; never accept absolute paths from flags.
- DCO sign-off required on commits (`DCO` file at repo root).
