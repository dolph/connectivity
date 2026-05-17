# Contributing to connectivity

Thanks for your interest in improving `connectivity`.

## Development setup

- Go 1.20+ (see `go.mod`)
- Linux is the primary target; routing checks require Linux for full behavior

## Build and test

```bash
./build.sh          # vet, build, and test with coverage
./build.sh bench    # run benchmarks
go test -race ./... # recommended before sending a PR
```

## Pull requests

1. Open an issue first for non-trivial changes so we can agree on approach.
2. Keep PRs focused — one logical change per PR.
3. Add or update tests for behavior changes.
4. Ensure `go vet ./...` and `go test ./...` pass locally.

## Release labels

Merged PRs to `main` are released automatically when CI passes. Add one of these labels to the PR:

| Label | Effect |
|-------|--------|
| `release:patch` | Bug fix (default) |
| `release:minor` | Backwards-compatible feature |
| `release:major` | Breaking change |
| `release:skip` | No release |
