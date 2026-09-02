# Semantics

The source declaration is a line-oriented Gooo surface. It is intentionally
small: declarations have a kind followed by whitespace-separated `key=value`
attributes. Empty values are meaningful missing values and become `null` in
machine output.

The transition ledger is append-only:

```text
CLAIM_DECLARED → EVIDENCE_LINKED → CLOSED|UNKNOWN|REFUTED
```

`CLOSED` removes the claim ID from the active frontier projection. `UNKNOWN`
and `REFUTED` retain it. The claim and every evidence record remain in their
original declaration order. Each record has a stable ID and a causal edge ID;
each transition has a digest chained to its predecessor.

The exact-match tuple is:

```text
(subject, predicate, scope_digest, contract_digest, toolchain_digest)
```

For discharge, the evidence state must be `SUPPORTS`, the selected branch must
be one of the declared proof branches, and the explicit rule resolution must
be `FIXED_POINT`. A state of `REFUTES` or `OPERATIONAL_REFUTED` on an exact
tuple wins before any support or missing-value path. An incomplete or
unrecognized value cannot close a claim.

The report never emits an aggregate score or percentage. It emits exact
integer counts only where the declaration fixes a vector or a conformance
cardinality.
