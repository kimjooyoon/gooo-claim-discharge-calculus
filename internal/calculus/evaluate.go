package calculus

import (
	"fmt"
	"sort"
)

func Evaluate(source SourceModel, sourcePath string) (MachineReport, error) {
	ir, err := GenerateIR(source, sourcePath)
	if err != nil {
		return MachineReport{}, err
	}
	claims := make([]Claim, 0, len(source.Claims))
	claimIndex := make(map[string]int, len(source.Claims))
	for index, declaration := range source.Claims {
		claimIndex[declaration.StableID] = index
		claims = append(claims, Claim{
			CaseID:          declaration.CaseID,
			StableID:        declaration.StableID,
			CausalEdgeID:    declaration.CausalEdgeID,
			Subject:         declaration.Subject,
			Predicate:       declaration.Predicate,
			ScopeDigest:     declaration.ScopeDigest,
			ContractDigest:  declaration.ContractDigest,
			ToolchainDigest: declaration.ToolchainDigest,
			Status:          "OPEN",
			Resolution:      "OPEN",
			EvidenceIDs:     []string{},
			AppendOnlyOrder: index + 1,
		})
	}
	evidenceByID := make(map[string]Evidence, len(source.Evidence))
	for index, evidence := range source.Evidence {
		evidence.AppendOnlyOrder = index + 1
		evidenceByID[evidence.StableID] = evidence
	}
	rules := make(map[string]DischargeRule, len(source.Rules))
	for _, rule := range source.Rules {
		rules[rule.StableID] = rule
	}
	cases := append([]CaseSpec{}, source.Cases...)
	sort.SliceStable(cases, func(left, right int) bool {
		return cases[left].Ordinal < cases[right].Ordinal
	})

	frontier := cloneProjection(source.Projection)
	frontier.ActiveClaimIDs = append([]string{}, frontier.InitialClaimIDs...)
	frontier.RemovedClaimIDs = []string{}
	frontier.Events = []FrontierEvent{}
	transitions := make([]Transition, 0, len(cases)*3)
	results := make([]CaseResult, 0, len(cases))
	previousDigest := ""
	appendTransition := func(transition Transition) {
		transition.Sequence = len(transitions) + 1
		transition.PreviousDigest = previousDigest
		digest, digestErr := DigestValue(struct {
			Sequence       int           `json:"sequence"`
			ClaimID        string        `json:"claim_id"`
			EvidenceIDs    []string      `json:"evidence_ids"`
			CausalEdgeID   string        `json:"causal_edge_id"`
			Event          string        `json:"event"`
			Before         string        `json:"before"`
			After          string        `json:"after"`
			Reason         string        `json:"reason"`
			Unknown        *UnknownState `json:"unknown"`
			PreviousDigest string        `json:"previous_digest"`
		}{transition.Sequence, transition.ClaimID, transition.EvidenceIDs, transition.CausalEdgeID, transition.Event, transition.Before, transition.After, transition.Reason, transition.Unknown, transition.PreviousDigest})
		if digestErr != nil {
			panic(digestErr)
		}
		transition.Digest = digest
		previousDigest = digest
		transitions = append(transitions, transition)
	}

	for _, caseSpec := range cases {
		claimPosition, ok := claimIndex[caseSpec.ClaimID]
		if !ok {
			return MachineReport{}, fmt.Errorf("case %q references unknown claim", caseSpec.StableID)
		}
		claim := &claims[claimPosition]
		appendTransition(Transition{
			ClaimID:      claim.StableID,
			CausalEdgeID: claim.CausalEdgeID,
			Event:        "CLAIM_DECLARED",
			Before:       "ABSENT",
			After:        "ACTIVE",
			Reason:       "claim-preserved-in-append-only-ledger",
			Unknown:      nil,
		})
		for _, evidenceID := range caseSpec.EvidenceIDs {
			evidence, exists := evidenceByID[evidenceID]
			if !exists {
				return MachineReport{}, fmt.Errorf("case %q references unknown evidence", caseSpec.StableID)
			}
			claim.EvidenceIDs = append(claim.EvidenceIDs, evidenceID)
			appendTransition(Transition{
				ClaimID:      claim.StableID,
				EvidenceIDs:  []string{evidenceID},
				CausalEdgeID: evidence.CausalEdgeID,
				Event:        "EVIDENCE_LINKED",
				Before:       "ACTIVE",
				After:        "ACTIVE",
				Reason:       "evidence-retained-and-linked",
				Unknown:      nil,
			})
		}
		rule, exists := rules[caseSpec.RuleID]
		if !exists {
			return MachineReport{}, fmt.Errorf("case %q references unknown rule", caseSpec.StableID)
		}
		linkedEvidence := make([]Evidence, 0, len(caseSpec.EvidenceIDs))
		for _, evidenceID := range caseSpec.EvidenceIDs {
			linkedEvidence = append(linkedEvidence, evidenceByID[evidenceID])
		}
		result := resolveCase(*claim, linkedEvidence, caseSpec, rule, source)
		claim.Status = result.Decision
		claim.Resolution = result.Resolution
		finalAfter := result.Decision
		finalReason := result.Reason
		if result.Decision == "CLOSED" {
			frontier.ActiveClaimIDs = removeString(frontier.ActiveClaimIDs, claim.StableID)
			frontier.RemovedClaimIDs = append(frontier.RemovedClaimIDs, claim.StableID)
			frontier.Events = append(frontier.Events, FrontierEvent{Sequence: len(frontier.Events) + 1, ClaimID: claim.StableID, Action: "REMOVE", Decision: result.Decision, CausalEdgeID: claim.CausalEdgeID})
		} else {
			finalAfter = "ACTIVE"
			frontier.Events = append(frontier.Events, FrontierEvent{Sequence: len(frontier.Events) + 1, ClaimID: claim.StableID, Action: "RETAIN", Decision: result.Decision, CausalEdgeID: claim.CausalEdgeID})
		}
		appendTransition(Transition{
			ClaimID:      claim.StableID,
			EvidenceIDs:  append([]string{}, caseSpec.EvidenceIDs...),
			CausalEdgeID: claim.CausalEdgeID,
			Event:        finalEvent(result),
			Before:       "ACTIVE",
			After:        finalAfter,
			Reason:       finalReason,
			Unknown:      result.Unknown,
		})
		results = append(results, result)
	}
	counts := DecisionCounts{}
	for _, result := range results {
		switch result.Decision {
		case "CLOSED":
			counts.Closed++
		case "UNKNOWN":
			counts.Unknown++
		case "REFUTED":
			counts.Refuted++
		}
	}
	evidenceRecords := make([]Evidence, 0, len(source.Evidence))
	for _, declaration := range source.Evidence {
		evidenceRecords = append(evidenceRecords, evidenceByID[declaration.StableID])
	}
	return MachineReport{
		Schema:         Schema,
		SourcePath:     sourcePath,
		SourceDigest:   source.RawSourceDigest,
		SemanticDigest: ir.SemanticDigest,
		SemanticIR:     ir,
		Claims:         claims,
		Evidence:       evidenceRecords,
		DischargeRules: cloneRules(source.Rules),
		Transitions:    transitions,
		Cases:          results,
		ActiveFrontier: frontier,
		FixedVector: FixedVector{
			DenominatorID:    source.Policy.DenominatorID,
			DenominatorCells: source.Policy.CellCount,
			ProofBranches:    vectorProofBranches(source.ProofBranches),
			Indicators:       vectorIndicators(source.Indicators),
			Cases:            counts,
			UnknownFields:    append([]string{}, source.Policy.UnknownFields...),
			Precedence:       append([]string{}, source.Policy.Precedence...),
		},
		PublicUtility: source.Policy.PublicUtility,
	}, nil
}

func resolveCase(claim Claim, evidence []Evidence, caseSpec CaseSpec, rule DischargeRule, source SourceModel) CaseResult {
	base := CaseResult{
		Ordinal:             caseSpec.Ordinal,
		CaseID:              caseSpec.StableID,
		ClaimID:             claim.StableID,
		EvidenceIDs:         append([]string{}, caseSpec.EvidenceIDs...),
		CausalEdgeID:        claim.CausalEdgeID,
		SelectedProofBranch: caseSpec.SelectedProofBranch,
		Decision:            "UNKNOWN",
		Resolution:          "UNKNOWN",
		FrontierAction:      "RETAINED",
		Reason:              "no exact discharge observation",
		Unknown:             unknownFromCase(caseSpec),
	}
	if len(evidence) > 0 {
		base.Reason = unknownReason(caseSpec, evidence[0])
	}
	for _, item := range evidence {
		if rule.StableID == "KNOWN_COUNTEREXAMPLE" && rule.EvidenceState == "REFUTES" && exactIdentity(claim, item) && evidenceState(item) == "REFUTES" {
			base.Decision = "REFUTED"
			base.Resolution = "REFUTED"
			base.FrontierAction = "RETAINED"
			base.Reason = dereference(item.Reason, "known-counterexample")
			base.Unknown = nil
			return base
		}
		if rule.StableID == "OPERATIONAL_REFUTED" && rule.EvidenceState == "OPERATIONAL_REFUTED" && exactIdentity(claim, item) && evidenceState(item) == "OPERATIONAL_REFUTED" {
			base.Decision = "REFUTED"
			base.Resolution = "OPERATIONAL_REFUTED"
			base.FrontierAction = "RETAINED"
			base.Reason = dereference(item.Reason, "failed-execution-preserved")
			base.Unknown = nil
			return base
		}
	}
	for _, item := range evidence {
		if !exactIdentity(claim, item) || evidenceState(item) != "SUPPORTS" {
			continue
		}
		if rule.StableID != "EXACT_PROOF_DISCHARGE" || rule.EvidenceState != "SUPPORTS" || rule.ProofBranch != "SELECTED" || rule.Resolution != source.Policy.ExplicitFixedPoint || rule.Decision != "CLOSED" || rule.FrontierAction != "REMOVE" {
			base.Reason = "discharge-rule-is-not-explicit-fixed-point"
			return base
		}
		if !declaredProofBranch(source, caseSpec.SelectedProofBranch) || item.ProofBranch == nil || *item.ProofBranch != caseSpec.SelectedProofBranch {
			base.Reason = "selected-proof-branch-is-unknown-or-unbound"
			return base
		}
		base.Decision = "CLOSED"
		base.Resolution = source.Policy.ExplicitFixedPoint
		base.FrontierAction = "REMOVED"
		base.Reason = "exact-five-field-identity-and-selected-proof-branch"
		base.Unknown = nil
		return base
	}
	for _, item := range evidence {
		if evidenceState(item) == "UNKNOWN" || !completeEvidence(item) || !declaredProofBranch(source, caseSpec.SelectedProofBranch) {
			base.Reason = unknownReason(caseSpec, item)
			break
		}
	}
	return base
}

func exactIdentity(claim Claim, evidence Evidence) bool {
	return dereference(evidence.Subject, "") == claim.Subject && dereference(evidence.Predicate, "") == claim.Predicate && dereference(evidence.ScopeDigest, "") == claim.ScopeDigest && dereference(evidence.ContractDigest, "") == claim.ContractDigest && dereference(evidence.ToolchainDigest, "") == claim.ToolchainDigest && completeIdentity(evidence)
}

func completeIdentity(evidence Evidence) bool {
	return evidence.Subject != nil && evidence.Predicate != nil && evidence.ScopeDigest != nil && evidence.ContractDigest != nil && evidence.ToolchainDigest != nil
}

func completeEvidence(evidence Evidence) bool {
	return completeIdentity(evidence) && evidence.State != nil && evidence.ProofBranch != nil
}

func evidenceState(evidence Evidence) string {
	if evidence.State == nil {
		return "UNKNOWN"
	}
	return *evidence.State
}

func declaredProofBranch(source SourceModel, value string) bool {
	for _, branch := range source.ProofBranches {
		if branch.Name == value {
			return true
		}
	}
	return false
}

func unknownFromCase(caseSpec CaseSpec) *UnknownState {
	unknown := caseSpec.Unknown
	return &unknown
}

func unknownReason(caseSpec CaseSpec, evidence Evidence) string {
	if caseSpec.Unknown.Reason != nil {
		return *caseSpec.Unknown.Reason
	}
	if evidence.Reason != nil {
		return *evidence.Reason
	}
	return "unknown-value-fail-closed"
}

func finalEvent(result CaseResult) string {
	switch result.Decision {
	case "CLOSED":
		return "CLAIM_CLOSED"
	case "REFUTED":
		return "CLAIM_REFUTED"
	default:
		return "CLAIM_UNKNOWN"
	}
}

func vectorProofBranches(branches []ProofBranch) []VectorCell {
	result := make([]VectorCell, 0, len(branches))
	for _, branch := range branches {
		result = append(result, VectorCell{Name: branch.Name, Cells: branch.Cells})
	}
	return result
}

func vectorIndicators(indicators []Indicator) []VectorCell {
	result := make([]VectorCell, 0, len(expectedIndicatorClasses))
	for _, class := range expectedIndicatorClasses {
		cells := 0
		for _, indicator := range indicators {
			if indicator.Class == class {
				cells += indicator.Cells
			}
		}
		result = append(result, VectorCell{Name: class, Cells: cells})
	}
	return result
}

func removeString(values []string, target string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}

func dereference(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}
