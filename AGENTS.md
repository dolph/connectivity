# Agent guide

This file is read by AI coding assistants before working in this repo. Human contributors may find it useful too — see `README.md` for the user-facing tool description.

## Project

`connectivity` is a small Go CLI that validates network connectivity at OSI layers 3, 4, and 7. Source layout is flat — every `.go` file lives in the repo root and belongs to `package main`. `README.md` is the source of truth for end-user behavior.

## Development loop

Standard Go workflow. Before opening a PR:

    go vet ./...
    gofmt -l .       # must print nothing
    go build ./...
    go test -race ./...

A handful of tests in `resolver_test.go` and `router_test.go` reach the real network and routing table; they may fail in sandboxed environments. That's a known gap (#24), not a regression. Before treating a test failure as your fault, confirm it doesn't reproduce on `origin/main`.

`./build.sh` runs vet, build with version-injected ldflags, and tests. It's primarily for release builds; daily development is fine with the plain `go` commands above.

## Test-driven development

Default to writing the test first.

1. Write a failing test that captures the bug or the new behavior.
2. Run `go test -run TestName` and confirm it fails *for the reason you expect* — not a compile error, not a typo.
3. Make the smallest change that turns it green.
4. Refactor while green; rerun the test after each step.

This catches a class of mistakes that retroactive tests miss: tests that happen to pass against the broken code (because they don't actually exercise the failure mode), and tests that pass trivially (because they don't assert what they claim to).

For Go specifically:

- Prefer table-driven tests. Each case is one row in a slice of structs; loop with `t.Run(tc.name, ...)` so failures point at the row that failed.
- Tests must be hermetic by default — no real network, no filesystem outside `t.TempDir()`, no routing-table reads. Several existing tests violate this (#24); don't follow that pattern in new code.
- When a fix has no unit-testable seam (the function couples directly to `net.Dial`, the kernel routing table, etc.), it's acceptable to ship the fix without a new test — but say so explicitly in the PR's Test Plan and file a follow-up to add the seam.
- Use descriptive test names: `TestParseDestinations_DropsFirstURLFromConfig` beats `TestParseDestinations_Bug5`.
- Use `t.Cleanup` and `t.TempDir` to manage fixtures; don't leave state behind between cases.

Don't go through the motions. A test that calls the function and `t.Fatal`s on a returned error without asserting on outputs is checklist theater — and worse, it gives false confidence that the behavior is covered.

## Code style

- Idiomatic Go. Prefer the standard library; new third-party deps need justification.
- `fmt.Errorf` with `%w` for wrapping. Older `errors.New(fmt.Sprintf(...))` can be modernized opportunistically — don't churn whole files for it.
- `log.Fatalf` is acceptable only at process startup; library-level code returns errors.
- `os.ReadFile` / `os.WriteFile`, not deprecated `ioutil`.
- `strings.ReplaceAll`, not `strings.Replace(..., -1)`.
- `net.JoinHostPort` when building `host:port` — the codebase is moving toward IPv6 support (#10).
- HTTP requests must close response bodies and set a `Timeout`; see #7.
- Plumb `context.Context` through network operations when the call site can supply one.
- Logging currently uses the `log` package. New structured logging should use `log/slog`.

## Repository conventions

- Branches: `claude/<slug>` for AI-generated work (e.g. `claude/fix-issue-13-toolchain`). Reference the issue number when the change closes one.
- Commits: imperative mood, lowercase first word, no trailing period. Keep the subject under 72 characters.
- PRs: include a Summary and Test Plan. Use `Fixes #N` / `Refs #N`.
- Every PR must carry exactly one `release:*` label. The release workflow reads it at merge time and defaults to `patch` when unlabeled, which means an unlabeled docs/test/CI PR cuts an empty patch release. Apply the label *before* merge:
    - `release:skip` — no user-visible behavior change (tests-only, docs-only, CI/tooling, internal refactor)
    - `release:patch` — bug fix, security patch, or dependency bump with no API / CLI / config-schema change
    - `release:minor` — additive change: new flag, new subcommand, new optional config key
    - `release:major` — breaking change to CLI surface, config schema, exit codes, or emitted metric names
- Don't merge your own PRs. Don't push to `main`. Don't commit generated artifacts (the `connectivity` binary is gitignored).

## Issue triage

Two label dimensions:

- Type: `bug` or `enhancement`.
- Priority: `priority:critical` / `priority:high` / `priority:medium` / `priority:low`.

Rubric:

- `critical` — drop everything; production-impacting.
- `high` — significant correctness, security, or reliability; fix in the next release cycle.
- `medium` — important quality-of-life or prevention work.
- `low` — nice to have.

Check open issues for overlap before filing; cross-reference rather than duplicate.

## Scope discipline

- Single-purpose PRs. A bug fix should not slip in unrelated cleanups; a refactor should not slip in behavior changes.
- If you discover a separate defect while working on something else, file an issue rather than expanding the current PR.
- Don't add backwards-compatibility shims for behavior that has no production users yet.
- Resist the urge to "fix it while I'm in here" if it's not in scope. The cost of an unrelated change is borne by every future reviewer.

## Known traps

Landmines in the current code. Don't reintroduce them after a fix lands, and be aware when touching adjacent code:

- `ParseDestinations` in `connectivity.go` unconditionally skips index 0 of the URL slice — a stale CLI-args convention that silently drops the first config-loaded URL (#5).
- The YAML config loader unmarshals into `map[string]string`, ignoring the typed `Config` struct, so `statsd_host` / `statsd_port` / `statsd_protocol` are silently dropped (#6).
- The HTTPS check uses Go's default `http.Client`: no timeout, no body close, no status-code check, follows up to 10 redirects (#7, #12, #15).
- IPv6 is silently filtered out in three places (`resolver.go`, `destinations.go`, `source.go`); IPv6-only destinations report success without being checked (#10).
- The statsd emitter opens a new connection per metric and has a 100-message queue that blocks callers when full (#11).
- `gopacket/routing` panics on non-Linux at runtime; the project is implicitly Linux-only (#34).

## Out of scope here

- See `README.md` for what the tool does and how end users invoke it.
- See `LICENSE` for licensing terms.
- See `.github/workflows/` for the exact CI and release pipelines.
