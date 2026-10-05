package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

func TestResearchPostDiscardPublicationAndRecovery(t *testing.T) {
	for _, mode := range []string{"group4postdiscard", "group4postverify"} {
		for _, interrupted := range []bool{false, true} {
			t.Run(mode+"/"+map[bool]string{false: "normal", true: "closed-before-terminal"}[interrupted], func(t *testing.T) {
				s, _, in := bridgeFixture(t)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				wrapped, err := researchpublicationstore.Wrap(s.store)
				if err != nil {
					t.Fatal(err)
				}
				s.store = wrapped
				a := researchmemory.New(1, 42)
				var records []researchmemory.RecordedPrediction
				for _, c := range in.Candidates {
					p, err := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
					if err != nil {
						t.Fatal(err)
					}
					r, err := a.Record(p.ID)
					if err != nil {
						t.Fatal(err)
					}
					r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
					records = append(records, r)
				}
				req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
				path := t.TempDir() + "/terminal.sqlite"
				d, err := researchmemory.OpenDurable(ctx, path, "tenant-a", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { d.Close() }()
				groups := [][]researchmemory.RecordedPrediction{records}
				result := researchGuardLoadResult{Mode: mode}
				phase := researchGuardPhase{}
				err = s.WithValidatedResearchBatchAdmission(ctx, records, req, func() error {
					return researchPersistGroupBatch(ctx, d, groups, &phase, &result)
				})
				if err != nil || result.Discards != 0 || result.Admits != uint64(len(records)) {
					t.Fatal("guarded stage", err, result)
				}
				// A real owned publication must complete before terminal cleanup. This
				// would deadlock/time out if cleanup still held the service guard.
				if _, err = s.store.BindBayesianPolicy(ctx, "changed-before-discard"); err != nil {
					t.Fatal(err)
				}
				called := false
				if err = s.WithValidatedResearchBatchAdmission(ctx, records, req, func() error { called = true; return nil }); err == nil || called {
					t.Fatal("stale admission escaped", err)
				}
				if interrupted {
					if err = d.Close(); err != nil {
						t.Fatal(err)
					}
					if err = researchFinishGroup(ctx, d, groups, &phase, &result); err == nil || result.Discards != 0 {
						t.Fatal("failed terminal counted", err)
					}
					d, err = researchmemory.OpenDurable(ctx, path, "tenant-a", "stream", 1, 42)
					if err != nil {
						t.Fatal(err)
					}
				}
				if err = researchFinishGroup(ctx, d, groups, &phase, &result); err != nil || result.Discards != result.Admits {
					t.Fatal("terminal", err, result)
				}
				if err = d.Close(); err != nil {
					t.Fatal(err)
				}
				d, err = researchmemory.OpenDurable(ctx, path, "tenant-a", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				for _, original := range records {
					saved, err := d.Admission(ctx, original.Prediction.ID)
					if err != nil || !reflect.DeepEqual(saved, original) {
						t.Fatal("original changed", err)
					}
					if _, err = d.Feedback(ctx, original.Prediction.ID, true, in.AsOf); err == nil {
						t.Fatal("discard turned into evidence")
					}
				}
				// No learning clock may advance: a new forecast at the same time is
				// still cold and uses its supplied baseline after replay.
				p, retry, err := d.Admit(ctx, uint64(len(records)+1), 1, .37, in.AsOf)
				if err != nil || retry || p.Probability != .37 {
					t.Fatal("discard learned", p, err)
				}
			})
		}
	}
}

func TestResearchPostDiscardLoadAccounting(t *testing.T) {
	for _, mode := range []string{"group4postdiscard", "group4postverify"} {
		for _, writes := range []int{0, 4} {
			r := researchGuardLoadProbeArm(t, mode, 0, 8, writes, true)
			checkQueuedGuardLoad(t, r, writes)
			if r.Accepted == 0 {
				t.Fatal("no post-discard completions")
			}
			for _, p := range r.Phases {
				if p.Accepted && p.PostGuardNS <= 0 {
					t.Fatal("missing terminal timing")
				}
			}
		}
	}
}

func TestResearchPostVerifyRejectsIncompleteOriginals(t *testing.T) {
	for _, kind := range []string{"missing", "mismatch", "canceled"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			now := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
			d, err := researchmemory.OpenDurable(ctx, t.TempDir()+"/verify.sqlite", "tenant", "stream", 1, 42)
			if err != nil {
				t.Fatal(err)
			}
			defer d.Close()
			r, err := d.AdmitBatchWithSnapshotReads(ctx, []researchmemory.AdmissionRequest{{ID: 1, Features: 1, Baseline: .7, At: now}})
			if err != nil {
				t.Fatal(err)
			}
			original := r[0].Record
			bad := original
			switch kind {
			case "missing":
				bad.Prediction.ID = 999
			case "mismatch":
				bad.Prediction.Features ^= 1
			case "canceled":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			result := researchGuardLoadResult{Mode: "group4postverify", Admits: 1}
			phase := researchGuardPhase{}
			if err = researchFinishGroup(ctx, d, [][]researchmemory.RecordedPrediction{{bad}}, &phase, &result); err == nil || result.Discards != 0 {
				t.Fatal("invalid read completed", err, result)
			}
			// Valid retry must still reach a fresh terminal, proving the failed
			// integrity check did not silently discard the pending original.
			if err = researchFinishGroup(context.Background(), d, [][]researchmemory.RecordedPrediction{{original}}, &phase, &result); err != nil || result.Discards != 1 {
				t.Fatal("valid recovery", err, result)
			}
		})
	}
}
