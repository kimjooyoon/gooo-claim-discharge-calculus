# Operational refutations

This file is append-only operational evidence. A failed operation is retained
as `OPERATIONAL_REFUTED`; it is not silently erased from the project history.

| sequence | event ID | operation | status | retained evidence |
| ---: | --- | --- | --- | --- |
| 1 | `operational-refuted:release-delete-v0.1.0` | deleting the first published `v0.1.0` release while repairing immutable-policy timing | `OPERATIONAL_REFUTED` | `v0.1.0` annotated tag, the original failed release run `33597163687`, and this record are retained; no further deletion is permitted |

The preserved release state is observed separately. The next release version is
`v0.1.1` from the required PR-first path and the same semantic merge lineage.
