package calculus

const (
	Schema                 = "gooo/claim-discharge-calculus/v1"
	ExpectedDenominator   = 12
	ExpectedCaseTotal     = 12
	ExpectedBranchCells   = 4
	ExpectedIndicatorSize = 4
)

var expectedProofBranches = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}

var expectedIndicatorClasses = []string{"DRIVER", "OUTCOME", "GUARDRAIL"}

var expectedUnknownFields = []string{
	"stage",
	"step",
	"reason",
	"unknown_class",
	"next_operation",
	"blocked_by",
}

type SourceModel struct {
	Language       string                       `json:"language"`
	Name           string                       `json:"name"`
	Version        string                       `json:"version"`
	Entities       []Entity                    `json:"entities"`
	Policy         Policy                       `json:"policy"`
	ProofBranches  []ProofBranch               `json:"proof_branches"`
	Indicators     []Indicator                 `json:"indicators"`
	Rules          []DischargeRule             `json:"rules"`
	Projection     ActiveFrontierProjection    `json:"projection"`
	Activities     []MetaActivity              `json:"activities"`
	Claims         []ClaimDeclaration          `json:"claims"`
	Evidence       []Evidence                  `json:"evidence"`
	Cases          []CaseSpec                  `json:"cases"`
	RawSourceDigest string                     `json:"raw_source_digest,omitempty"`
}

type Entity struct {
	Name         string `json:"name"`
	StableID     string `json:"stable_id"`
	CausalEdgeID string `json:"causal_edge_id"`
}

type Policy struct {
	DenominatorID     string           `json:"denominator_id"`
	CellCount         int              `json:"cell_count"`
	Runtime           RuntimeAuthority `json:"runtime"`
	Precedence        []string         `json:"precedence"`
	UnknownFields     []string         `json:"unknown_fields"`
	ExplicitFixedPoint string          `json:"explicit_fixed_point"`
	PublicUtility     string           `json:"public_utility"`
}

type RuntimeAuthority struct {
	RepositoryWrites          int `json:"repository_writes"`
	LocalTestExecutions       int `json:"local_test_executions"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
}

type ProofBranch struct {
	Name         string `json:"name"`
	Cells        int    `json:"cells"`
	CausalEdgeID string `json:"causal_edge_id"`
}

type Indicator struct {
	Class        string `json:"class"`
	StableID     string `json:"stable_id"`
	Cells        int    `json:"cells"`
	CausalEdgeID string `json:"causal_edge_id"`
}

type DischargeRule struct {
	StableID           string   `json:"stable_id"`
	CausalEdgeID       string   `json:"causal_edge_id"`
	MatchFields        []string `json:"match_fields"`
	EvidenceState      string   `json:"evidence_state"`
	ProofBranch        string   `json:"proof_branch"`
	FrontierAction     string   `json:"frontier_action"`
	Decision           string   `json:"decision"`
	Resolution         string   `json:"resolution"`
	Precedence         int      `json:"precedence"`
}

type ActiveFrontierProjection struct {
	StableID       string           `json:"stable_id"`
	CausalEdgeID   string           `json:"causal_edge_id"`
	InitialClaimIDs []string        `json:"initial_claim_ids"`
	ActiveClaimIDs  []string        `json:"active_claim_ids"`
	RemovedClaimIDs []string        `json:"removed_claim_ids"`
	Events          []FrontierEvent `json:"events"`
}

type FrontierEvent struct {
	Sequence     int    `json:"sequence"`
	ClaimID      string `json:"claim_id"`
	Action       string `json:"action"`
	Decision     string `json:"decision"`
	CausalEdgeID string `json:"causal_edge_id"`
}

type MetaActivity struct {
	Ordinal      int      `json:"ordinal"`
	StableID     string   `json:"stable_id"`
	CausalEdgeID string   `json:"causal_edge_id"`
	Inputs       []string `json:"inputs"`
	Output       string   `json:"output"`
}

type ClaimDeclaration struct {
	CaseID        string `json:"case_id"`
	StableID      string `json:"stable_id"`
	CausalEdgeID  string `json:"causal_edge_id"`
	Subject       string `json:"subject"`
	Predicate     string `json:"predicate"`
	ScopeDigest   string `json:"scope_digest"`
	ContractDigest string `json:"contract_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
}

type Claim struct {
	CaseID          string   `json:"case_id"`
	StableID        string   `json:"stable_id"`
	CausalEdgeID    string   `json:"causal_edge_id"`
	Subject         string   `json:"subject"`
	Predicate       string   `json:"predicate"`
	ScopeDigest     string   `json:"scope_digest"`
	ContractDigest  string   `json:"contract_digest"`
	ToolchainDigest string   `json:"toolchain_digest"`
	Status          string   `json:"status"`
	Resolution      string   `json:"resolution"`
	EvidenceIDs     []string `json:"evidence_ids"`
	AppendOnlyOrder int      `json:"append_only_order"`
}

type Evidence struct {
	CaseID          string  `json:"case_id"`
	StableID        string  `json:"stable_id"`
	CausalEdgeID    string  `json:"causal_edge_id"`
	ClaimID         string  `json:"claim_id"`
	Subject         *string `json:"subject"`
	Predicate       *string `json:"predicate"`
	ScopeDigest     *string `json:"scope_digest"`
	ContractDigest  *string `json:"contract_digest"`
	ToolchainDigest *string `json:"toolchain_digest"`
	State           *string `json:"state"`
	ProofBranch     *string `json:"proof_branch"`
	Reason          *string `json:"reason"`
	AppendOnlyOrder int     `json:"append_only_order"`
}

type CaseSpec struct {
	Ordinal             int         `json:"ordinal"`
	StableID            string      `json:"stable_id"`
	ClaimID             string      `json:"claim_id"`
	EvidenceIDs         []string    `json:"evidence_ids"`
	RuleID              string      `json:"rule_id"`
	SelectedProofBranch string      `json:"selected_proof_branch"`
	ExpectedDecision    string      `json:"expected_decision"`
	ExpectedResolution  string      `json:"expected_resolution"`
	ExpectedFrontier    string      `json:"expected_frontier"`
	Unknown             UnknownState `json:"unknown"`
}

type UnknownState struct {
	Stage        *string `json:"stage"`
	Step         *string `json:"step"`
	Reason       *string `json:"reason"`
	UnknownClass *string `json:"unknown_class"`
	NextOperation *string `json:"next_operation"`
	BlockedBy    *string `json:"blocked_by"`
}

type Transition struct {
	Sequence           int          `json:"sequence"`
	ClaimID            string       `json:"claim_id"`
	EvidenceIDs        []string     `json:"evidence_ids"`
	CausalEdgeID       string       `json:"causal_edge_id"`
	Event              string       `json:"event"`
	Before             string       `json:"before"`
	After              string       `json:"after"`
	Reason             string       `json:"reason"`
	Unknown            *UnknownState `json:"unknown"`
	PreviousDigest     string       `json:"previous_digest"`
	Digest             string       `json:"digest"`
}

type CaseResult struct {
	Ordinal             int          `json:"ordinal"`
	CaseID              string       `json:"case_id"`
	ClaimID             string       `json:"claim_id"`
	EvidenceIDs         []string     `json:"evidence_ids"`
	CausalEdgeID        string       `json:"causal_edge_id"`
	SelectedProofBranch string       `json:"selected_proof_branch"`
	Decision            string       `json:"decision"`
	Resolution          string       `json:"resolution"`
	FrontierAction      string       `json:"frontier_action"`
	Reason              string       `json:"reason"`
	Unknown             *UnknownState `json:"unknown"`
}

type DecisionCounts struct {
	Closed  int `json:"closed"`
	Unknown int `json:"unknown"`
	Refuted int `json:"refuted"`
}

type VectorCell struct {
	Name  string `json:"name"`
	Cells int    `json:"cells"`
}

type FixedVector struct {
	DenominatorID   string           `json:"denominator_id"`
	DenominatorCells int             `json:"denominator_cells"`
	ProofBranches   []VectorCell     `json:"proof_branches"`
	Indicators      []VectorCell     `json:"indicators"`
	Cases           DecisionCounts   `json:"cases"`
	UnknownFields   []string         `json:"unknown_fields"`
	Precedence      []string         `json:"precedence"`
}

type Artifact struct {
	Name       string `json:"name"`
	Bytes      int64  `json:"bytes"`
	SHA256     string `json:"sha256"`
}

type WorkspaceMeasurements struct {
	Files                         int64  `json:"files"`
	Directories                   int64  `json:"directories"`
	GoLines                       int64  `json:"go_lines"`
	GoooLinesExcludingRootREADME  int64  `json:"gooo_lines_excluding_root_readme"`
	GeneratedArtifacts            []Artifact `json:"generated_artifacts"`
	WallMS                        int64  `json:"wall_ms"`
	PeakRSSKiB                    *int64 `json:"peak_rss_kib"`
	Resolution                    string `json:"resolution"`
}

type RuntimeMetadata struct {
	Runner       string              `json:"runner"`
	GoVersion    string              `json:"go_version"`
	Authority    RuntimeAuthority    `json:"authority"`
	Measurements WorkspaceMeasurements `json:"measurements"`
}

type SemanticIR struct {
	Schema        string                    `json:"schema"`
	SourcePath    string                    `json:"source_path"`
	SourceDigest  string                    `json:"source_digest"`
	SemanticDigest string                   `json:"semantic_digest"`
	Language      string                    `json:"language"`
	Name          string                    `json:"name"`
	Version       string                    `json:"version"`
	Entities      []Entity                  `json:"entities"`
	Policy        Policy                    `json:"policy"`
	ProofBranches []ProofBranch              `json:"proof_branches"`
	Indicators    []Indicator                `json:"indicators"`
	Rules         []DischargeRule            `json:"rules"`
	Projection    ActiveFrontierProjection  `json:"projection"`
	Activities    []MetaActivity             `json:"activities"`
	Claims        []ClaimDeclaration         `json:"claims"`
	Evidence      []Evidence                 `json:"evidence"`
	Cases         []CaseSpec                 `json:"cases"`
}

type MachineReport struct {
	Schema          string                       `json:"schema"`
	SourcePath      string                       `json:"source_path"`
	SourceDigest    string                       `json:"source_digest"`
	SemanticDigest  string                       `json:"semantic_digest"`
	SemanticIR      SemanticIR                   `json:"semantic_ir"`
	Claims          []Claim                      `json:"claims"`
	Evidence        []Evidence                  `json:"evidence"`
	DischargeRules  []DischargeRule             `json:"discharge_rules"`
	Transitions     []Transition                `json:"transitions"`
	Cases           []CaseResult                `json:"cases"`
	ActiveFrontier  ActiveFrontierProjection    `json:"active_claim_frontier"`
	FixedVector     FixedVector                 `json:"fixed_vector"`
	Runtime         RuntimeMetadata             `json:"runtime"`
	PublicUtility   string                       `json:"public_utility"`
	ReportDigest    string                       `json:"report_digest"`
}

type RunOptions struct {
	Root       string
	SourcePath string
	OutputDir  string
	Runner     string
}

type RunResult struct {
	Report   MachineReport
	Artifacts []Artifact
}
