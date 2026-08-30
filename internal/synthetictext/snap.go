package synthetictext

import (
	"fmt"
	"sort"

	graphpolicy "github.com/JuanHuaXu/eventframed/internal/graph"
	"github.com/JuanHuaXu/eventframed/internal/model"
)

const SnapSchemaVersion = "eventframe.public-fact-sheaf-snap.v1"

type SnapCandidate struct {
	ID          string                `json:"id"`
	Description string                `json:"description"`
	Graph       model.PredictiveGraph `json:"graph"`
}

type SnapOracle struct {
	ExpectedCandidateID  string                  `json:"expected_candidate_id"`
	RequiredPartitions   map[string][]string     `json:"required_partitions"`
	ForbiddenCoMembers   [][2]string             `json:"forbidden_co_members"`
	RequiredEdges        [][2]string             `json:"required_edges"`
	ForbiddenEdges       [][2]string             `json:"forbidden_edges"`
	ExpectedClosure      model.DependencyClosure `json:"expected_dependency_closure"`
	ConfirmationCriteria []string                `json:"confirmation_criteria"`
	Falsifiers           []string                `json:"falsifiers"`
}

type SnapCase struct {
	SchemaVersion        string                       `json:"schema_version"`
	ID                   string                       `json:"id"`
	Split                string                       `json:"split"`
	Topic                string                       `json:"topic"`
	TenantID             string                       `json:"tenant_id"`
	CurrentGraph         model.PredictiveGraph        `json:"current_graph"`
	CandidateFamily      []SnapCandidate              `json:"candidate_family"`
	Obligations          []model.ComparisonObligation `json:"comparison_obligations"`
	DesignQueryIDs       []string                     `json:"design_query_ids"`
	ConfirmationQueryIDs []string                     `json:"confirmation_query_ids"`
	Sources              []SourceRef                  `json:"sources"`
	Oracle               SnapOracle                   `json:"oracle"`
}

type SnapManifest struct {
	SchemaVersion       string         `json:"schema_version"`
	Generator           string         `json:"generator"`
	Cases               int            `json:"cases"`
	CandidateGraphs     int            `json:"candidate_graphs"`
	Topics              int            `json:"topics"`
	SplitCases          map[string]int `json:"split_cases"`
	CasesSHA256         string         `json:"cases_sha256"`
	TextCorpusSHA256    string         `json:"text_corpus_sha256"`
	ContainsPrivateText bool           `json:"contains_private_text"`
	ContainsPII         bool           `json:"contains_pii"`
	IdentifierScope     string         `json:"identifier_scope"`
	Grounding           string         `json:"grounding"`
	Scope               string         `json:"scope"`
	Sources             []SourceRef    `json:"sources"`
}

// BuildSnapCases creates known-topology snapping cases over the generated
// public-fact conversations. The facts are sourced; graph defects are injected
// benchmark perturbations and are never represented as source claims.
func BuildSnapCases(records []Record) ([]SnapCase, SnapManifest, error) {
	bySession := make(map[string][]Record)
	for _, record := range records {
		bySession[record.Capture.Turn.SessionID] = append(bySession[record.Capture.Turn.SessionID], record)
	}
	sessions := make([]string, 0, len(bySession))
	for sessionID := range bySession {
		sessions = append(sessions, sessionID)
	}
	sort.Strings(sessions)

	cases := make([]SnapCase, 0, len(sessions))
	splitCases := map[string]int{"design": 0, "confirmation": 0}
	for _, sessionID := range sessions {
		session := bySession[sessionID]
		sort.Slice(session, func(i, j int) bool {
			return session[i].Capture.Turn.Sequence < session[j].Capture.Turn.Sequence
		})
		if len(session) != 12 {
			return nil, SnapManifest{}, fmt.Errorf("session %s has %d turns, want 12", sessionID, len(session))
		}
		caseID := "snap-" + sessionID
		prefix := caseID + ":"
		tenantID := "public-fact-corpus"
		ids := func(indices ...int) []string {
			out := make([]string, 0, len(indices))
			for _, index := range indices {
				out = append(out, session[index].Capture.Turn.ID)
			}
			return out
		}
		node := func(name string, members []string) model.CompatibilityNode {
			return model.CompatibilityNode{
				ID: prefix + name, Kind: "bucket", MemberEventIDs: members,
				PosteriorKeys: []string{"ap:" + prefix + name}, LawSpace: model.RetrievalUsefulnessHorizon,
			}
		}
		edge := func(name, from, to, effect string) model.CompatibilityEdge {
			return model.CompatibilityEdge{
				ID: prefix + name, From: prefix + from, To: prefix + to,
				ComparisonMap: "identity_bernoulli", Effect: effect, Weight: 1,
			}
		}

		current := model.PredictiveGraph{TenantID: tenantID, Version: 1,
			Nodes: []model.CompatibilityNode{
				node("overmerged", ids(0, 2, 3, 4, 8)),
				node("distinction", ids(5)),
				node("distractor", ids(1, 7)),
			},
			Edges: []model.CompatibilityEdge{edge("wrong-overmerged-distractor", "overmerged", "distractor", model.CompatibilityEffectCompatible)},
		}
		correctNodes := []model.CompatibilityNode{
			node("correction", ids(0, 3, 4, 8)),
			node("misconception", ids(2)),
			node("distinction", ids(5)),
			node("distractor", ids(1, 7)),
		}
		correct := model.PredictiveGraph{TenantID: tenantID, Nodes: correctNodes,
			Edges: []model.CompatibilityEdge{
				edge("correction-distinction", "correction", "distinction", model.CompatibilityEffectCompatible),
				edge("correction-supersedes-misconception", "correction", "misconception", model.CompatibilityEffectSupersedes),
			}}
		splitOnly := model.PredictiveGraph{TenantID: tenantID, Nodes: correctNodes,
			Edges: []model.CompatibilityEdge{edge("wrong-correction-distractor", "correction", "distractor", model.CompatibilityEffectCompatible)}}
		harmful := model.PredictiveGraph{TenantID: tenantID,
			Nodes: []model.CompatibilityNode{
				node("everything", ids(0, 2, 3, 4, 5, 8)),
				node("distractor", ids(1, 7)),
			},
			Edges: []model.CompatibilityEdge{edge("distractor-supports-everything", "distractor", "everything", model.CompatibilityEffectSupports)}}
		obligations := []model.ComparisonObligation{{From: prefix + "correction", To: prefix + "distinction", Weight: 1}}
		closure := graphpolicy.DependencyClosure(current, correct, 1)

		caseValue := SnapCase{
			SchemaVersion: SnapSchemaVersion, ID: caseID, Split: session[0].Split,
			Topic: session[0].Topic, TenantID: tenantID, CurrentGraph: current,
			CandidateFamily: []SnapCandidate{
				{ID: "unchanged", Description: "Control retaining the injected overmerge, wrong edge, and missing edge.", Graph: current},
				{ID: "split_only", Description: "Separates the misconception but retains a wrong cross-topic edge and misses the required local relation.", Graph: splitOnly},
				{ID: "split_and_rewire", Description: "Separates correction, misconception, distinction, and distractors; adds the sourced local relation and a directional supersession from verified correction to obsolete claim.", Graph: correct},
				{ID: "harmful_overmerge", Description: "Overmerges local correction, misconception, and distinction, then lets an unrelated distractor support that bucket.", Graph: harmful},
			},
			Obligations: obligations, DesignQueryIDs: ids(4, 6), ConfirmationQueryIDs: ids(9, 11),
			Sources: session[0].Sources,
			Oracle: SnapOracle{
				ExpectedCandidateID: "split_and_rewire",
				RequiredPartitions: map[string][]string{
					"correction": ids(0, 3, 4, 8), "misconception": ids(2),
					"distinction": ids(5), "distractor": ids(1, 7),
				},
				ForbiddenCoMembers: [][2]string{{ids(2)[0], ids(3)[0]}, {ids(2)[0], ids(8)[0]}, {ids(1)[0], ids(3)[0]}},
				RequiredEdges: [][2]string{
					{prefix + "correction", prefix + "distinction"},
					{prefix + "correction", prefix + "misconception"},
				},
				ForbiddenEdges:  [][2]string{{prefix + "correction", prefix + "distractor"}, {prefix + "misconception", prefix + "correction"}},
				ExpectedClosure: closure,
				ConfirmationCriteria: []string{
					"improves or preserves proper retrieval loss on both confirmation queries",
					"retrieves sourced corrections without promoting the recorded misconception",
					"keeps the nearby public-fact distinction available without importing distractor support",
					"reports candidate cost, graph churn, and affected-state closure",
				},
				Falsifiers: []string{
					"split_and_rewire harms confirmation proper loss beyond the predeclared margin",
					"the misconception and correction remain co-members after publication",
					"a deleted wrong edge survives or the required local edge is absent",
					"the implementation invalidates state outside the declared dependency closure",
				},
			},
		}
		cases = append(cases, caseValue)
		splitCases[caseValue.Split]++
	}

	sources := make([]SourceRef, 0, len(factPacks))
	for _, pack := range factPacks {
		sources = append(sources, pack.source)
	}
	return cases, SnapManifest{
		SchemaVersion: SnapSchemaVersion, Generator: "go run ./cmd/eventframe-synthetic-snap",
		Cases: len(cases), CandidateGraphs: len(cases) * 4, Topics: len(factPacks), SplitCases: splitCases,
		ContainsPrivateText: false, ContainsPII: false,
		IdentifierScope: "Deterministic dataset-local labels only. Graph references have no meaning outside this benchmark.",
		Grounding:       "Node membership and expected compatibility are derived from the cited public-fact text corpus. The malformed graph topologies are deliberate benchmark perturbations, not factual assertions.",
		Scope:           "Bounded sheaf-inspired compatibility snapping over retrieval-usefulness-v1 with identity_bernoulli comparison maps and predictive compatible/supports/supersedes effects; not a general sheaf or causal graph benchmark.",
		Sources:         sources,
	}, nil
}
