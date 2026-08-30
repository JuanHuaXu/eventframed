package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func BenchmarkResolutionPreference200Candidates(b *testing.B) {
	base := make([]model.Candidate, 200)
	for index := range base {
		base[index] = model.Candidate{Event: model.Event{ID: fmt.Sprintf("event-%03d", index)}, Score: .5}
	}
	for index := 0; index < 20; index++ {
		members := []string{fmt.Sprintf("event-%03d", index*3), fmt.Sprintf("event-%03d", index*3+1), fmt.Sprintf("event-%03d", index*3+2)}
		base[100+index].Event.Composition = &model.Composition{MemberEventIDs: members, RepresentativeEventID: members[0]}
	}
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		candidates := append([]model.Candidate(nil), base...)
		applyResolutionPreference(candidates, model.RecallResolutionCoarse)
	}
}

func BenchmarkRecallWith20AuthorizedCompositions(b *testing.B) {
	ctx := context.Background()
	memory := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(32)
	runtime, err := New(memory, embedder, Config{DefaultRecallK: 100, DefaultPackK: 20, DefaultTokenBudget: 10000})
	if err != nil {
		b.Fatal(err)
	}
	now := time.Now().UTC()
	pairs := make([][]string, 20)
	for group := range pairs {
		for member := 0; member < 2; member++ {
			id := fmt.Sprintf("stage-%02d-%d", group, member)
			pairs[group] = append(pairs[group], id)
			event := testutil.Event(id, fmt.Sprintf("public mission %02d stage %d", group, member), now.Add(-time.Hour))
			if _, err := runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}); err != nil {
				b.Fatal(err)
			}
		}
	}
	for group, members := range pairs {
		base := memory.Snapshot(ctx)
		certificate := model.AntiPigeonCertificate{ID: fmt.Sprintf("ap-%02d", group), TenantID: "tenant-a", MemberEventIDs: members, HorizonKey: model.RetrievalUsefulnessHorizon, GraphVersion: base.GraphVersion, EvidenceEpoch: base.EvidenceEpoch, TargetDiameterUCB: .01, DiameterLimit: .05, EffectiveSupport: 40, MinEffectiveSupport: 30, SimultaneousCoverage: .95, Procedure: "benchmark", Issuer: "benchmark", ExternalAudit: true, ValidUntil: now.Add(time.Hour)}
		certified, err := runtime.PublishAntiPigeonCertificate(ctx, model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: certificate})
		if err != nil {
			b.Fatal(err)
		}
		if _, err := runtime.ComposeInvariant(ctx, model.ComposeInvariantRequest{ProtocolVersion: model.ProtocolVersion, ID: fmt.Sprintf("mission-%02d", group), TenantID: "tenant-a", SessionID: "session-a", MemberEventIDs: members, RepresentativeEventID: members[0], Label: fmt.Sprintf("public mission %02d", group), RuleID: "benchmark-stages", Resolution: "mission", Confidence: .9, AntiPigeonCertificateID: certificate.ID, PublishedAt: now, BaseSnapshot: certified.Snapshot}); err != nil {
			b.Fatal(err)
		}
	}
	request := model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", SessionID: "session-a", Query: "public mission", AsOf: now.Add(time.Second), RecallK: 100, PackK: 20, TokenBudget: 10000, Resolution: model.RecallResolutionCoarse}
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		if _, err := runtime.Recall(ctx, request); err != nil {
			b.Fatal(err)
		}
	}
}
