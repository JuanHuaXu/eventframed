package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/researchmemory"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

// Opt-in characterization, not a safety guarantee. Once the durable service
// owner is added, keep this artifact as the unsafe-composition control rather
// than treating duplicate-label acceptance as behavior to preserve.
func TestResearchDurableIdentityAudit(t *testing.T) {
	path := os.Getenv("EVENTFRAME_IDENTITY_AUDIT_ARTIFACT")
	if path == "" {
		t.Skip("opt-in lifecycle characterization")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	sources, hashes := map[string]string{}, map[string]string{}
	for _, name := range []string{"internal/service/research_durable_identity_audit_test.go", "internal/service/research_durable_validation.go", "internal/service/research_batch_validation.go", "internal/service/research_feedback.go", "internal/researchmemory/durable.go", "internal/researchmemory/durable_batch_admit.go", "internal/researchmemory/durable_prepared.go", "internal/researchmemory/replay.go", "internal/researchmemory/record.go"} {
		b, err := os.ReadFile("../../" + name)
		if err != nil {
			t.Fatal(err)
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
	}
	if err = enc.Encode(map[string]any{"kind": "header", "expected_cases": 8, "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for _, prepared := range []bool{false, true} {
		for _, batch := range []bool{false, true} {
			for _, reopen := range []bool{false, true} {
				t.Run(fmt.Sprintf("prepared%t/batch%t/reopen%t", prepared, batch, reopen), func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					s, bridge, in := bridgeFixture(t)
					// The older bounded bridge already refuses the same journal twice.
					if _, err = bridge.Admit(ctx, in); err != nil {
						t.Fatal(err)
					}
					_, bridgeErr := bridge.Admit(ctx, in)
					if bridgeErr == nil {
						t.Fatal("bridge control lost deduplication")
					}
					wrapped, e := researchpublicationstore.Wrap(s.store)
					if e != nil {
						t.Fatal(e)
					}
					s.store = wrapped
					open := researchmemory.OpenDurable
					if prepared {
						open = researchmemory.OpenDurablePreparedBatches
					}
					dbpath := t.TempDir() + "/identity.sqlite"
					d, e := open(ctx, dbpath, "tenant-a", "audit", 1, 42)
					if e != nil {
						t.Fatal(e)
					}
					defer func() { d.Close() }()
					c := in.Candidates[0]
					a := researchmemory.New(1, 42)
					p, e := a.Predict(c.Features, c.Baseline, 1, in.AsOf)
					if e != nil {
						t.Fatal(e)
					}
					r, e := a.Record(p.ID)
					if e != nil {
						t.Fatal(e)
					}
					r.Binding = &researchmemory.ServiceBinding{Tenant: "tenant-a", JournalID: in.JournalID, EventID: c.EventID, Snapshot: in.Snapshot}
					req := model.RecallRequest{TenantID: "tenant-a", SessionID: "session-a", Query: "public query", Embedding: []float32{1, 0, 0, 0, 0, 0, 0, 0}, EmbeddingModel: s.embedder.ModelKey(), AsOf: in.AsOf}
					admit := func(id uint64) (bool, error) {
						record := r
						record.Prediction.ID = id
						var retry bool
						err := s.WithValidatedResearchAdmission(ctx, record, req, func() error {
							if batch {
								got, e := d.AdmitBatchWithSnapshotReads(ctx, []researchmemory.AdmissionRequest{{ID: id, Features: c.Features, Baseline: c.Baseline, At: in.AsOf, Binding: r.Binding}})
								if e == nil {
									retry = got[0].Retry
								}
								return e
							}
							_, v, e := d.AdmitBound(ctx, id, c.Features, c.Baseline, in.AsOf, *r.Binding)
							retry = v
							return e
						})
						return retry, err
					}
					if retry, e := admit(1); e != nil || retry {
						t.Fatal("initial", retry, e)
					}
					if reopen {
						if e = d.Close(); e != nil {
							t.Fatal(e)
						}
						d, e = open(ctx, dbpath, "tenant-a", "audit", 1, 42)
						if e != nil {
							t.Fatal(e)
						}
					}
					exactRetry, e := admit(1)
					if e != nil || !exactRetry {
						t.Fatal("exact retry", e)
					}
					duplicateRetry, duplicateErr := admit(2)
					if duplicateErr != nil || duplicateRetry {
						t.Fatal("characterization changed: re-audit desired owner boundary", duplicateErr)
					}
					if _, e = s.store.BindBayesianPolicy(ctx, "after-original"); e != nil {
						t.Fatal(e)
					}
					_, staleAdmissionErr := admit(3)
					if staleAdmissionErr == nil {
						t.Fatal("stale admission control failed")
					}
					// Direct storage API has no service proof input. This deliberately
					// characterizes unsafe composition, not authorized evidence.
					for _, id := range []uint64{1, 2} {
						if _, e = d.Feedback(ctx, id, true, in.AsOf); e != nil {
							t.Fatal(e)
						}
					}
					if e = d.Close(); e != nil {
						t.Fatal(e)
					}
					log, e := researchledger.Open(dbpath)
					if e != nil {
						t.Fatal(e)
					}
					learner, e := researchmemory.ReplayLedger(ctx, log, "tenant-a", "audit", 1, 42)
					if e != nil {
						log.Close()
						t.Fatal(e)
					}
					labels, pending := learner.Counts()
					if e = log.Close(); e != nil {
						t.Fatal(e)
					}
					if labels != 2 || pending != 0 {
						t.Fatal("duplicate evidence not reproduced", labels, pending)
					}
					result := map[string]any{"Prepared": prepared, "Batch": batch, "Reopen": reopen, "BridgeRejectsDuplicate": bridgeErr != nil, "ExactIDRetry": exactRetry, "NewIDSameBindingAccepted": duplicateErr == nil, "StaleAdmissionRejected": staleAdmissionErr != nil, "RawFeedbackAcceptedAfterPolicyChange": true, "ReplayedLabelsFromOneJournalEvent": labels}
					if e = enc.Encode(result); e != nil {
						t.Fatal(e)
					}
					if e = f.Sync(); e != nil {
						t.Fatal(e)
					}
				})
			}
		}
	}
}
