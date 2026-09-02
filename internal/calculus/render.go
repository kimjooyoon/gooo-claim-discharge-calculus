package calculus

import (
	"fmt"
	"strings"
)

func renderDossier(report MachineReport) string {
	var builder strings.Builder
	builder.WriteString("# Gooo Claim Discharge Calculus dossier\n\n")
	fmt.Fprintf(&builder, "- Schema: `%s`\n- Source digest: `%s`\n- Semantic digest: `%s`\n- Public utility: `%s`\n\n", report.Schema, report.SourceDigest, report.SemanticDigest, report.PublicUtility)
	builder.WriteString("## Meaning\n\n")
	builder.WriteString("A claim is preserved in the append-only ledger. It leaves `active_claim_frontier` only after exact five-field identity equality and a selected declared proof branch produce an explicit `FIXED_POINT` discharge. Exact counterexamples remain in the ledger and resolve to `REFUTED`.\n\n")
	builder.WriteString("## Fixed vectors\n\n")
	fmt.Fprintf(&builder, "- Denominator `%s`: %d cells\n- Decision precedence: `%s`\n- UNKNOWN fields: `%s`\n\n", report.FixedVector.DenominatorID, report.FixedVector.DenominatorCells, strings.Join(report.FixedVector.Precedence, " > "), strings.Join(report.FixedVector.UnknownFields, ", "))
	builder.WriteString("| proof branch | cells |\n| --- | ---: |\n")
	for _, branch := range report.FixedVector.ProofBranches {
		fmt.Fprintf(&builder, "| `%s` | %d |\n", branch.Name, branch.Cells)
	}
	builder.WriteString("\n| indicator class | cells |\n| --- | ---: |\n")
	for _, indicator := range report.FixedVector.Indicators {
		fmt.Fprintf(&builder, "| `%s` | %d |\n", indicator.Name, indicator.Cells)
	}
	fmt.Fprintf(&builder, "\nCase vector: `CLOSED=%d`, `UNKNOWN=%d`, `REFUTED=%d`.\n\n", report.FixedVector.Cases.Closed, report.FixedVector.Cases.Unknown, report.FixedVector.Cases.Refuted)
	builder.WriteString("## Case trace\n\n")
	builder.WriteString("| # | case | claim | evidence | decision | resolution | frontier |\n| ---: | --- | --- | --- | --- | --- | --- |\n")
	for _, item := range report.Cases {
		fmt.Fprintf(&builder, "| %d | `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n", item.Ordinal, item.CaseID, item.ClaimID, strings.Join(item.EvidenceIDs, ", "), item.Decision, item.Resolution, item.FrontierAction)
	}
	builder.WriteString("\n## UNKNOWN packets\n\n")
	for _, item := range report.Cases {
		if item.Decision != "UNKNOWN" {
			continue
		}
		unknown := item.Unknown
		fmt.Fprintf(&builder, "### `%s`\n\n- stage: `%s`\n- step: `%s`\n- reason: `%s`\n- unknown_class: `%s`\n- next_operation: `%s`\n- blocked_by: `%s`\n\n", item.CaseID, displayPointer(unknown.Stage), displayPointer(unknown.Step), displayPointer(unknown.Reason), displayPointer(unknown.UnknownClass), displayPointer(unknown.NextOperation), displayPointer(unknown.BlockedBy))
	}
	builder.WriteString("## Frontier projection\n\n")
	fmt.Fprintf(&builder, "Initial claims: `%s`\n\nRemoved claims: `%s`\n\nActive claims: `%s`\n\n", strings.Join(report.ActiveFrontier.InitialClaimIDs, "`, `"), strings.Join(report.ActiveFrontier.RemovedClaimIDs, "`, `"), strings.Join(report.ActiveFrontier.ActiveClaimIDs, "`, `"))
	builder.WriteString("| sequence | claim | action | decision | causal edge |\n| ---: | --- | --- | --- | --- |\n")
	for _, event := range report.ActiveFrontier.Events {
		fmt.Fprintf(&builder, "| %d | `%s` | `%s` | `%s` | `%s` |\n", event.Sequence, event.ClaimID, event.Action, event.Decision, event.CausalEdgeID)
	}
	builder.WriteString("\n## Append-only ledger\n\n")
	builder.WriteString("| sequence | event | claim | evidence | before | after | digest |\n| ---: | --- | --- | --- | --- | --- | --- |\n")
	for _, transition := range report.Transitions {
		fmt.Fprintf(&builder, "| %d | `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n", transition.Sequence, transition.Event, transition.ClaimID, strings.Join(transition.EvidenceIDs, ", "), transition.Before, transition.After, transition.Digest)
	}
	builder.WriteString("\n## Runtime observations\n\n")
	fmt.Fprintf(&builder, "- runner: `%s`\n- Go version: `%s`\n- files: %d\n- directories: %d\n- Go lines: %d\n- `.gooo` lines excluding root README: %d\n- wall_ms: %d\n- peak_rss_kib: %s\n- measurement resolution: `%s`\n- runtime authority: `repository_writes=%d`, `local_test_executions=%d`, `cross_project_required_gates=%d`\n\n", report.Runtime.Runner, report.Runtime.GoVersion, report.Runtime.Measurements.Files, report.Runtime.Measurements.Directories, report.Runtime.Measurements.GoLines, report.Runtime.Measurements.GoooLinesExcludingRootREADME, report.Runtime.Measurements.WallMS, displayPointer(report.Runtime.Measurements.PeakRSSKiB), report.Runtime.Measurements.Resolution, report.Runtime.Authority.RepositoryWrites, report.Runtime.Authority.LocalTestExecutions, report.Runtime.Authority.CrossProjectRequiredGates)
	builder.WriteString("## Generated artifacts\n\n")
	builder.WriteString("| name | bytes | sha256 |\n| --- | ---: | --- |\n")
	for _, artifact := range report.Runtime.Measurements.GeneratedArtifacts {
		fmt.Fprintf(&builder, "| `%s` | %d | `%s` |\n", artifact.Name, artifact.Bytes, artifact.SHA256)
	}
	builder.WriteString("\nPublic utility remains `UNKNOWN`; no public utility evidence is asserted by this release.\n")
	return builder.String()
}

func displayPointer[T any](value *T) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprint(*value)
}
