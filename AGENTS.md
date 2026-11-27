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
   - Run `make check` (`fmt-check`, `vet`, `lint`, `test-coverage`) and `GOWORK=off make check`.
   - Run `code-review-and-quality` (5-axes review: Correctness, Readability, Architecture, Security, Performance) against the full `git diff` and resolve all blocking (Critical / High) findings.
5. **Language Policy**:
   - OpenSpec artifacts (`proposal.md`, `design.md`, `specs/`, `tasks.md`) use Traditional Chinese (正體中文). Code identifiers and concise technical comments stay in English.
6. **Multi-Repo Workspace Discipline**:
   - Local cross-repo development may use the parent `go.work`, but NEVER commit local `replace` directives in `go.mod`.
   - As the lowest-level upstream module (`ginmill` -> `go-kittyd` -> downstream consumers), tag releases first and verify `GOWORK=off make check` before archiving or tagging.

## Architecture Invariants (`ginmill`)

1. **Minimalism & Zero Dependency Bloat**:
   - Depend ONLY on `github.com/gin-gonic/gin` (plus Go stdlib). MUST NOT pull in unrelated third-party libraries or domain-specific abstractions.
2. **Public API Compatibility (SemVer)**:
   - Any breaking change to exported contracts (`Ginmill`, `Server`, `Features`, `NewFeatures`, `With`, `GetRoutes`) MUST explicitly document downstream migration impact (`go-kittyd`, etc.) in the OpenSpec proposal.
3. **Real HTTP Behavior Verification**:
   - Tests for route composition, prefixes, and middleware chains MUST exercise real HTTP requests via `httptest` against `gin.Engine` (verifying status codes, responses, and execution order), not just inspect struct fields.

## Verification Commands

- `make test` — Run all unit tests.
- `make test-race` — Run full test suite with Go race detector (`-race -count=1`).
- `make test-coverage` — Run tests with `-race` and generate coverage profile + per-function summary.
- `make coverage-check` — Verify total coverage meets `COVERAGE_MIN` (default `80%`).
- `make lint` — Run `golangci-lint` using `.golangci.yml`.
- `make check` — Run the complete pre-archive quality gate (`fmt-check`, `vet`, `lint`, `test-coverage`).
- `GOWORK=off make check` — Verify standalone build/test/lint without local `go.work`.
