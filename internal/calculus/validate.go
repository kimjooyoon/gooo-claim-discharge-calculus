package calculus

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ValidateSource(source SourceModel) error {
	if source.Language != "gooo" || source.Name != "claim_discharge_calculus" || source.Version != "v1" {
		return fmt.Errorf("unsupported Gooo header")
	}
	if len(source.Entities) != 4 {
		return fmt.Errorf("authority must declare exactly four entities")
	}
	wantEntities := map[string]bool{
		"Claim":                    false,
		"Evidence":                 false,
		"DischargeRule":            false,
		"ActiveFrontierProjection": false,
	}
	for _, entity := range source.Entities {
		if _, ok := wantEntities[entity.Name]; !ok || wantEntities[entity.Name] {
			return fmt.Errorf("invalid or duplicate authority entity %q", entity.Name)
		}
		wantEntities[entity.Name] = true
		if entity.StableID == "" || entity.CausalEdgeID == "" {
			return fmt.Errorf("entity %q must have a stable ID and causal edge ID", entity.Name)
		}
	}
	for name, present := range wantEntities {
		if !present {
			return fmt.Errorf("missing authority entity %q", name)
		}
	}
	if source.Policy.DenominatorID == "" || source.Policy.CellCount != ExpectedDenominator {
		return fmt.Errorf("denominator must be 12 cells")
	}
	if source.Policy.Runtime != (RuntimeAuthority{}) {
		return fmt.Errorf("runtime authority values must all be zero")
	}
	if !equalStrings(source.Policy.Precedence, []string{"REFUTED", "UNKNOWN", "CLOSED"}) {
		return fmt.Errorf("decision precedence must be REFUTED, UNKNOWN, CLOSED")
	}
	if !equalStrings(source.Policy.UnknownFields, expectedUnknownFields) {
		return fmt.Errorf("unknown field contract is not exact")
	}
	if source.Policy.ExplicitFixedPoint != "FIXED_POINT" {
		return fmt.Errorf("only explicit FIXED_POINT may be a fixed point")
	}
	if source.Policy.PublicUtility != "UNKNOWN" {
		return fmt.Errorf("public utility must remain UNKNOWN")
	}
	if err := validateProofBranches(source.ProofBranches); err != nil {
		return err
	}
	if err := validateIndicators(source.Indicators); err != nil {
		return err
	}
	if err := validateActivities(source.Activities); err != nil {
		return err
	}
	if err := validateRules(source.Rules); err != nil {
		return err
	}
	if len(source.Claims) != ExpectedCaseTotal || len(source.Cases) != ExpectedCaseTotal {
		return fmt.Errorf("authority must declare exactly 12 claims and 12 cases")
	}
	if err := validateClaims(source.Claims); err != nil {
		return err
	}
	if err := validateEvidence(source.Evidence, source.Claims); err != nil {
		return err
	}
	if err := validateCases(source.Cases, source.Claims, source.Evidence, source.Rules); err != nil {
		return err
	}
	claimIDs := make([]string, 0, len(source.Claims))
	for _, claim := range source.Claims {
		claimIDs = append(claimIDs, claim.StableID)
	}
	if !equalStrings(source.Projection.InitialClaimIDs, claimIDs) {
		return fmt.Errorf("active frontier projection must start with every claim in declaration order")
	}
	if source.Projection.StableID == "" || source.Projection.CausalEdgeID == "" {
		return fmt.Errorf("active frontier projection needs a stable ID and causal edge ID")
	}
	return nil
}

func validateProofBranches(branches []ProofBranch) error {
	if len(branches) != len(expectedProofBranches) {
		return fmt.Errorf("proof vector must have exactly three branches")
	}
	for index, branch := range branches {
		if branch.Name != expectedProofBranches[index] || branch.Cells != ExpectedBranchCells || branch.CausalEdgeID == "" {
			return fmt.Errorf("proof branch vector is not FOUNDATION/COHERENCE/REGRESSION with four cells")
		}
	}
	return nil
}

func validateIndicators(indicators []Indicator) error {
	if len(indicators) != ExpectedDenominator {
		return fmt.Errorf("indicator vector must have exactly 12 cells")
	}
	counts := make(map[string]int, len(expectedIndicatorClasses))
	seenIDs := make(map[string]bool, len(indicators))
	seenEdges := make(map[string]bool, len(indicators))
	for _, indicator := range indicators {
		validClass := false
		for _, class := range expectedIndicatorClasses {
			if indicator.Class == class {
				validClass = true
				break
			}
		}
		if !validClass || indicator.Cells != 1 || indicator.StableID == "" || indicator.CausalEdgeID == "" {
			return fmt.Errorf("indicator must use one four-cell class with stable identity")
		}
		if seenIDs[indicator.StableID] || seenEdges[indicator.CausalEdgeID] {
			return fmt.Errorf("indicator IDs and causal edge IDs must be unique")
		}
		seenIDs[indicator.StableID] = true
		seenEdges[indicator.CausalEdgeID] = true
		counts[indicator.Class]++
	}
	for _, class := range expectedIndicatorClasses {
		if counts[class] != ExpectedIndicatorSize {
			return fmt.Errorf("indicator class %s must have four cells", class)
		}
	}
	return nil
}

func validateActivities(activities []MetaActivity) error {
	if len(activities) != ExpectedDenominator {
		return fmt.Errorf("there must be exactly 12 meta activities")
	}
	seenIDs := make(map[string]bool, len(activities))
	seenEdges := make(map[string]bool, len(activities))
	for index, activity := range activities {
		if activity.Ordinal != index+1 || activity.StableID == "" || activity.CausalEdgeID == "" || activity.Output == "" {
			return fmt.Errorf("meta activities must have contiguous ordinals and identities")
		}
		if seenIDs[activity.StableID] || seenEdges[activity.CausalEdgeID] {
			return fmt.Errorf("meta activity IDs and causal edge IDs must be unique")
		}
		seenIDs[activity.StableID] = true
		seenEdges[activity.CausalEdgeID] = true
	}
	return nil
}

func validateRules(rules []DischargeRule) error {
	if len(rules) != 4 {
		return fmt.Errorf("there must be exactly four discharge rules")
	}
	want := map[string]bool{
		"EXACT_PROOF_DISCHARGE": false,
		"KNOWN_COUNTEREXAMPLE":  false,
		"UNKNOWN_FAIL_CLOSED":   false,
		"OPERATIONAL_REFUTED":   false,
	}
	for _, rule := range rules {
		if _, ok := want[rule.StableID]; !ok || want[rule.StableID] {
			return fmt.Errorf("invalid or duplicate discharge rule %q", rule.StableID)
		}
		want[rule.StableID] = true
		if !equalStrings(rule.MatchFields, []string{"subject", "predicate", "scope_digest", "contract_digest", "toolchain_digest"}) || rule.CausalEdgeID == "" {
			return fmt.Errorf("rule %q does not bind the exact five-field identity", rule.StableID)
		}
		switch rule.StableID {
		case "EXACT_PROOF_DISCHARGE":
			if rule.EvidenceState != "SUPPORTS" || rule.ProofBranch != "SELECTED" || rule.FrontierAction != "REMOVE" || rule.Decision != "CLOSED" || rule.Resolution != "FIXED_POINT" || rule.Precedence != 3 {
				return fmt.Errorf("exact discharge rule is not explicit FIXED_POINT")
			}
		case "KNOWN_COUNTEREXAMPLE":
			if rule.EvidenceState != "REFUTES" || rule.ProofBranch != "ANY" || rule.FrontierAction != "RETAIN" || rule.Decision != "REFUTED" || rule.Resolution != "REFUTED" || rule.Precedence != 1 {
				return fmt.Errorf("known counterexample rule is not precedence-one REFUTED")
			}
		case "UNKNOWN_FAIL_CLOSED":
			if rule.EvidenceState != "UNKNOWN" || rule.ProofBranch != "ANY" || rule.FrontierAction != "RETAIN" || rule.Decision != "UNKNOWN" || rule.Resolution != "UNKNOWN" || rule.Precedence != 2 {
				return fmt.Errorf("unknown rule is not fail-closed")
			}
		case "OPERATIONAL_REFUTED":
			if rule.EvidenceState != "OPERATIONAL_REFUTED" || rule.ProofBranch != "ANY" || rule.FrontierAction != "RETAIN" || rule.Decision != "REFUTED" || rule.Resolution != "OPERATIONAL_REFUTED" || rule.Precedence != 1 {
				return fmt.Errorf("operational failure rule is not retained as REFUTED")
			}
		}
	}
	for id, present := range want {
		if !present {
			return fmt.Errorf("missing discharge rule %q", id)
		}
	}
	return nil
}

func validateClaims(claims []ClaimDeclaration) error {
	seenIDs := make(map[string]bool, len(claims))
	seenEdges := make(map[string]bool, len(claims))
	for _, claim := range claims {
		if claim.CaseID == "" || claim.StableID == "" || claim.CausalEdgeID == "" || claim.Subject == "" || claim.Predicate == "" || claim.ScopeDigest == "" || claim.ContractDigest == "" || claim.ToolchainDigest == "" {
			return fmt.Errorf("every claim needs a complete identity and causal edge")
		}
		if seenIDs[claim.StableID] || seenEdges[claim.CausalEdgeID] {
			return fmt.Errorf("claim IDs and causal edge IDs must be unique")
		}
		seenIDs[claim.StableID] = true
		seenEdges[claim.CausalEdgeID] = true
	}
	return nil
}

func validateEvidence(evidence []Evidence, claims []ClaimDeclaration) error {
	claimIDs := make(map[string]bool, len(claims))
	for _, claim := range claims {
		claimIDs[claim.StableID] = true
	}
	seenIDs := make(map[string]bool, len(evidence))
	seenEdges := make(map[string]bool, len(evidence))
	if len(evidence) < ExpectedCaseTotal {
		return fmt.Errorf("there must be at least one evidence record per case")
	}
	for _, item := range evidence {
		if item.CaseID == "" || item.StableID == "" || item.CausalEdgeID == "" || !claimIDs[item.ClaimID] {
			return fmt.Errorf("every evidence record needs stable identity, causal edge, and known claim")
		}
		if seenIDs[item.StableID] || seenEdges[item.CausalEdgeID] {
			return fmt.Errorf("evidence IDs and causal edge IDs must be unique")
		}
		seenIDs[item.StableID] = true
		seenEdges[item.CausalEdgeID] = true
	}
	return nil
}

func validateCases(cases []CaseSpec, claims []ClaimDeclaration, evidence []Evidence, rules []DischargeRule) error {
	claimIDs := make(map[string]bool, len(claims))
	for _, claim := range claims {
		claimIDs[claim.StableID] = true
	}
	evidenceIDs := make(map[string]bool, len(evidence))
	for _, item := range evidence {
		evidenceIDs[item.StableID] = true
	}
	ruleIDs := make(map[string]bool, len(rules))
	for _, rule := range rules {
		ruleIDs[rule.StableID] = true
	}
	seenIDs := make(map[string]bool, len(cases))
	counts := make(map[string]int, 3)
	for index, caseSpec := range cases {
		if caseSpec.Ordinal != index+1 || caseSpec.StableID == "" || seenIDs[caseSpec.StableID] || !claimIDs[caseSpec.ClaimID] || !ruleIDs[caseSpec.RuleID] || len(caseSpec.EvidenceIDs) == 0 {
			return fmt.Errorf("each case needs a unique ordinal, claim, rule, and evidence")
		}
		for _, evidenceID := range caseSpec.EvidenceIDs {
			if !evidenceIDs[evidenceID] {
				return fmt.Errorf("case %q references unknown evidence %q", caseSpec.StableID, evidenceID)
			}
		}
		if caseSpec.ExpectedDecision != "CLOSED" && caseSpec.ExpectedDecision != "UNKNOWN" && caseSpec.ExpectedDecision != "REFUTED" {
			return fmt.Errorf("case %q has an unknown expected decision", caseSpec.StableID)
		}
		if caseSpec.ExpectedFrontier != "REMOVED" && caseSpec.ExpectedFrontier != "RETAINED" {
			return fmt.Errorf("case %q has an unknown expected frontier action", caseSpec.StableID)
		}
		seenIDs[caseSpec.StableID] = true
		counts[caseSpec.ExpectedDecision]++
	}
	if counts["CLOSED"] != 4 || counts["UNKNOWN"] != 4 || counts["REFUTED"] != 4 {
		return fmt.Errorf("fixed cases must be four CLOSED, four UNKNOWN, and four REFUTED")
	}
	return nil
}

func ValidateConformance(source SourceModel, report MachineReport) error {
	if err := ValidateSource(source); err != nil {
		return err
	}
	if len(report.Claims) != ExpectedCaseTotal || len(report.Cases) != ExpectedCaseTotal || len(report.Transitions) != expectedTransitionCount(source.Cases) {
		return fmt.Errorf("runtime report cardinality is not 12 claims, 12 cases, and one final transition per case plus every evidence link")
	}
	if report.FixedVector.DenominatorCells != ExpectedDenominator || report.FixedVector.DenominatorID != source.Policy.DenominatorID {
		return fmt.Errorf("runtime denominator does not preserve the fixed 12-cell vector")
	}
	if !equalStrings(report.FixedVector.Precedence, []string{"REFUTED", "UNKNOWN", "CLOSED"}) {
		return fmt.Errorf("runtime precedence is not fail-closed")
	}
	if report.FixedVector.Cases != (DecisionCounts{Closed: 4, Unknown: 4, Refuted: 4}) {
		return fmt.Errorf("runtime case vector is not four/four/four")
	}
	if len(report.ActiveFrontier.InitialClaimIDs) != ExpectedCaseTotal || len(report.ActiveFrontier.RemovedClaimIDs) != 4 || len(report.ActiveFrontier.ActiveClaimIDs) != 8 {
		return fmt.Errorf("frontier projection must remove four and retain eight claims")
	}
	if report.PublicUtility != "UNKNOWN" || report.Runtime.Authority != (RuntimeAuthority{}) {
		return fmt.Errorf("runtime authority or public utility changed")
	}
	for index, expected := range source.Cases {
		actual := report.Cases[index]
		if actual.CaseID != expected.StableID || actual.Decision != expected.ExpectedDecision || actual.Resolution != expected.ExpectedResolution || actual.FrontierAction != expected.ExpectedFrontier {
			return fmt.Errorf("case %q does not match its fixed expected vector", expected.StableID)
		}
		if actual.Decision == "UNKNOWN" {
			if actual.Unknown == nil || !unknownComplete(actual.Unknown) {
				return fmt.Errorf("case %q lost one of the six UNKNOWN fields", expected.StableID)
			}
		} else if actual.Unknown != nil {
			return fmt.Errorf("case %q has UNKNOWN fields despite a terminal non-UNKNOWN decision", expected.StableID)
		}
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal conformance report: %w", err)
	}
	serialized := strings.ToLower(string(encoded))
	if strings.Contains(serialized, "score") || strings.Contains(serialized, "percentage") {
		return fmt.Errorf("aggregate score or percentage is forbidden")
	}
	return nil
}

func unknownComplete(unknown *UnknownState) bool {
	return unknown != nil && unknown.Stage != nil && unknown.Step != nil && unknown.Reason != nil && unknown.UnknownClass != nil && unknown.NextOperation != nil && unknown.BlockedBy != nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func expectedTransitionCount(cases []CaseSpec) int {
	count := 0
	for _, caseSpec := range cases {
		count += 2 + len(caseSpec.EvidenceIDs)
	}
	return count
}
