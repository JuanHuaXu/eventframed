package translationexperiment

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/fuzzing"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/translation"
)

const sourceURL = "https://www.nist.gov/pml/special-publication-811/nist-guide-si-chapter-4-two-classes-si-units-and-si-prefixes"

type Case struct {
	ID       string                        `json:"id"`
	Split    string                        `json:"split"`
	Family   string                        `json:"family"`
	Source   string                        `json:"source"`
	Expected model.ChainTranslationClass   `json:"expected"`
	Request  model.ChainTranslationRequest `json:"request"`
	Chains   translation.Chains            `json:"chains"`
}

type CaseResult struct {
	ID                     string                      `json:"id"`
	Family                 string                      `json:"family"`
	Expected               model.ChainTranslationClass `json:"expected"`
	Observed               model.ChainTranslationClass `json:"observed"`
	Pass                   bool                        `json:"pass"`
	DomainAMovement        float64                     `json:"domain_a_movement"`
	DomainBMovement        float64                     `json:"domain_b_movement"`
	PredictionEffectDefect float64                     `json:"prediction_effect_defect"`
	PredictionEvaluated    bool                        `json:"prediction_evaluated"`
	LocalitySatisfied      bool                        `json:"locality_satisfied"`
	EdgewiseCommuting      bool                        `json:"edgewise_commuting"`
}

type Result struct {
	SchemaVersion  string            `json:"schema_version"`
	Dataset        string            `json:"dataset"`
	Split          string            `json:"split"`
	Source         string            `json:"source"`
	Cases          int               `json:"cases"`
	Passed         int               `json:"passed"`
	ByFamily       map[string][2]int `json:"by_family_passed_total"`
	Results        []CaseResult      `json:"results"`
	Interpretation []string          `json:"interpretation"`
}

func Run(split string) (Result, error) {
	cases, err := Cases(split)
	if err != nil {
		return Result{}, err
	}
	embedder, err := embed.NewHashEmbedder(256)
	if err != nil {
		return Result{}, err
	}
	result := Result{SchemaVersion: "eventframe.chain-translation-results.v1", Dataset: "NIST Celsius-Kelvin public facts", Split: split, Source: sourceURL, ByFamily: make(map[string][2]int)}
	for _, testCase := range cases {
		predictor, err := fuzzing.NewEmbeddingNominationPredictor(context.Background(), embedder, testCase.Request.Query)
		if err != nil {
			return Result{}, err
		}
		response, err := translation.Evaluate(context.Background(), testCase.Request, testCase.Chains, predictor)
		if err != nil {
			return Result{}, fmt.Errorf("case %s: %w", testCase.ID, err)
		}
		item := CaseResult{ID: testCase.ID, Family: testCase.Family, Expected: testCase.Expected, Observed: response.Classification,
			Pass: response.Classification == testCase.Expected, DomainAMovement: response.DomainAMovement,
			DomainBMovement: response.DomainBMovement, PredictionEffectDefect: response.PredictionEffectDefect,
			PredictionEvaluated: response.PredictionEvaluated,
			LocalitySatisfied:   response.LocalitySatisfied, EdgewiseCommuting: response.EdgewiseCommuting}
		result.Results = append(result.Results, item)
		counts := result.ByFamily[testCase.Family]
		if item.Pass {
			result.Passed++
			counts[0]++
		}
		counts[1]++
		result.ByFamily[testCase.Family] = counts
	}
	result.Cases = len(result.Results)
	result.Interpretation = []string{
		"These are deterministic contract controls grounded in NIST's exact Celsius-Kelvin relation, not evidence of autonomous map discovery.",
		"Translation requires mapped change at every stage, unchanged-coordinate locality, terminal agreement, and bounded predictor-effect defect.",
		"Higher-order invariance requires an aligned upstream change, an unchanged terminal stage, and bounded predictor movement in both representations.",
		"The ordinary EventFrame graph remains predictive; no result establishes an intervention or causal claim.",
	}
	return result, nil
}

func Cases(split string) ([]Case, error) {
	queries := map[string][]string{
		"design": {
			"water temperature conversion", "Celsius Kelvin water phase", "freezing boiling temperature", "thermodynamic temperature scale",
			"convert a water temperature", "temperature state transition", "SI temperature relation", "water phase points",
		},
		"confirmation": {
			"compare Celsius with kelvin", "temperature representation and phase", "SI scale mapping", "water at two reference temperatures",
			"aligned temperature measurements", "freezing versus boiling values", "temperature units and outcomes", "thermodynamic measurement chain",
		},
	}
	selected, ok := queries[split]
	if !ok {
		return nil, fmt.Errorf("unknown split %q", split)
	}
	cases := make([]Case, 0, 3*len(selected))
	for i, query := range selected {
		cases = append(cases,
			makeCase(split, i, "translation", query, model.ChainPredictiveTranslation),
			makeCase(split, i, "invariant", query, model.ChainHigherOrderInvariant),
			makeCase(split, i, "divergence-control", query, model.ChainDivergence),
		)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	return cases, nil
}

func makeCase(split string, index int, family, query string, expected model.ChainTranslationClass) Case {
	id := fmt.Sprintf("%s-%s-%02d", split, family, index+1)
	aBefore := []string{"0 degrees Celsius", "Celsius temperature at the water freezing point", "water freezes at about 0 degrees Celsius"}
	aAfter := []string{"100 degrees Celsius", "Celsius temperature at the water boiling point", "water boils at about 100 degrees Celsius"}
	bBefore := []string{"273.15 kelvins", "kelvin temperature at the water freezing point", "water freezes at about 273.15 kelvins"}
	bAfter := []string{"373.15 kelvins", "kelvin temperature at the water boiling point", "water boils at about 373.15 kelvins"}
	field := model.FuzzWhat
	if family == "invariant" {
		aBefore = []string{"Celsius observer A", "water phase reference", "water phase result"}
		aAfter = []string{"Celsius observer B", "water phase reference", "water phase result"}
		bBefore = []string{"kelvin observer A", "water phase reference", "water phase result"}
		bAfter = []string{"kelvin observer B", "water phase reference", "water phase result"}
		field = model.FuzzWho
	}
	chains := makeChains(id, field, aBefore, aAfter, bBefore, bAfter, family == "divergence-control")
	maps := make([]model.ChainStageMap, len(aBefore))
	for stage := range maps {
		maps[stage] = model.ChainStageMap{Stage: stage, DomainAField: field, DomainBField: field,
			DomainABefore: aBefore[stage], DomainAAfter: aAfter[stage], DomainBBefore: bBefore[stage], DomainBAfter: bAfter[stage],
			CorrespondenceID: fmt.Sprintf("nist-c-k-%s-stage-%d", family, stage), ValidityEvidence: sourceURL}
	}
	request := model.ChainTranslationRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "public-chain-corpus", Query: query,
		AsOf: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC), DomainA: chainIDs(chains.DomainABaseline, chains.DomainARevealed),
		DomainB: chainIDs(chains.DomainBBaseline, chains.DomainBRevealed), StageMaps: maps,
		InvariantThreshold: .02, TranslationThreshold: .01}
	return Case{ID: id, Split: split, Family: family, Source: sourceURL, Expected: expected, Request: request, Chains: chains}
}

func makeChains(prefix string, field model.FuzzField, aBefore, aAfter, bBefore, bAfter []string, breakLocality bool) translation.Chains {
	makeOne := func(chain string, values []string) []model.Event {
		result := make([]model.Event, len(values))
		for stage, value := range values {
			event := publicEvent(fmt.Sprintf("%s-%s-%d", prefix, chain, stage), stage)
			setField(&event, field, value)
			result[stage] = event
		}
		return result
	}
	chains := translation.Chains{DomainABaseline: makeOne("a0", aBefore), DomainARevealed: makeOne("a1", aAfter), DomainBBaseline: makeOne("b0", bBefore), DomainBRevealed: makeOne("b1", bAfter)}
	if breakLocality {
		chains.DomainBRevealed[1].Where.Value = "undeclared second coordinate"
	}
	return chains
}

func publicEvent(id string, stage int) model.Event {
	when := time.Date(2026, 8, 1, 12+stage, 0, 0, 0, time.UTC)
	observed := func(value string) model.Field {
		return model.Field{Value: value, Source: model.SourceObserved, Confidence: 1, Evidence: sourceURL}
	}
	return model.Event{ID: id, TenantID: "public-chain-corpus", SessionID: "nist-temperature-chain", Sequence: uint64(stage + 1), Kind: "public_fact",
		Content: "NIST Celsius-Kelvin temperature relation", OccurredAt: when, ObservedAt: when, AvailableAt: when,
		Who: observed("NIST"), What: observed("temperature fact"), Where: observed("SI reference"), When: observed(fmt.Sprintf("stage %d", stage)),
		Why: observed("Celsius temperature is offset from thermodynamic temperature by 273.15 K"), How: observed("T/K = t/degree Celsius + 273.15"),
		Priority: .5, Provenance: model.Provenance{Producer: "eventframe-public-chain-generator"}, Attributes: map[string]string{"source_url": sourceURL}}
}

func setField(event *model.Event, field model.FuzzField, value string) {
	target := &event.What
	switch field {
	case model.FuzzWho:
		target = &event.Who
	case model.FuzzWhere:
		target = &event.Where
	case model.FuzzWhen:
		target = &event.When
	case model.FuzzWhy:
		target = &event.Why
	case model.FuzzHow:
		target = &event.How
	}
	target.Value = value
}

func chainIDs(baseline, revealed []model.Event) model.ChainTrajectory {
	result := model.ChainTrajectory{BaselineEventIDs: make([]string, len(baseline)), RevealedEventIDs: make([]string, len(revealed))}
	for i := range baseline {
		result.BaselineEventIDs[i], result.RevealedEventIDs[i] = baseline[i].ID, revealed[i].ID
	}
	return result
}
