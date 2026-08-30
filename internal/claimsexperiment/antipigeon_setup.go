package claimsexperiment

import (
	"context"
	"fmt"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

type antiPigeonRunner struct {
	name    string
	runtime *service.Service
}

var antiPigeonIDs = []string{"positive-a", "positive-b", "negative-a", "negative-b"}

func newAntiPigeonRunner(ctx context.Context, name string) (antiPigeonRunner, error) {
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	groupPolicy := bayes.GroupPolicy{PriorSplit: .5, DecisionThreshold: .95, MinMemberSupport: 8, MaxMembers: 64, EquivalenceMargin: .15, EquivalenceThreshold: .80, MaxUncertainBorrowing: .10}
	groupPolicy.DisableAutoRevision = name == "naive_shared"
	runtime, err := service.New(memory, embedder, service.Config{
		DefaultRecallK: 4, DefaultPackK: 4, DefaultTokenBudget: 400, BayesianScoreWeight: .25,
		BayesianPolicy:       bayes.Policy{VectorWeight: 1, Threshold: 0, CriticalThreshold: 0, AuditProbability: 1, MaxActive: 4, AuditSeed: name},
		BayesianChangePolicy: bayes.ChangePolicy{Hazard: .001, Threshold: 1, MaxRun: 64},
		BayesianGroupPolicy:  groupPolicy,
		ResidualPolicy:       residual.Policy{Clip: .15, MinSupport: 1e9, MinConfidence: 1, ConfidenceDelta: .5, MotionLimit: 0, MaxAge: 24 * time.Hour, ImprovementDelta: .001},
	})
	if err != nil {
		return antiPigeonRunner{}, err
	}
	for _, id := range antiPigeonIDs {
		event := testutil.Event(id, "observationally identical memory", experimentAnchor.Add(-time.Hour))
		event.TenantID, event.SessionID, event.Priority = "anti-pigeon-experiment", "source", 1
		event.Embedding, event.EmbeddingModel = []float32{1, 0, 0, 0, 0, 0, 0, 0}, embedder.ModelKey()
		if err := observeEvent(ctx, runtime, event); err != nil {
			return antiPigeonRunner{}, err
		}
	}
	if err := publishBayesianCertificates(ctx, runtime, memory.Snapshot(ctx), name); err != nil {
		return antiPigeonRunner{}, err
	}
	return antiPigeonRunner{name: name, runtime: runtime}, nil
}

func publishBayesianCertificates(ctx context.Context, runtime *service.Service, snapshot model.Snapshot, mode string) error {
	now := time.Now().UTC()
	selection := model.SelectionSupportCertificate{ID: "selection-" + mode, TenantID: "anti-pigeon-experiment", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, MinSelectionProbability: 1, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic frontier", Issuer: "claims-experiment", ExternalAudit: true, ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour)}
	if _, err := runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		return err
	}
	omitted := model.OmittedInfluenceCertificate{ID: "omitted-" + mode, TenantID: "anti-pigeon-experiment", PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch, DivergenceUCB: 0, DivergenceLimit: .05, AuditProbability: 1, SimultaneousCoverage: .95, Procedure: "exhaustive synthetic frontier", Issuer: "claims-experiment", ExternalAudit: true, ValidUntil: now.Add(time.Hour)}
	if _, err := runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted}); err != nil {
		return err
	}
	groups := [][]string(nil)
	if mode == "naive_shared" {
		groups = [][]string{antiPigeonIDs}
	} else if mode == "oracle_ap" {
		groups = [][]string{antiPigeonIDs[:2], antiPigeonIDs[2:]}
	}
	for index, members := range groups {
		certificate := model.AntiPigeonCertificate{ID: fmt.Sprintf("%s-%d", mode, index), TenantID: "anti-pigeon-experiment", MemberEventIDs: members, HorizonKey: model.RetrievalUsefulnessHorizon, GraphVersion: snapshot.GraphVersion, EvidenceEpoch: snapshot.EvidenceEpoch, TargetDiameterUCB: .01, DiameterLimit: .05, EffectiveSupport: 100, MinEffectiveSupport: 30, SimultaneousCoverage: .95, Procedure: "synthetic held-out target diameter", Issuer: "claims-experiment", ExternalAudit: true, ValidUntil: now.Add(time.Hour)}
		if _, err := runtime.PublishAntiPigeonCertificate(ctx, model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: certificate}); err != nil {
			return err
		}
	}
	return nil
}
