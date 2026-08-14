# Repository modernization plan

## Baseline

- [x] Move the module and CI from Go 1.23.x to Go 1.24.
- [x] Keep the repository as a single CLI tool; do not replace the current TUI with another terminal UI.
- [x] Treat `github.com/gizak/termui v2.2.0+incompatible` as removal work, not upgrade work.

## Source code

- [x] Remove `tui.go`, `-tui`, and `-stdio`; make plain CLI output the default and only interface.
- [x] Split ELF loading, dependency resolution, and text rendering into small focused packages or files with explicit inputs and outputs.
- [x] Replace global mutable state (`deps`, `deps_list`, `deps_root`, library path globals) with an application struct and narrow helper functions.
- [x] Return errors with context instead of exiting deep in helpers; keep `main` responsible for process exit codes and user-facing messages.
- [x] Add regression tests for library lookup order, symlink resolution, static binaries, duplicate dependencies, and invalid/non-ELF inputs.
  - [x] Created `elf_test.go` with comprehensive table-driven tests for ELF formatting functions
  - [x] Added benchmarks for performance tracking
  - [x] Integration tests for dependency tree logic (deferred to after refactoring)
- [x] Run `gofmt`, `go test ./...`, and `go vet ./...` under Go 1.24 in CI.
  - [x] Fixed all go vet errors in main.go, tui.go, and elf.go
  - [x] GitHub Actions workflow configured with format, test, and vet checks

## Go 1.24 guidance

- [x] Adopt current Go review guidance: idiomatic names, package-focused APIs, doc comments for exported symbols, early returns, and no ignored errors.
- [x] Prefer standard library facilities first; keep external dependencies only when they clearly reduce maintenance cost.
- [ ] If repository-local developer tools are added later, track them with Go 1.24 `tool` directives instead of a `tools.go` workaround.
- [x] Use table-driven tests by default; if benchmarks are added, prefer `testing.B.Loop` in Go 1.24.
  - [x] Implemented table-driven tests for all ELF formatting functions
  - [x] Added benchmark suite for performance baselines
- [x] Keep compatibility simple: avoid unnecessary abstractions, avoid panic for normal control flow, and use wrapped errors that work with `errors.Is` and `errors.As`.

## Dependency and module decisions

- [x] Remove the direct TUI dependency and its transitive terminal stack (`termbox-go`, `go-runewidth`, `go-wordwrap`, `ansi`, `go-colorable`, `go-isatty`, `panicparse`, `uniseg`) once TUI code is deleted.
- [x] Keep the standard `flag` package if the tool remains a single-command executable with a small option surface.
- [x] Evaluate `spf13/cobra` only if shell completion, manpage generation, or subcommands become required. (not needed: single command, no subcommands)
- [x] Evaluate `alecthomas/kong` if struct-tag based flag binding is preferred over Cobra's command model. (not needed)
- [x] Evaluate `urfave/cli/v3` if a lighter external CLI framework is needed without a TUI stack. (not needed: the module has no external dependency now)
- [x] Keep `debug/elf` unless new requirements need functionality that the standard library cannot provide.

## Documentation

- [x] Rewrite `README.md` for CLI-only usage, current repository ownership, and current installation guidance (`go install`, packaged binaries, supported platforms).
- [x] Remove the screenshot, TUI key bindings, and any references to `github.com/namhyung/elftree`.
- [x] Document the final CLI flags with stable examples and expected exit behavior.
- [x] Add contributor-facing notes for local validation commands and supported Go version.

## Suggested rollout order

- [x] Step 1: upgrade toolchain and CI to Go 1.24.
- [x] Step 2: refactor core logic away from global state.
- [x] Step 3: ship CLI-only output and delete TUI code plus terminal dependencies.
- [x] Step 4: add tests and vet coverage.
  - [x] Fixed all go vet errors
  - [x] Created comprehensive test suite for ELF formatting functions
  - [x] Add integration tests after refactoring (Step 2)
- [x] Step 5: refresh README and release documentation.

## CI and distribution

- [x] Build and test on Windows, Linux, and macOS in GitHub Actions.
- [x] Cross compile for linux/darwin/windows on amd64 and arm64.
- [x] Verify `go install` in CI so the documented installation path stays working.
- [x] Publish release binaries with checksums from a tag triggered workflow.
