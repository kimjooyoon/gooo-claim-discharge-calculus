# Operational refutations

This file is append-only operational evidence. A failed operation is retained
as `OPERATIONAL_REFUTED`; it is not silently erased from the project history.

| sequence | event ID | operation | status | retained evidence |
| ---: | --- | --- | --- | --- |
| 1 | `operational-refuted:release-delete-v0.1.0` | deleting the first published `v0.1.0` release while repairing immutable-policy timing | `OPERATIONAL_REFUTED` | `v0.1.0` annotated tag, the original failed release run `33597163687`, and this record are retained; no further deletion is permitted |

The preserved release state is observed separately. The next release version is
`v0.1.1` from the required PR-first path and the same semantic merge lineage.

## Append-only additions

| sequence | event ID | operation | status | retained evidence |
| ---: | --- | --- | --- | --- |
| 2 | `operational-refuted:publish-before-immutable-policy` | publishing the first `v0.1.0` release before the repository immutable-releases policy was enabled | `OPERATIONAL_REFUTED` | release workflow `33597163687` attempt 1 and its failed immutable verification are retained; the later immutable state is verified on `v0.1.0` and `v0.1.1` |
| 3 | `operational-refuted:pr-transition-cardinality` | first PR conformance run asserted 36 transitions while the declared two-evidence case produced 37 | `OPERATIONAL_REFUTED` | PR workflow `33596835876`, job `100141879602`, and the correcting run `33596981492` remain retained |
