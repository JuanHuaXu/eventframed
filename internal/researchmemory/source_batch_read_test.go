package researchmemory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
)

func openBatchSourceTest(t *testing.T, path string) *SourceOwner {
	t.Helper()
	o, err := OpenSourceOwnerBatchReads(context.Background(), path, "tenant", "stream", 1, 42)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { o.Close() })
	return o
}

func TestSourceBatchReadWarmParity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	paths := []string{t.TempDir() + "/point.sqlite", t.TempDir() + "/batch.sqlite", t.TempDir() + "/resolved.sqlite"}
	owners := []*SourceOwner{openSourceTest(t, paths[0]), openBatchSourceTest(t, paths[1]), openResolvedSourceTest(t, paths[2])}
	// Test-only trusted labels with identical publication barriers. This is not
	// an exposed feedback path; equal final counts alone do not match histories.
	for id := 1; id <= 64; id++ {
		req := sourceRequest()
		req.Features = uint16(id)
		req.Binding.EventID = fmt.Sprint(id)
		var first RecordedPrediction
		for i, o := range owners {
			r, err := o.Admit(ctx, []SourceAdmissionRequest{req})
			if err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				first = r[0].Record
			} else if !reflect.DeepEqual(first, r[0].Record) {
				t.Fatal("training originals differ")
			}
			if _, err = o.d.Feedback(ctx, r[0].Record.Prediction.ID, id%3 == 0, req.At); err != nil {
				t.Fatal(err)
			}
			if err = o.d.worker.WaitProcessed(ctx, uint64(id)); err != nil {
				t.Fatal(err)
			}
		}
	}
	requests := make([]SourceAdmissionRequest, 32)
	refs := make([]SourceReference, 32)
	discards := make([]SourceDiscardRequest, 32)
	for i := range requests {
		r := sourceRequest()
		r.Features = uint16(i)
		r.Binding.EventID = fmt.Sprint(65 + i)
		requests[i] = r
		refs[i] = SourceReference{r.Binding.JournalID, r.Binding.EventID}
		discards[i] = SourceDiscardRequest{r.Binding.JournalID, r.Binding.EventID, r.At}
	}
	var original []AdmissionResult
	for i, o := range owners {
		r, err := o.Admit(ctx, requests)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			original = r
		} else if !reflect.DeepEqual(original, r) {
			t.Fatal("warm originals differ")
		}
		read, err := o.LookupBatch(ctx, refs)
		if err != nil {
			t.Fatal(err)
		}
		for j, v := range r {
			if !v.Record.Ready || v.Retry || !reflect.DeepEqual(v.Record, read[j]) {
				t.Fatal("warm readback mismatch")
			}
		}
		if err = o.Close(); err != nil {
			t.Fatal(err)
		}
		// Swap constructors on reopen to show the format/meaning did not change.
		if i == 0 {
			o = openBatchSourceTest(t, paths[i])
		} else {
			o = openSourceTest(t, paths[i])
		}
		owners[i] = o
		retry, err := o.Admit(ctx, requests)
		if err != nil {
			t.Fatal(err)
		}
		for j, v := range retry {
			if !v.Retry || !reflect.DeepEqual(v.Record, original[j].Record) {
				t.Fatal("replayed original changed")
			}
		}
		if _, err = o.DiscardBatch(ctx, discards); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSourceBatchReadPreflightAndAliases(t *testing.T) {
	o := openBatchSourceTest(t, t.TempDir()+"/batch.sqlite")
	ctx := context.Background()
	a, b := sourceRequest(), sourceRequest()
	b.Binding.EventID = "second"
	if _, err := o.Admit(ctx, []SourceAdmissionRequest{a, b}); err != nil {
		t.Fatal(err)
	}
	refs := []SourceReference{{b.Binding.JournalID, b.Binding.EventID}, {a.Binding.JournalID, a.Binding.EventID}, {b.Binding.JournalID, b.Binding.EventID}}
	r, err := o.LookupBatch(ctx, refs)
	if err != nil {
		t.Fatal(err)
	}
	r[0].Binding.EventID = "mutated"
	if r[2].Binding.EventID != "second" {
		t.Fatal("aliased readback")
	}
	for _, bad := range [][]SourceReference{nil, make([]SourceReference, 257), {refs[0], {JournalID: "journal", EventID: "missing"}}} {
		if got, e := o.LookupBatch(ctx, bad); e == nil || got != nil {
			t.Fatal("partial/invalid read", e)
		}
	}
	fresh, bad := a, a
	fresh.Binding.EventID = "new"
	bad.Binding.Snapshot.PolicyVersion++
	if got, e := o.Admit(ctx, []SourceAdmissionRequest{fresh, bad}); e == nil || got != nil {
		t.Fatal("changed source accepted", e)
	}
	if got, e := o.Admit(ctx, []SourceAdmissionRequest{fresh, fresh}); e == nil || got != nil {
		t.Fatal("duplicate source accepted", e)
	}
	if _, e := o.Lookup(ctx, "journal", "new"); !errors.Is(e, sql.ErrNoRows) {
		t.Fatal("partial staged original", e)
	}
	discards := []SourceDiscardRequest{{"journal", "event", a.At}, {"journal", "missing", a.At}}
	if got, e := o.DiscardBatch(ctx, discards); e == nil || got != nil {
		t.Fatal("partial discard", e)
	}
	if n, f, p, q := o.d.worker.Counts(); n != 0 || f != 0 || p != 2 || q != 0 || o.d.stopped {
		t.Fatal(n, f, p, q, o.d.stopped)
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if got, e := o.LookupBatch(c, refs); !errors.Is(e, context.Canceled) || got != nil {
		t.Fatal("canceled read", e)
	}
	if got, e := o.Admit(c, []SourceAdmissionRequest{fresh}); !errors.Is(e, context.Canceled) || got != nil {
		t.Fatal("canceled admit", e)
	}
	if _, e := o.Admit(ctx, []SourceAdmissionRequest{fresh}); e != nil {
		t.Fatal("preflight stopped owner", e)
	}
}

func TestSourceBatchReadRejectsMalformedResults(t *testing.T) {
	o := openBatchSourceTest(t, t.TempDir()+"/batch.sqlite")
	ctx := context.Background()
	req := sourceRequest()
	if _, err := o.Admit(ctx, []SourceAdmissionRequest{req}); err != nil {
		t.Fatal(err)
	}
	keys := []researchledger.ServiceIdentity{o.key("journal", "event")}
	valid, err := o.log.GetServiceAdmissions(ctx, keys)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"count", "source", "missing", "sequence", "kind", "stream", "canonical", "seed", "epoch", "binding", "id", "oversized"} {
		t.Run(kind, func(t *testing.T) {
			rows := append([]researchledger.ServiceLookupResult(nil), valid...)
			rows[0].Entry.Payload = append([]byte(nil), valid[0].Entry.Payload...)
			switch kind {
			case "count":
				rows = nil
			case "source":
				rows[0].Source.Event = "other"
			case "missing":
				rows[0].Found = false
			case "sequence":
				rows[0].Entry.Sequence = 0
			case "kind":
				rows[0].Entry.Kind = "feedback"
			case "stream":
				rows[0].Entry.Key.Journal = "other"
			case "canonical":
				rows[0].Entry.Payload = append(rows[0].Entry.Payload, ' ')
			case "oversized":
				rows[0].Entry.Payload = make([]byte, (1<<20)+1)
			default:
				var r RecordedPrediction
				if e := json.Unmarshal(rows[0].Entry.Payload, &r); e != nil {
					t.Fatal(e)
				}
				switch kind {
				case "seed":
					r.Seed++
				case "epoch":
					r.Prediction.Epoch++
				case "binding":
					r.Binding.EventID = "other"
				case "id":
					r.Prediction.ID++
				}
				rows[0].Entry.Payload, err = json.Marshal(r)
				if err != nil {
					t.Fatal(err)
				}
			}
			if got, e := o.decodeSourceBatch(keys, rows); e == nil || got != nil {
				t.Fatal("malformed result accepted", kind, e)
			}
		})
	}
}

func TestSourceBatchReadCommitRecovery(t *testing.T) {
	for _, discard := range []bool{false, true} {
		for _, commit := range []bool{false, true} {
			t.Run(fmt.Sprintf("discard%t/commit%t", discard, commit), func(t *testing.T) {
				ctx := context.Background()
				path := t.TempDir() + "/batch.sqlite"
				o := openBatchSourceTest(t, path)
				r := sourceRequest()
				admit := []SourceAdmissionRequest{r}
				terminal := []SourceDiscardRequest{{r.Binding.JournalID, r.Binding.EventID, r.At}}
				if discard {
					if _, e := o.Admit(ctx, admit); e != nil {
						t.Fatal(e)
					}
				}
				o.d.log = sourceFaultLog{o.d.log.(preparedBatchLedger), commit}
				var err error
				if discard {
					_, err = o.DiscardBatch(ctx, terminal)
				} else {
					_, err = o.Admit(ctx, admit)
				}
				if err == nil {
					t.Fatal("uncertain acknowledgment accepted")
				}
				if got, e := o.LookupBatch(ctx, []SourceReference{{r.Binding.JournalID, r.Binding.EventID}}); e == nil || got != nil {
					t.Fatal("stopped owner served")
				}
				if err = o.Close(); err != nil {
					t.Fatal(err)
				}
				o = openBatchSourceTest(t, path)
				if discard {
					got, e := o.DiscardBatch(ctx, terminal)
					if e != nil || got[0] != commit {
						t.Fatal(got, e)
					}
				} else {
					got, e := o.Admit(ctx, admit)
					if e != nil || got[0].Retry != commit || got[0].Record.Prediction.ID != 1 {
						t.Fatal(got, e)
					}
				}
			})
		}
	}
}
