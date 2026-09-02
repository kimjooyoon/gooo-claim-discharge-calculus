# Conformance

The pull-request workflow runs the only supported verification sequence under
Go 1.27.0:

1. formatting in the verification checkout;
2. build;
3. test;
4. vet; and
5. the executable 12-case conformance command.

The conformance command checks the four fixed proof branches/cells, the three
four-cell indicator classes, the exact 12 cases, decision precedence, the six
UNKNOWN fields, append-only preservation, frontier cardinality, and forbidden
aggregate fields. Its outputs are uploaded as a GitHub Actions artifact.

Main runs produce the same evidence in a separate artifact. The release
workflow is tag-gated, creates a draft release first, uploads a digest-bound
bundle, publishes it, and verifies the published asset identity.
