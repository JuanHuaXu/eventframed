package researchbatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func prototypeWrite(id string, at time.Time) libravdbstore.ResearchEventWrite {
	return libravdbstore.ResearchEventWrite{
		Event:  testutil.Event(id, "public batch fixture", at),
		Vector: []float32{1, 0, 0, 0}, Digest: id,
	}
}

func openPrototypeForTest(t testing.TB, root string, create bool) *batchPrototype {
	t.Helper()
	p, err := openBatchPrototype(root, create)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBatchIntentCrashReconcile(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	for _, phase := range []string{"intent", "backend", "sidecar", ""} {
		t.Run("phase_"+phase, func(t *testing.T) {
			root := t.TempDir()
			p := openPrototypeForTest(t, root, true)
			before := p.backend.Snapshot(ctx)
			writes := []libravdbstore.ResearchEventWrite{prototypeWrite("a", at), prototypeWrite("b", at)}
			_, err := p.apply(ctx, writes, phase)
			if phase == "" && err != nil || phase != "" && !errors.Is(err, errInjectedCrash) {
				t.Fatal("wrong injected phase result", phase, err)
			}
			if phase != "" && p.asOf(ctx, before, at.Add(-time.Second)) {
				t.Fatal("pending operation granted protected as-of authority")
			}
			if err := p.close(); err != nil {
				t.Fatal(err)
			}
			p = openPrototypeForTest(t, root, false)
			defer p.close()
			committed := phase != "intent"
			wantVersion := before.RuntimeVersion
			if committed {
				wantVersion += 2
			}
			if p.snapshot.RuntimeVersion != wantVersion || p.snapshot.EvidenceEpoch != before.EvidenceEpoch+wantVersion-before.RuntimeVersion {
				t.Fatal("reopened checkpoint mismatch", p.snapshot)
			}
			if !p.asOf(ctx, before, at.Add(-time.Second)) || committed && p.asOf(ctx, before, at) {
				t.Fatal("incorrect reopened as-of decision")
			}
			retry, err := p.apply(ctx, writes, "")
			if err != nil || len(retry) != 2 || retry[0].Duplicate != committed || retry[1].Duplicate != committed {
				t.Fatal("exact retry did not resolve original state", retry, err)
			}
			if p.snapshot.RuntimeVersion != before.RuntimeVersion+2 {
				t.Fatal("retry lost or duplicated accepted versions")
			}
			for _, version := range []uint64{before.RuntimeVersion + 1, before.RuntimeVersion + 2} {
				if !p.motion[version].Equal(at) {
					t.Fatal("missing per-event motion", version)
				}
			}
		})
	}
}

func TestBatchIntentMixedRetryAndBackfill(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	p := openPrototypeForTest(t, t.TempDir(), true)
	defer p.close()
	initial := p.snapshot
	if _, err := p.apply(ctx, []libravdbstore.ResearchEventWrite{prototypeWrite("a", at)}, ""); err != nil {
		t.Fatal(err)
	}
	mixed := []libravdbstore.ResearchEventWrite{
		prototypeWrite("a", at), prototypeWrite("b", at), prototypeWrite("b", at), prototypeWrite("c", at),
	}
	results, err := p.apply(ctx, mixed, "")
	if err != nil || len(results) != 4 {
		t.Fatal(err)
	}
	for i, want := range []bool{true, false, true, false} {
		if results[i].Duplicate != want {
			t.Fatal("wrong duplicate assignment", i, results[i])
		}
	}
	if p.snapshot.RuntimeVersion != initial.RuntimeVersion+3 || !p.asOf(ctx, initial, at.Add(-time.Second)) {
		t.Fatal("mixed batch lost versions or future-only authority")
	}
	for i, id := range []string{"a", "b", "c"} {
		var version uint64
		if err := p.db.QueryRowContext(ctx, "SELECT version FROM accepted WHERE tenant=? AND event=?", "tenant-a", id).Scan(&version); err != nil || version != initial.RuntimeVersion+uint64(i)+1 {
			t.Fatal("source touch version mismatch", id, version, err)
		}
	}
	for _, bad := range []libravdbstore.ResearchEventWrite{
		func() libravdbstore.ResearchEventWrite {
			w := prototypeWrite("a", at)
			w.Digest = "different"
			return w
		}(),
		func() libravdbstore.ResearchEventWrite {
			w := prototypeWrite("a", at)
			w.Event.Content = "changed payload"
			return w
		}(),
	} {
		if _, err := p.apply(ctx, []libravdbstore.ResearchEventWrite{bad}, ""); err == nil {
			t.Fatal("changed retry accepted")
		}
	}
	if p.snapshot.RuntimeVersion != initial.RuntimeVersion+3 {
		t.Fatal("conflict changed checkpoint")
	}
	if _, err := p.apply(ctx, []libravdbstore.ResearchEventWrite{prototypeWrite("backfill", at.Add(-time.Hour))}, ""); err != nil {
		t.Fatal(err)
	}
	if p.asOf(ctx, initial, at.Add(-time.Second)) {
		t.Fatal("visible backfill accepted as future-only")
	}
}

func TestBatchIntentCorruptionAndMismatchFailClosed(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	for _, corruption := range []string{"malformed_intent", "wrong_payload", "extra_backend_write"} {
		t.Run(corruption, func(t *testing.T) {
			root := t.TempDir()
			p := openPrototypeForTest(t, root, true)
			writes := []libravdbstore.ResearchEventWrite{prototypeWrite("a", at), prototypeWrite("b", at)}
			phase := "backend"
			if corruption == "malformed_intent" {
				phase = "intent"
			}
			if _, err := p.apply(ctx, writes, phase); !errors.Is(err, errInjectedCrash) {
				t.Fatal("missing crash", err)
			}
			switch corruption {
			case "malformed_intent":
				if _, err := p.db.ExecContext(ctx, "UPDATE intent SET payload=? WHERE id=1", []byte("not-json")); err != nil {
					t.Fatal(err)
				}
			case "wrong_payload":
				intent, err := p.readIntent(ctx)
				if err != nil {
					t.Fatal(err)
				}
				intent.Writes[1].Event.Content = "different event"
				encoded, err := json.Marshal(intent)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := p.db.ExecContext(ctx, "UPDATE intent SET payload=? WHERE id=1", encoded); err != nil {
					t.Fatal(err)
				}
			case "extra_backend_write":
				if _, err := p.backend.PutResearchEventBatch(ctx, []libravdbstore.ResearchEventWrite{prototypeWrite("extra", at)}); err != nil {
					t.Fatal(err)
				}
			}
			if err := p.close(); err != nil {
				t.Fatal(err)
			}
			if reopened, err := openBatchPrototype(root, false); err == nil {
				_ = reopened.close()
				t.Fatal("corrupt or unknown state reopened as valid")
			}
		})
	}
	root := t.TempDir()
	backend, err := libravdbstore.Open(libravdbstore.Config{
		Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}
	if reopened, err := openBatchPrototype(root, false); err == nil {
		_ = reopened.close()
		t.Fatal("missing sidecar recreated")
	}
}

func TestBatchIntentRetryWithMonotonicClock(t *testing.T) {
	ctx := context.Background()
	at := time.Now().Add(time.Second)
	root := t.TempDir()
	p := openPrototypeForTest(t, root, true)
	writes := []libravdbstore.ResearchEventWrite{prototypeWrite("monotonic", at)}
	if _, err := p.apply(ctx, writes, ""); err != nil {
		t.Fatal("first write rejected", err)
	}
	if err := p.close(); err != nil {
		t.Fatal(err)
	}
	p = openPrototypeForTest(t, root, false)
	defer p.close()
	results, err := p.apply(ctx, writes, "")
	if err != nil || len(results) != 1 || !results[0].Duplicate {
		t.Fatal("exact persisted retry rejected", results, err)
	}
}

func TestBatchIntentTamperedAcceptedRowCannotAuthorizeDuplicate(t *testing.T) {
	ctx := context.Background()
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	p := openPrototypeForTest(t, t.TempDir(), true)
	defer p.close()
	original := prototypeWrite("a", at)
	if _, err := p.apply(ctx, []libravdbstore.ResearchEventWrite{original}, ""); err != nil {
		t.Fatal(err)
	}
	forged := original
	forged.Event.Content = "forged content"
	encoded, err := json.Marshal(forged)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.db.ExecContext(ctx, "UPDATE accepted SET payload=? WHERE tenant=? AND event=?", encoded, original.Event.TenantID, original.Event.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.apply(ctx, []libravdbstore.ResearchEventWrite{forged}, ""); err == nil {
		t.Fatal("sidecar-only duplicate authority accepted forged event")
	}
}

func TestBatchIntentCannotCreateSidecarForExistingHistory(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	backend, err := libravdbstore.Open(libravdbstore.Config{
		Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if _, err := backend.PutResearchEventBatch(ctx, []libravdbstore.ResearchEventWrite{prototypeWrite("prior", at)}); err != nil {
		t.Fatal(err)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}
	if p, err := openBatchPrototype(root, true); err == nil {
		_ = p.close()
		t.Fatal("created empty authority history over existing backend event")
	}
}

func BenchmarkBatchIntentPrototype(b *testing.B) {
	at := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	for _, mode := range []string{"raw", "intent"} {
		for _, size := range []int{1, 4, 16} {
			b.Run(fmt.Sprintf("%s/batch%d", mode, size), func(b *testing.B) {
				b.ReportAllocs()
				for repeat := 0; repeat < b.N; repeat++ {
					b.StopTimer()
					root := b.TempDir()
					writes := make([]libravdbstore.ResearchEventWrite, 128)
					for i := range writes {
						writes[i] = prototypeWrite(fmt.Sprintf("event-%d", i), at)
					}
					if mode == "raw" {
						backend, err := libravdbstore.Open(libravdbstore.Config{
							Path: filepath.Join(root, "events.libravdb"), Dimension: 4,
							EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true,
						})
						if err != nil {
							b.Fatal(err)
						}
						b.StartTimer()
						for i := 0; i < len(writes); i += size {
							if _, err := backend.PutResearchEventBatch(context.Background(), writes[i:i+size]); err != nil {
								b.Fatal(err)
							}
						}
						b.StopTimer()
						if err := backend.Close(); err != nil {
							b.Fatal(err)
						}
						continue
					}
					p := openPrototypeForTest(b, root, true)
					b.StartTimer()
					for i := 0; i < len(writes); i += size {
						if _, err := p.apply(context.Background(), writes[i:i+size], ""); err != nil {
							b.Fatal(err)
						}
					}
					b.StopTimer()
					if err := p.close(); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
