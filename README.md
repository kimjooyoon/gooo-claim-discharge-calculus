# gooo-claim-discharge-calculus

An executable Gooo-level semantics for the question: “When evidence appears,
does a claim disappear?”

The authority is the `.gooo` declaration at
[`examples/claim-discharge-calculus.gooo`](examples/claim-discharge-calculus.gooo).
It declares `Claim`, `Evidence`, `DischargeRule`,
`ActiveFrontierProjection`, the exact 12-cell vectors, 12 meta activities, and
12 conformance cases. Go is the parser, generator, evaluator, and CLI.

Claims and evidence are never deleted from the report. A claim leaves the
active frontier only when the selected `DischargeRule` sees exact equality for
`subject`, `predicate`, `scope_digest`, `contract_digest`, and
`toolchain_digest`, plus a declared proof branch. A matching counterexample is
`REFUTED` and has precedence over `UNKNOWN` and `CLOSED`. Only the explicit
`FIXED_POINT` resolution is accepted as `FIXED_POINT`; unknown values fail
closed to `UNKNOWN`.

## Run through the CLI

The output directory must be absolute and empty (or absent). The command emits
machine JSON, a human dossier, a semantic IR, and a run manifest.

```text
go run ./cmd/gooo-claim-discharge-calculus conformance \
  --root "$PWD" \
  --source examples/claim-discharge-calculus.gooo \
  --output-dir /tmp/gooo-claim-discharge-calculus \
  --runner github-actions/ubuntu-latest
```

The repository intentionally does not run Go tests, builds, vet, formatting,
shell assertions, JSON assertions, or conformance locally. The pull-request
workflow is the authoritative verification boundary. The runtime authority
vector remains exactly:

```text
repository_writes=0
local_test_executions=0
cross_project_required_gates=0
```

Public utility evidence is deliberately `UNKNOWN` in v0.1.0.

## Semantics

The fixed conformance vector is four `CLOSED`, four `UNKNOWN`, and four
`REFUTED` cases. Every `UNKNOWN` packet carries `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`. Failed executions remain
as `OPERATIONAL_REFUTED` evidence; they are not erased.

See [`docs/semantics.md`](docs/semantics.md) and
[`docs/conformance.md`](docs/conformance.md).
