# Contributing

Changes are proposed through a pull request. The `.gooo` declaration is the
semantic authority; generated output and Go implementation must remain
derivable from it. Do not add external Go modules or external runtime tools.

The repository-level verification is performed by GitHub Actions on the pull
request. Local executions of test, build, vet, fmt, actionlint, bash checks,
JSON assertions, and conformance are intentionally outside this project’s
runtime authority contract.
