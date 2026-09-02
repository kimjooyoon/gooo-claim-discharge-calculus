package calculus

func GenerateIR(source SourceModel, sourcePath string) (SemanticIR, error) {
	if err := ValidateSource(source); err != nil {
		return SemanticIR{}, err
	}
	semanticSource := source
	semanticSource.RawSourceDigest = ""
	semanticDigest, err := DigestValue(semanticSource)
	if err != nil {
		return SemanticIR{}, err
	}
	return SemanticIR{
		Schema:         Schema,
		SourcePath:     sourcePath,
		SourceDigest:   source.RawSourceDigest,
		SemanticDigest: semanticDigest,
		Language:       source.Language,
		Name:           source.Name,
		Version:        source.Version,
		Entities:       cloneEntities(source.Entities),
		Policy:         source.Policy,
		ProofBranches:  cloneProofBranches(source.ProofBranches),
		Indicators:     cloneIndicators(source.Indicators),
		Rules:          cloneRules(source.Rules),
		Projection:     cloneProjection(source.Projection),
		Activities:     cloneActivities(source.Activities),
		Claims:         cloneClaimDeclarations(source.Claims),
		Evidence:       cloneEvidence(source.Evidence),
		Cases:          cloneCases(source.Cases),
	}, nil
}

func cloneEntities(values []Entity) []Entity {
	return append([]Entity{}, values...)
}

func cloneProofBranches(values []ProofBranch) []ProofBranch {
	return append([]ProofBranch{}, values...)
}

func cloneIndicators(values []Indicator) []Indicator {
	return append([]Indicator{}, values...)
}

func cloneRules(values []DischargeRule) []DischargeRule {
	result := make([]DischargeRule, len(values))
	for index, value := range values {
		result[index] = value
		result[index].MatchFields = append([]string{}, value.MatchFields...)
	}
	return result
}

func cloneProjection(value ActiveFrontierProjection) ActiveFrontierProjection {
	value.InitialClaimIDs = append([]string{}, value.InitialClaimIDs...)
	value.ActiveClaimIDs = append([]string{}, value.ActiveClaimIDs...)
	value.RemovedClaimIDs = append([]string{}, value.RemovedClaimIDs...)
	value.Events = append([]FrontierEvent{}, value.Events...)
	return value
}

func cloneActivities(values []MetaActivity) []MetaActivity {
	result := make([]MetaActivity, len(values))
	for index, value := range values {
		result[index] = value
		result[index].Inputs = append([]string{}, value.Inputs...)
	}
	return result
}

func cloneClaimDeclarations(values []ClaimDeclaration) []ClaimDeclaration {
	return append([]ClaimDeclaration{}, values...)
}

func cloneEvidence(values []Evidence) []Evidence {
	result := make([]Evidence, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}

func cloneCases(values []CaseSpec) []CaseSpec {
	result := make([]CaseSpec, len(values))
	for index, value := range values {
		result[index] = value
		result[index].EvidenceIDs = append([]string{}, value.EvidenceIDs...)
	}
	return result
}

func fixedVector(source SourceModel) FixedVector {
	proof := make([]VectorCell, 0, len(source.ProofBranches))
	for _, branch := range source.ProofBranches {
		proof = append(proof, VectorCell{Name: branch.Name, Cells: branch.Cells})
	}
	indicators := make([]VectorCell, 0, len(expectedIndicatorClasses))
	for _, class := range expectedIndicatorClasses {
		cells := 0
		for _, indicator := range source.Indicators {
			if indicator.Class == class {
				cells += indicator.Cells
			}
		}
		indicators = append(indicators, VectorCell{Name: class, Cells: cells})
	}
	counts := DecisionCounts{}
	for _, caseSpec := range source.Cases {
		switch caseSpec.ExpectedDecision {
		case "CLOSED":
			counts.Closed++
		case "UNKNOWN":
			counts.Unknown++
		case "REFUTED":
			counts.Refuted++
		}
	}
	return FixedVector{
		DenominatorID:    source.Policy.DenominatorID,
		DenominatorCells: source.Policy.CellCount,
		ProofBranches:    proof,
		Indicators:       indicators,
		Cases:            counts,
		UnknownFields:    append([]string{}, source.Policy.UnknownFields...),
		Precedence:       append([]string{}, source.Policy.Precedence...),
	}
}
