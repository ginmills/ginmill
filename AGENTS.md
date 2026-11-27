# AGENTS.md

Project instructions and workflow orchestration for `ginmill`.

## Workflow Orchestration (OpenSpec-Led Go TDD)

1. **Spec First**: All feature and behavior changes MUST start with an OpenSpec proposal and `tasks.md`. Do not write production code before the spec is approved.
2. **TDD Execution**: During OpenSpec `apply`, strictly follow `test-driven-development` (RED -> GREEN -> REFACTOR) for every task.
3. **Go Standards**:
   - Follow `use-modern-go` aligned with `go.mod`.
   - Apply `golang-testing`, `golang-design-patterns`, `golang-error-handling`, and `golang-concurrency`.
   - Every green verification MUST run `go test -race ./...` (or `make test-race`) — not just a single test file.
4. **Pre-Archive Gate**: Before archiving an OpenSpec change or opening a PR:
   - Run `make check` (`fmt-check`, `vet`, `lint`, `test-coverage`).
   - Run `code-review-and-quality` (5-axes review: Correctness, Readability, Architecture, Security, Performance) against the full `git diff` and resolve all blocking (Critical / High) findings.

## Verification Commands

- `make test` — Run all unit tests.
- `make test-race` — Run full test suite with Go race detector (`-race -count=1`).
- `make test-coverage` — Run tests with `-race` and generate coverage profile + per-function summary.
- `make coverage-check` — Verify total coverage meets `COVERAGE_MIN` (default `80%`).
- `make lint` — Run `golangci-lint` using `.golangci.yml`.
- `make check` — Run the complete pre-archive quality gate (`fmt-check`, `vet`, `lint`, `test-coverage`).
