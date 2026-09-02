package calculus

import (
	"path/filepath"
	"testing"
)

func TestFixedConformanceVector(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "examples", "claim-discharge-calculus.gooo")
	source, err := ParseFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Evaluate(source, sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateConformance(source, report); err != nil {
		t.Fatal(err)
	}
	if report.Cases[9].Decision != "REFUTED" {
		t.Fatalf("counterexample precedence was not preserved: %s", report.Cases[9].Decision)
	}
}

func TestMissingIdentityIsUnknownAndNull(t *testing.T) {
	evidence := Evidence{
		StableID:     "evidence:test",
		CausalEdgeID: "edge:test",
		ClaimID:      "claim:test",
		State:        stringPointer("SUPPORTS"),
	}
	claim := Claim{StableID: "claim:test", Subject: "service", Predicate: "healthy", ScopeDigest: "scope", ContractDigest: "contract", ToolchainDigest: "toolchain"}
	caseSpec := CaseSpec{StableID: "case:test", ClaimID: claim.StableID, EvidenceIDs: []string{evidence.StableID}, Unknown: UnknownState{Stage: stringPointer("EVIDENCE"), Step: stringPointer("READ"), Reason: stringPointer("missing"), UnknownClass: stringPointer("MISSING_VALUE"), NextOperation: stringPointer("collect"), BlockedBy: stringPointer(evidence.StableID)}}
	rule := DischargeRule{StableID: "UNKNOWN_FAIL_CLOSED"}
	result := resolveCase(claim, []Evidence{evidence}, caseSpec, rule, SourceModel{Policy: Policy{ExplicitFixedPoint: "FIXED_POINT"}})
	if result.Decision != "UNKNOWN" || result.Unknown == nil || evidence.ToolchainDigest != nil {
		t.Fatalf("missing identity did not remain UNKNOWN: %#v", result)
	}
}
