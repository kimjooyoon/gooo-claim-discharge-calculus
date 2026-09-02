package calculus

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseFile(path string) (SourceModel, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SourceModel{}, fmt.Errorf("read source: %w", err)
	}
	source, err := ParseSource(path, string(raw))
	if err != nil {
		return SourceModel{}, err
	}
	source.RawSourceDigest = DigestBytes(raw)
	return source, nil
}

func ParseSource(path string, text string) (SourceModel, error) {
	var source SourceModel
	scanner := bufio.NewScanner(strings.NewReader(text))
	lineNumber := 0
	headerSeen := false
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if !headerSeen {
			if len(parts) != 3 || parts[0] != "gooo" {
				return source, parseError(path, lineNumber, "first declaration must be `gooo name version`")
			}
			source.Language = parts[0]
			source.Name = parts[1]
			source.Version = parts[2]
			headerSeen = true
			continue
		}
		attributes, err := parseAttributes(parts[1:])
		if err != nil {
			return source, parseError(path, lineNumber, err.Error())
		}
		switch parts[0] {
		case "entity":
			entity, err := parseEntity(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Entities = append(source.Entities, entity)
		case "denominator":
			if err := parsePolicyDenominator(&source.Policy, attributes); err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
		case "runtime":
			if err := parseRuntime(&source.Policy.Runtime, attributes); err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
		case "precedence":
			source.Policy.Precedence = splitList(attributes["values"])
		case "unknown_fields":
			source.Policy.UnknownFields = splitList(attributes["values"])
		case "explicit_fixed_point":
			source.Policy.ExplicitFixedPoint = attributes["value"]
		case "public_utility":
			source.Policy.PublicUtility = attributes["value"]
		case "proof":
			branch, err := parseProofBranch(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.ProofBranches = append(source.ProofBranches, branch)
		case "indicator":
			indicator, err := parseIndicator(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Indicators = append(source.Indicators, indicator)
		case "rule":
			rule, err := parseRule(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Rules = append(source.Rules, rule)
		case "projection":
			projection, err := parseProjection(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Projection = projection
		case "activity":
			activity, err := parseActivity(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Activities = append(source.Activities, activity)
		case "claim":
			claim, err := parseClaimDeclaration(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Claims = append(source.Claims, claim)
		case "evidence":
			evidence, err := parseEvidence(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Evidence = append(source.Evidence, evidence)
		case "case":
			caseSpec, err := parseCase(attributes)
			if err != nil {
				return source, parseError(path, lineNumber, err.Error())
			}
			source.Cases = append(source.Cases, caseSpec)
		default:
			return source, parseError(path, lineNumber, "unknown declaration "+parts[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return source, fmt.Errorf("scan source: %w", err)
	}
	if !headerSeen {
		return source, fmt.Errorf("%s: source is empty", path)
	}
	return source, nil
}

func parseAttributes(parts []string) (map[string]string, error) {
	attributes := make(map[string]string, len(parts))
	for _, part := range parts {
		key, value, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("attribute %q must be key=value", part)
		}
		if _, exists := attributes[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q", key)
		}
		attributes[key] = value
	}
	return attributes, nil
}

func parseEntity(attributes map[string]string) (Entity, error) {
	return Entity{
		Name:         attributes["name"],
		StableID:     attributes["id"],
		CausalEdgeID: attributes["edge"],
	}, requireAttributes(attributes, "name", "id", "edge")
}

func parsePolicyDenominator(policy *Policy, attributes map[string]string) error {
	cells, err := integerAttribute(attributes, "cells")
	if err != nil {
		return err
	}
	policy.DenominatorID = attributes["id"]
	policy.CellCount = cells
	return requireAttributes(attributes, "id", "cells")
}

func parseRuntime(runtime *RuntimeAuthority, attributes map[string]string) error {
	repositoryWrites, err := integerAttribute(attributes, "repository_writes")
	if err != nil {
		return err
	}
	localTests, err := integerAttribute(attributes, "local_test_executions")
	if err != nil {
		return err
	}
	crossProjectGates, err := integerAttribute(attributes, "cross_project_required_gates")
	if err != nil {
		return err
	}
	runtime.RepositoryWrites = repositoryWrites
	runtime.LocalTestExecutions = localTests
	runtime.CrossProjectRequiredGates = crossProjectGates
	return requireAttributes(attributes, "repository_writes", "local_test_executions", "cross_project_required_gates")
}

func parseProofBranch(attributes map[string]string) (ProofBranch, error) {
	cells, err := integerAttribute(attributes, "cells")
	if err != nil {
		return ProofBranch{}, err
	}
	branch := ProofBranch{Name: attributes["branch"], Cells: cells, CausalEdgeID: attributes["edge"]}
	return branch, requireAttributes(attributes, "branch", "cells", "edge")
}

func parseIndicator(attributes map[string]string) (Indicator, error) {
	cells, err := integerAttribute(attributes, "cells")
	if err != nil {
		return Indicator{}, err
	}
	indicator := Indicator{Class: attributes["class"], StableID: attributes["id"], Cells: cells, CausalEdgeID: attributes["edge"]}
	return indicator, requireAttributes(attributes, "class", "id", "cells", "edge")
}

func parseRule(attributes map[string]string) (DischargeRule, error) {
	precedence, err := integerAttribute(attributes, "precedence")
	if err != nil {
		return DischargeRule{}, err
	}
	rule := DischargeRule{
		StableID:       attributes["id"],
		CausalEdgeID:   attributes["edge"],
		MatchFields:    splitList(attributes["matches"]),
		EvidenceState:  attributes["evidence_state"],
		ProofBranch:    attributes["proof_branch"],
		FrontierAction: attributes["frontier_action"],
		Decision:       attributes["decision"],
		Resolution:     attributes["resolution"],
		Precedence:     precedence,
	}
	return rule, requireAttributes(attributes, "id", "edge", "matches", "evidence_state", "proof_branch", "frontier_action", "decision", "resolution", "precedence")
}

func parseProjection(attributes map[string]string) (ActiveFrontierProjection, error) {
	projection := ActiveFrontierProjection{
		StableID:        attributes["id"],
		CausalEdgeID:    attributes["edge"],
		InitialClaimIDs: splitList(attributes["initial_claim_ids"]),
		ActiveClaimIDs:  []string{},
		RemovedClaimIDs: []string{},
		Events:          []FrontierEvent{},
	}
	return projection, requireAttributes(attributes, "id", "edge", "initial_claim_ids")
}

func parseActivity(attributes map[string]string) (MetaActivity, error) {
	ordinal, err := integerAttribute(attributes, "ordinal")
	if err != nil {
		return MetaActivity{}, err
	}
	activity := MetaActivity{
		Ordinal:      ordinal,
		StableID:     attributes["id"],
		CausalEdgeID: attributes["edge"],
		Inputs:       splitList(attributes["inputs"]),
		Output:       attributes["output"],
	}
	return activity, requireAttributes(attributes, "ordinal", "id", "edge", "inputs", "output")
}

func parseClaimDeclaration(attributes map[string]string) (ClaimDeclaration, error) {
	claim := ClaimDeclaration{
		CaseID:          attributes["case"],
		StableID:        attributes["id"],
		CausalEdgeID:    attributes["edge"],
		Subject:         attributes["subject"],
		Predicate:       attributes["predicate"],
		ScopeDigest:     attributes["scope_digest"],
		ContractDigest:  attributes["contract_digest"],
		ToolchainDigest: attributes["toolchain_digest"],
	}
	return claim, requireAttributes(attributes, "case", "id", "edge", "subject", "predicate", "scope_digest", "contract_digest", "toolchain_digest")
}

func parseEvidence(attributes map[string]string) (Evidence, error) {
	evidence := Evidence{
		CaseID:          attributes["case"],
		StableID:        attributes["id"],
		CausalEdgeID:    attributes["edge"],
		ClaimID:         attributes["claim_id"],
		Subject:         optionalAttribute(attributes, "subject"),
		Predicate:       optionalAttribute(attributes, "predicate"),
		ScopeDigest:     optionalAttribute(attributes, "scope_digest"),
		ContractDigest:  optionalAttribute(attributes, "contract_digest"),
		ToolchainDigest: optionalAttribute(attributes, "toolchain_digest"),
		State:           optionalAttribute(attributes, "state"),
		ProofBranch:     optionalAttribute(attributes, "proof_branch"),
		Reason:          optionalAttribute(attributes, "reason"),
	}
	return evidence, requireAttributes(attributes, "case", "id", "edge", "claim_id")
}

func parseCase(attributes map[string]string) (CaseSpec, error) {
	ordinal, err := integerAttribute(attributes, "ordinal")
	if err != nil {
		return CaseSpec{}, err
	}
	caseSpec := CaseSpec{
		Ordinal:             ordinal,
		StableID:            attributes["id"],
		ClaimID:             attributes["claim_id"],
		EvidenceIDs:         splitList(attributes["evidence_ids"]),
		RuleID:              attributes["rule_id"],
		SelectedProofBranch: attributes["selected_branch"],
		ExpectedDecision:    attributes["expected"],
		ExpectedResolution:  attributes["expected_resolution"],
		ExpectedFrontier:    attributes["expected_frontier"],
		Unknown: UnknownState{
			Stage:         optionalAttribute(attributes, "unknown_stage"),
			Step:          optionalAttribute(attributes, "unknown_step"),
			Reason:        optionalAttribute(attributes, "unknown_reason"),
			UnknownClass:  optionalAttribute(attributes, "unknown_class"),
			NextOperation: optionalAttribute(attributes, "next_operation"),
			BlockedBy:     optionalAttribute(attributes, "blocked_by"),
		},
	}
	return caseSpec, requireAttributes(attributes, "ordinal", "id", "claim_id", "evidence_ids", "rule_id", "selected_branch", "expected", "expected_resolution", "expected_frontier", "unknown_stage", "unknown_step", "unknown_reason", "unknown_class", "next_operation", "blocked_by")
}

func requireAttributes(attributes map[string]string, keys ...string) error {
	for _, key := range keys {
		if value, ok := attributes[key]; !ok || value == "" {
			return fmt.Errorf("missing attribute %q", key)
		}
	}
	return nil
}

func integerAttribute(attributes map[string]string, key string) (int, error) {
	value, ok := attributes[key]
	if !ok || value == "" {
		return 0, fmt.Errorf("missing attribute %q", key)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("attribute %q must be an integer: %w", key, err)
	}
	return parsed, nil
}

func optionalAttribute(attributes map[string]string, key string) *string {
	value, ok := attributes[key]
	if !ok || value == "" || value == "null" {
		return nil
	}
	return stringPointer(value)
}

func splitList(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func stringPointer(value string) *string {
	copyValue := value
	return &copyValue
}

func parseError(path string, line int, message string) error {
	return fmt.Errorf("%s:%d: %s", path, line, message)
}
