package rerankexperiment

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/packing"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

const (
	SchemaVersion  = "eventframe.synthetic-bidirectional-rerank.v1"
	recallK        = 50
	packK          = 10
	trainingRounds = 16
)

type BlockConfig struct {
	Name                 string
	Seed                 int64
	BidirectionalRepeats int
	RetentionRepeats     int
	EnvelopeRepeats      int
}

type Dataset struct {
	SchemaVersion  string     `json:"schema_version"`
	Block          string     `json:"block"`
	Seed           int64      `json:"seed"`
	RecallK        int        `json:"recall_k"`
	PackK          int        `json:"pack_k"`
	TrainingRounds int        `json:"training_rounds"`
	Cases          []CasePlan `json:"cases"`
}

type CasePlan struct {
	ID             string    `json:"id"`
	Kind           string    `json:"kind"`
	TargetRank     int       `json:"target_rank,omitempty"`
	HarmfulRank    int       `json:"harmful_rank,omitempty"`
	RetainRank     int       `json:"retain_rank,omitempty"`
	WideMargin     bool      `json:"wide_margin"`
	ContractOrder  []string  `json:"contract_order"`
	ContractScores []float64 `json:"contract_scores"`
}

type CaseResult struct {
	ID                      string   `json:"id"`
	Kind                    string   `json:"kind"`
	TargetRank              int      `json:"target_rank,omitempty"`
	HarmfulRank             int      `json:"harmful_rank,omitempty"`
	RetainRank              int      `json:"retain_rank,omitempty"`
	PassPacked              []string `json:"pass_packed"`
	ActivePacked            []string `json:"active_packed"`
	Promoted                bool     `json:"promoted,omitempty"`
	Demoted                 bool     `json:"demoted,omitempty"`
	JointRepair             bool     `json:"joint_repair,omitempty"`
	Retained                bool     `json:"retained,omitempty"`
	EnvelopePromoted        bool     `json:"envelope_promoted,omitempty"`
	PassUseful              int      `json:"pass_useful"`
	ActiveUseful            int      `json:"active_useful"`
	UnnecessaryChurn        int      `json:"unnecessary_churn"`
	ActiveBayesianApplied   int      `json:"active_bayesian_applied"`
	ActiveRecallNanoseconds int64    `json:"active_recall_nanoseconds"`
}

type Rate struct {
	Success int     `json:"success"`
	Total   int     `json:"total"`
	Value   float64 `json:"value"`
	Lower95 float64 `json:"lower_95"`
	Upper95 float64 `json:"upper_95"`
}

type RankStratum struct {
	Promotion Rate `json:"promotion"`
	Demotion  Rate `json:"demotion"`
}

type Report struct {
	SchemaVersion              string              `json:"schema_version"`
	Block                      string              `json:"block"`
	Cases                      int                 `json:"cases"`
	Promotion                  Rate                `json:"promotion"`
	Demotion                   Rate                `json:"demotion"`
	JointRepair                Rate                `json:"joint_repair"`
	Retention                  Rate                `json:"retention"`
	EnvelopePromotion          Rate                `json:"envelope_promotion"`
	PassPacketPrecision        float64             `json:"pass_packet_precision"`
	ActivePacketPrecision      float64             `json:"active_packet_precision"`
	MeanUnnecessaryChurn       float64             `json:"mean_unnecessary_churn"`
	ActiveRecallP50Nanoseconds int64               `json:"active_recall_p50_nanoseconds"`
	ActiveRecallP95Nanoseconds int64               `json:"active_recall_p95_nanoseconds"`
	ActiveRecallP99Nanoseconds int64               `json:"active_recall_p99_nanoseconds"`
	ActiveRecallMaxNanoseconds int64               `json:"active_recall_max_nanoseconds"`
	ByTargetRank               map[int]RankStratum `json:"by_target_rank"`
	ByHarmfulRank              map[int]RankStratum `json:"by_harmful_rank"`
	Criteria                   map[string]bool     `json:"criteria"`
	OverallPassed              bool                `json:"overall_passed"`
	Results                    []CaseResult        `json:"results"`
}

type SuiteDataset struct {
	SchemaVersion string  `json:"schema_version"`
	Design        Dataset `json:"design"`
	Confirmation  Dataset `json:"confirmation"`
}

type SuiteReport struct {
	SchemaVersion string `json:"schema_version"`
	Design        Report `json:"design"`
	Confirmation  Report `json:"confirmation"`
}

func Generate(config BlockConfig) (Dataset, error) {
	if config.Name != "design" && config.Name != "confirmation" {
		return Dataset{}, errors.New("block must be design or confirmation")
	}
	if config.BidirectionalRepeats < 0 || config.RetentionRepeats < 0 || config.EnvelopeRepeats < 0 {
		return Dataset{}, errors.New("case repetitions cannot be negative")
	}
	rng := rand.New(rand.NewSource(config.Seed))
	dataset := Dataset{SchemaVersion: SchemaVersion, Block: config.Name, Seed: config.Seed, RecallK: recallK, PackK: packK, TrainingRounds: trainingRounds}
	sequence := 0
	for repeat := 0; repeat < config.BidirectionalRepeats; repeat++ {
		for _, targetRank := range []int{11, 15, 25, 40, 50} {
			for _, harmfulRank := range []int{1, 3, 5, 7, 10} {
				sequence++
				order := controlledOrder(rng, "bidirectional", targetRank, harmfulRank, 0)
				dataset.Cases = append(dataset.Cases, CasePlan{ID: fmt.Sprintf("%s-bi-%04d", config.Name, sequence), Kind: "bidirectional", TargetRank: targetRank, HarmfulRank: harmfulRank, ContractOrder: order, ContractScores: contractScores(false)})
			}
		}
	}
	for repeat := 0; repeat < config.RetentionRepeats; repeat++ {
		for _, retainRank := range []int{1, 5, 10} {
			sequence++
			order := controlledOrder(rng, "retention", 0, 0, retainRank)
			dataset.Cases = append(dataset.Cases, CasePlan{ID: fmt.Sprintf("%s-retain-%04d", config.Name, sequence), Kind: "retention", RetainRank: retainRank, ContractOrder: order, ContractScores: contractScores(false)})
		}
	}
	for repeat := 0; repeat < config.EnvelopeRepeats; repeat++ {
		sequence++
		order := controlledOrder(rng, "envelope", 50, 0, 0)
		dataset.Cases = append(dataset.Cases, CasePlan{ID: fmt.Sprintf("%s-envelope-%04d", config.Name, sequence), Kind: "envelope", TargetRank: 50, WideMargin: true, ContractOrder: order, ContractScores: contractScores(true)})
	}
	return dataset, nil
}

func Run(ctx context.Context, dataset Dataset) (Report, error) {
	if dataset.SchemaVersion != SchemaVersion || dataset.RecallK != recallK || dataset.PackK != packK || dataset.TrainingRounds != trainingRounds || len(dataset.Cases) == 0 {
		return Report{}, errors.New("dataset does not match the frozen reranking protocol")
	}
	base := time.Now().UTC().Add(-48 * time.Hour)
	runtimes := make(map[string]*runtimePair)
	for _, kind := range []string{"bidirectional", "retention", "envelope"} {
		pair, err := newRuntimePair(ctx, kind, base)
		if err != nil {
			return Report{}, err
		}
		runtimes[kind] = pair
		if err := pair.train(ctx, base); err != nil {
			return Report{}, fmt.Errorf("train %s: %w", kind, err)
		}
	}
	results := make([]CaseResult, 0, len(dataset.Cases))
	for index, plan := range dataset.Cases {
		pair := runtimes[plan.Kind]
		if pair == nil {
			return Report{}, fmt.Errorf("case %q has unknown kind %q", plan.ID, plan.Kind)
		}
		pair.ranker.put(plan.ID, plan.ContractOrder, plan.ContractScores)
		asOf := base.Add(8*time.Hour + time.Duration(index)*time.Minute)
		request := recallRequest(pair.tenant, plan.ID, asOf)
		passPacket, err := pair.pass.Recall(ctx, request)
		if err != nil {
			return Report{}, fmt.Errorf("pass recall %s: %w", plan.ID, err)
		}
		started := time.Now()
		activePacket, err := pair.active.Recall(ctx, request)
		duration := time.Since(started)
		if err != nil {
			return Report{}, fmt.Errorf("active recall %s: %w", plan.ID, err)
		}
		result := scoreCase(plan, passPacket, activePacket)
		result.ActiveRecallNanoseconds = duration.Nanoseconds()
		results = append(results, result)
	}
	return summarize(dataset.Block, results), nil
}

type rankPlan struct {
	order  []string
	scores []float64
}

type controlledRanker struct {
	plans map[string]rankPlan
}

func newControlledRanker() *controlledRanker {
	return &controlledRanker{plans: make(map[string]rankPlan)}
}

func (r *controlledRanker) put(query string, order []string, scores []float64) {
	r.plans[query] = rankPlan{order: append([]string(nil), order...), scores: append([]float64(nil), scores...)}
}

func (r *controlledRanker) RankCandidates(_ context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	plan, ok := r.plans[request.QueryText]
	if !ok {
		return nil, fmt.Errorf("no synthetic rank plan for %q", request.QueryText)
	}
	byID := make(map[string]retrieval.Candidate, len(request.Candidates))
	for _, candidate := range request.Candidates {
		byID[candidate.ID] = candidate
	}
	limit := min(request.K2, len(plan.order))
	result := make([]retrieval.Candidate, 0, limit)
	for index := 0; index < limit; index++ {
		candidate, found := byID[plan.order[index]]
		if !found {
			return nil, fmt.Errorf("rank plan references absent candidate %q", plan.order[index])
		}
		candidate.Score = plan.scores[index]
		result = append(result, candidate)
	}
	return result, nil
}

func (*controlledRanker) ContractName() string { return "synthetic-controlled-libravdb-order-v1" }

type runtimePair struct {
	tenant string
	ranker *controlledRanker
	pass   *service.Service
	active *service.Service
	kind   string
}

func newRuntimePair(ctx context.Context, kind string, base time.Time) (*runtimePair, error) {
	ranker := newControlledRanker()
	tenant := "synthetic-rerank-" + kind
	passStore, activeStore := memorystore.New(), memorystore.New()
	embedder, err := embed.NewHashEmbedder(2)
	if err != nil {
		return nil, err
	}
	pass, err := service.New(passStore, embedder, service.Config{DefaultRecallK: recallK, DefaultPackK: packK, DefaultTokenBudget: 100_000, OverfetchMultiplier: 3, CandidateRanker: ranker, ResidualMode: service.ResidualModeDisabled, RankingPolicy: ranking.DefaultPolicy(), PackingPolicy: packing.Policy{MaxPack: packK}})
	if err != nil {
		return nil, err
	}
	activePolicy := ranking.DefaultPolicy()
	activePolicy.ContextualEnabled, activePolicy.HierarchicalEnabled = true, true
	active, err := service.New(activeStore, embedder, service.Config{
		DefaultRecallK: recallK, DefaultPackK: packK, DefaultTokenBudget: 100_000, OverfetchMultiplier: 3,
		CandidateRanker: ranker, ResidualMode: service.ResidualModeDisabled, RankingPolicy: activePolicy,
		PackingPolicy:        packing.Policy{MaxPack: packK},
		BayesianPolicy:       bayes.Policy{VectorWeight: .6, NeighborWeight: .15, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: .65, CriticalThreshold: .45, AuditProbability: 1, MaxActive: recallK, AuditSeed: "synthetic-bidirectional-rerank-v1", CheapUpdateAll: true},
		BayesianChangePolicy: bayes.ChangePolicy{Hazard: .000001, Threshold: 1, MaxRun: 128},
		ResidualPolicy:       residual.Policy{Clip: .15, MinSupport: 1e9, MinConfidence: 1, ConfidenceDelta: .05, MotionLimit: 0, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001},
	})
	if err != nil {
		return nil, err
	}
	for index, id := range candidateIDs(kind) {
		event := testutil.Event(id, candidateContent(kind, id), base.Add(time.Duration(index)*time.Second))
		event.TenantID, event.SessionID = tenant, "evidence"
		event.Embedding, event.EmbeddingModel = []float32{0, 1}, embedder.ModelKey()
		for _, runtime := range []*service.Service{pass, active} {
			if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
				return nil, err
			}
		}
	}
	snapshot := activeStore.Snapshot(ctx)
	now := time.Now().UTC()
	selection := model.SelectionSupportCertificate{ID: "selection-" + kind, TenantID: tenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: 1, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic frontier", Issuer: "synthetic-rerank-experiment", ExternalAudit: true, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(24 * time.Hour)}
	if _, err := active.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		return nil, err
	}
	omitted := model.OmittedInfluenceCertificate{ID: "omitted-" + kind, TenantID: tenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: 0, DivergenceLimit: .05, AuditProbability: 1, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic frontier", Issuer: "synthetic-rerank-experiment", ExternalAudit: true, ValidUntil: now.Add(24 * time.Hour)}
	if _, err := active.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted}); err != nil {
		return nil, err
	}
	return &runtimePair{tenant: tenant, ranker: ranker, pass: pass, active: active, kind: kind}, nil
}

func (p *runtimePair) train(ctx context.Context, base time.Time) error {
	order := candidateIDs(p.kind)
	scores := contractScores(false)
	for round := 0; round < trainingRounds; round++ {
		query := fmt.Sprintf("train-%s-%02d", p.kind, round)
		p.ranker.put(query, order, scores)
		asOf := base.Add(2*time.Hour + time.Duration(round)*time.Minute)
		packet, err := p.active.Recall(ctx, recallRequest(p.tenant, query, asOf))
		if err != nil {
			return err
		}
		labels := trainingLabels(p.kind)
		for eventID, useful := range labels {
			_, err := p.active.ObserveBayesianOutcome(ctx, model.BayesianOutcomeRequest{
				ProtocolVersion: model.ProtocolVersion, IdempotencyKey: fmt.Sprintf("%s-%02d-%s", p.kind, round, eventID),
				TenantID: p.tenant, JournalID: packet.BayesianShadow.JournalID, EventID: eventID, Useful: useful,
				ObservedAt: asOf.Add(time.Second), AvailableAt: asOf.Add(time.Second), Source: model.OutcomeFullStream, InclusionProbability: 1,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func recallRequest(tenant, query string, asOf time.Time) model.RecallRequest {
	return model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: tenant, SessionID: "evaluation", Query: query, Embedding: []float32{1, 0}, EmbeddingModel: "feature-hash-v1:d2", AsOf: asOf, RecallK: recallK, PackK: packK, TokenBudget: 100_000}
}

func controlledOrder(rng *rand.Rand, kind string, targetRank, harmfulRank, retainRank int) []string {
	ids := candidateIDs(kind)
	special := make(map[string]int)
	switch kind {
	case "bidirectional":
		special["target"] = targetRank
		special["harmful"] = harmfulRank
	case "retention":
		special["suspicious-useful"] = retainRank
	case "envelope":
		special["target"] = targetRank
	}
	fillers := make([]string, 0, len(ids)-len(special))
	for _, id := range ids {
		if _, ok := special[id]; !ok {
			fillers = append(fillers, id)
		}
	}
	rng.Shuffle(len(fillers), func(i, j int) { fillers[i], fillers[j] = fillers[j], fillers[i] })
	order := make([]string, len(ids))
	for id, rank := range special {
		order[rank-1] = id
	}
	next := 0
	for index := range order {
		if order[index] == "" {
			order[index], next = fillers[next], next+1
		}
	}
	return order
}

func candidateIDs(kind string) []string {
	ids := make([]string, 0, recallK)
	switch kind {
	case "bidirectional":
		ids = append(ids, "target", "harmful")
	case "retention":
		ids = append(ids, "suspicious-useful")
	case "envelope":
		ids = append(ids, "target")
	}
	for len(ids) < recallK {
		ids = append(ids, fmt.Sprintf("filler-%02d", len(ids)))
	}
	return ids
}

func candidateContent(kind, id string) string {
	switch id {
	case "target":
		return "current verified answer for the synthetic query"
	case "harmful":
		return "stale superseded answer known to be harmful"
	case "suspicious-useful":
		return "old oddly phrased record that remains verified and useful"
	default:
		return fmt.Sprintf("semantically similar distractor %s for %s", id, kind)
	}
}

func trainingLabels(kind string) map[string]bool {
	switch kind {
	case "bidirectional":
		return map[string]bool{"target": true, "harmful": false}
	case "retention":
		return map[string]bool{"suspicious-useful": true}
	case "envelope":
		return map[string]bool{"target": true}
	default:
		return nil
	}
}

func usefulSet(kind string) map[string]struct{} {
	result := make(map[string]struct{})
	for id, useful := range trainingLabels(kind) {
		if useful {
			result[id] = struct{}{}
		}
	}
	return result
}

func contractScores(wide bool) []float64 {
	scores := make([]float64, recallK)
	for index := range scores {
		if wide {
			scores[index] = .90 - float64(index)*.005
		} else {
			scores[index] = .70 - float64(index)*.0003
		}
	}
	return scores
}

func scoreCase(plan CasePlan, pass, active model.ContextPacket) CaseResult {
	passIDs, activeIDs := packetIDs(pass), packetIDs(active)
	passSet, activeSet := stringSet(passIDs), stringSet(activeIDs)
	result := CaseResult{ID: plan.ID, Kind: plan.Kind, TargetRank: plan.TargetRank, HarmfulRank: plan.HarmfulRank, RetainRank: plan.RetainRank, PassPacked: passIDs, ActivePacked: activeIDs}
	switch plan.Kind {
	case "bidirectional":
		result.Promoted = !contains(passSet, "target") && contains(activeSet, "target")
		result.Demoted = contains(passSet, "harmful") && !contains(activeSet, "harmful")
		result.JointRepair = result.Promoted && result.Demoted
		result.UnnecessaryChurn = membershipChurn(passSet, activeSet, map[string]struct{}{"target": {}, "harmful": {}})
	case "retention":
		result.Retained = contains(passSet, "suspicious-useful") && contains(activeSet, "suspicious-useful")
		result.UnnecessaryChurn = membershipChurn(passSet, activeSet, map[string]struct{}{"suspicious-useful": {}})
	case "envelope":
		result.EnvelopePromoted = !contains(passSet, "target") && contains(activeSet, "target")
		result.UnnecessaryChurn = membershipChurn(passSet, activeSet, map[string]struct{}{"target": {}})
	}
	useful := usefulSet(plan.Kind)
	for _, id := range passIDs {
		if _, ok := useful[id]; ok {
			result.PassUseful++
		}
	}
	for _, candidate := range active.Candidates {
		if _, ok := useful[candidate.Event.ID]; ok {
			result.ActiveUseful++
		}
		if candidate.BayesianApplied {
			result.ActiveBayesianApplied++
		}
	}
	return result
}

func summarize(block string, results []CaseResult) Report {
	report := Report{SchemaVersion: SchemaVersion, Block: block, Cases: len(results), Results: results, ByTargetRank: make(map[int]RankStratum), ByHarmfulRank: make(map[int]RankStratum), Criteria: make(map[string]bool)}
	var promotion, demotion, joint, retention, envelope []bool
	var passUseful, activeUseful, packedTotal, churn int
	latencies := make([]int64, 0, len(results))
	byTarget, byHarm := make(map[int][]CaseResult), make(map[int][]CaseResult)
	for _, result := range results {
		switch result.Kind {
		case "bidirectional":
			promotion, demotion, joint = append(promotion, result.Promoted), append(demotion, result.Demoted), append(joint, result.JointRepair)
			byTarget[result.TargetRank] = append(byTarget[result.TargetRank], result)
			byHarm[result.HarmfulRank] = append(byHarm[result.HarmfulRank], result)
		case "retention":
			retention = append(retention, result.Retained)
		case "envelope":
			envelope = append(envelope, result.EnvelopePromoted)
		}
		passUseful += result.PassUseful
		activeUseful += result.ActiveUseful
		packedTotal += len(result.PassPacked)
		churn += result.UnnecessaryChurn
		latencies = append(latencies, result.ActiveRecallNanoseconds)
	}
	report.Promotion, report.Demotion, report.JointRepair = rate(promotion), rate(demotion), rate(joint)
	report.Retention, report.EnvelopePromotion = rate(retention), rate(envelope)
	if packedTotal > 0 {
		report.PassPacketPrecision = float64(passUseful) / float64(packedTotal)
		report.ActivePacketPrecision = float64(activeUseful) / float64(packedTotal)
	}
	if len(results) > 0 {
		report.MeanUnnecessaryChurn = float64(churn) / float64(len(results))
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	report.ActiveRecallP50Nanoseconds = percentile(latencies, .50)
	report.ActiveRecallP95Nanoseconds = percentile(latencies, .95)
	report.ActiveRecallP99Nanoseconds = percentile(latencies, .99)
	if len(latencies) > 0 {
		report.ActiveRecallMaxNanoseconds = latencies[len(latencies)-1]
	}
	for rank, cases := range byTarget {
		promoted, demoted := make([]bool, 0, len(cases)), make([]bool, 0, len(cases))
		for _, item := range cases {
			promoted, demoted = append(promoted, item.Promoted), append(demoted, item.Demoted)
		}
		report.ByTargetRank[rank] = RankStratum{Promotion: rate(promoted), Demotion: rate(demoted)}
	}
	for rank, cases := range byHarm {
		promoted, demoted := make([]bool, 0, len(cases)), make([]bool, 0, len(cases))
		for _, item := range cases {
			promoted, demoted = append(promoted, item.Promoted), append(demoted, item.Demoted)
		}
		report.ByHarmfulRank[rank] = RankStratum{Promotion: rate(promoted), Demotion: rate(demoted)}
	}
	report.Criteria["promotion"] = report.Promotion.Value >= .80 && report.Promotion.Lower95 >= .70
	report.Criteria["demotion"] = report.Demotion.Value >= .80 && report.Demotion.Lower95 >= .70
	report.Criteria["joint_repair"] = report.JointRepair.Value >= .75
	report.Criteria["retention"] = report.Retention.Value >= .99 && report.Retention.Lower95 >= .95
	report.Criteria["precision"] = report.ActivePacketPrecision > report.PassPacketPrecision
	report.Criteria["unnecessary_churn"] = report.MeanUnnecessaryChurn <= .25
	report.Criteria["latency"] = report.ActiveRecallP99Nanoseconds < int64(100*time.Millisecond)
	report.Criteria["bounded_envelope"] = report.EnvelopePromotion.Value <= .05
	report.OverallPassed = true
	for _, passed := range report.Criteria {
		report.OverallPassed = report.OverallPassed && passed
	}
	return report
}

func rate(values []bool) Rate {
	success := 0
	for _, value := range values {
		if value {
			success++
		}
	}
	result := Rate{Success: success, Total: len(values)}
	if len(values) == 0 {
		return result
	}
	result.Value = float64(success) / float64(len(values))
	result.Lower95, result.Upper95 = wilson(success, len(values), 1.959963984540054)
	return result
}

func wilson(success, total int, z float64) (float64, float64) {
	if total == 0 {
		return 0, 0
	}
	n := float64(total)
	p := float64(success) / n
	z2 := z * z
	center := (p + z2/(2*n)) / (1 + z2/n)
	half := z * math.Sqrt(p*(1-p)/n+z2/(4*n*n)) / (1 + z2/n)
	return math.Max(0, center-half), math.Min(1, center+half)
}

func percentile(sorted []int64, quantile float64) int64 {
	if len(sorted) == 0 {
		return 0
	}
	index := int(math.Ceil(float64(len(sorted))*quantile)) - 1
	index = max(0, min(index, len(sorted)-1))
	return sorted[index]
}

func packetIDs(packet model.ContextPacket) []string {
	ids := make([]string, 0, len(packet.Candidates))
	for _, candidate := range packet.Candidates {
		ids = append(ids, candidate.Event.ID)
	}
	return ids
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func contains(set map[string]struct{}, value string) bool { _, ok := set[value]; return ok }

func membershipChurn(left, right, excluded map[string]struct{}) int {
	count := 0
	for id := range left {
		if _, skip := excluded[id]; skip {
			continue
		}
		if _, present := right[id]; !present {
			count++
		}
	}
	for id := range right {
		if _, skip := excluded[id]; skip {
			continue
		}
		if _, present := left[id]; !present {
			count++
		}
	}
	return count
}
