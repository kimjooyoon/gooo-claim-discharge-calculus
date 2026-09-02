# v0.1.0 release contract

The release is created only from a tag whose commit is the checked-out `main`
commit. GitHub Actions runs the fixed conformance command, creates the GitHub
release as a draft, uploads the tarball and its SHA-256 sidecar, publishes the
draft, and verifies the public release and asset identity. The release workflow
does not replace an existing release.

The public utility field remains `UNKNOWN`. Release evidence is operational
evidence for the calculus itself, not a claim of public utility.
