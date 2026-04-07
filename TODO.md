# Repository modernization plan

## Baseline
- [ ] Move the module and CI from Go 1.23.x to Go 1.24.
- [ ] Keep the repository as a single CLI tool; do not replace the current TUI with another terminal UI.
- [ ] Treat `github.com/gizak/termui v2.2.0+incompatible` as removal work, not upgrade work.

## Source code
- [ ] Remove `tui.go`, `-tui`, and `-stdio`; make plain CLI output the default and only interface.
- [ ] Split ELF loading, dependency resolution, and text rendering into small focused packages or files with explicit inputs and outputs.
- [ ] Replace global mutable state (`deps`, `deps_list`, `deps_root`, library path globals) with an application struct and narrow helper functions.
- [ ] Return errors with context instead of exiting deep in helpers; keep `main` responsible for process exit codes and user-facing messages.
- [ ] Add regression tests for library lookup order, symlink resolution, static binaries, duplicate dependencies, and invalid/non-ELF inputs.
- [ ] Run `gofmt`, `go test ./...`, and `go vet ./...` under Go 1.24 in CI.

## Go 1.24 guidance
- [ ] Adopt current Go review guidance: idiomatic names, package-focused APIs, doc comments for exported symbols, early returns, and no ignored errors.
- [ ] Prefer standard library facilities first; keep external dependencies only when they clearly reduce maintenance cost.
- [ ] If repository-local developer tools are added later, track them with Go 1.24 `tool` directives instead of a `tools.go` workaround.
- [ ] Use table-driven tests by default; if benchmarks are added, prefer `testing.B.Loop` in Go 1.24.
- [ ] Keep compatibility simple: avoid unnecessary abstractions, avoid panic for normal control flow, and use wrapped errors that work with `errors.Is` and `errors.As`.

## Dependency and module decisions
- [ ] Remove the direct TUI dependency and its transitive terminal stack (`termbox-go`, `go-runewidth`, `go-wordwrap`, `ansi`, `go-colorable`, `go-isatty`, `panicparse`, `uniseg`) once TUI code is deleted.
- [ ] Keep the standard `flag` package if the tool remains a single-command executable with a small option surface.
- [ ] Evaluate `spf13/cobra` only if shell completion, manpage generation, or subcommands become required.
- [ ] Evaluate `alecthomas/kong` if struct-tag based flag binding is preferred over Cobra's command model.
- [ ] Evaluate `urfave/cli/v3` if a lighter external CLI framework is needed without a TUI stack.
- [ ] Keep `debug/elf` unless new requirements need functionality that the standard library cannot provide.

## Documentation
- [ ] Rewrite `README.md` for CLI-only usage, current repository ownership, and current installation guidance (`go install`, packaged binaries, supported platforms).
- [ ] Remove the screenshot, TUI key bindings, and any references to `github.com/namhyung/elftree`.
- [ ] Document the final CLI flags with stable examples and expected exit behavior.
- [ ] Add contributor-facing notes for local validation commands and supported Go version.

## Suggested rollout order
- [ ] Step 1: upgrade toolchain and CI to Go 1.24.
- [ ] Step 2: refactor core logic away from global state.
- [ ] Step 3: ship CLI-only output and delete TUI code plus terminal dependencies.
- [ ] Step 4: add tests and vet coverage.
- [ ] Step 5: refresh README and release documentation.
