package researchmemory

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchledger"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestBoundDurableEpochRestartAndSeal(t *testing.T) {
	ctx := context.Background()
	labels, target, cutoff := boundTransferFixture(64, true, false)
	key := make([]byte, 32)
	key[0] = 9
	build := func(source []BoundLabel) *SealedBoundBootstrap {
		bootstrap, retained, err := RebuildSealedBoundLabels(ctx, target, "tenant", "new-epoch", 2, 42, cutoff, source, retainSourceB, "test-key", key)
		if err != nil || retained != 32 || bootstrap == nil || len(bootstrap.state.seal) != 32 {
			t.Fatalf("rebuild retained=%d err=%v", retained, err)
		}
		return bootstrap
	}
	bootstrap := build(labels)
	path := t.TempDir() + "/bound.sqlite"
	validateOriginal := func(_ context.Context, r RecordedPrediction) error {
		if r.Witness == nil || r.Binding == nil {
			return errors.New("missing source witness")
		}
		e := testutil.Event(r.Binding.EventID, "public fixture", r.At.Add(-time.Second))
		e.TenantID = "tenant"
		matched, err := r.Witness.Verify("test-key", key, *r.Binding, strings.Repeat("0", 64), e, r.Prediction.Features, r.Outer[0])
		if err != nil || !matched {
			return errors.New("invalid source witness")
		}
		return nil
	}
	d, err := OpenBoundDurable(ctx, path, bootstrap, validateOriginal)
	if err != nil {
		t.Fatal(err)
	}
	if duplicate, err := OpenBoundDurable(ctx, t.TempDir()+"/other.sqlite", bootstrap, validateOriginal); err == nil {
		duplicate.Close()
		t.Fatal("one sealed adapter opened by two workers")
	}
	copyOfBootstrap := *bootstrap
	if duplicate, err := OpenBoundDurable(ctx, t.TempDir()+"/copy.sqlite", &copyOfBootstrap, validateOriginal); err == nil {
		duplicate.Close()
		t.Fatal("copied bootstrap opened a second worker")
	}
	firstAt := cutoff.Add(time.Second)
	binding := ServiceBinding{Tenant: "tenant", JournalID: "new-journal", EventID: "new-source", Snapshot: target}
	event := testutil.Event(binding.EventID, "public fixture", firstAt.Add(-time.Second))
	event.TenantID = "tenant"
	witness, err := NewSourceWitness("test-key", key, binding, strings.Repeat("0", 64), event, 1, .4)
	if err != nil {
		t.Fatal(err)
	}
	p, retried, err := d.AdmitSourceBoundWithWitness(ctx, 1, .4, firstAt, binding, witness)
	if err != nil || retried || p.ID != 1 || p.Epoch != 2 {
		t.Fatalf("first admission %+v retry=%v err=%v", p, retried, err)
	}
	if same, retry, err := d.AdmitSourceBoundWithWitness(ctx, 1, .4, firstAt, binding, witness); err != nil || !retry || same != p {
		t.Fatalf("exact source retry %+v retry=%v err=%v", same, retry, err)
	}
	if _, _, err := d.AdmitSourceBoundWithWitness(ctx, 2, .4, firstAt, binding, witness); err == nil {
		t.Fatal("changed source features became a second admission")
	}
	available := firstAt.Add(time.Second)
	if retried, err = d.Feedback(ctx, 1, true, available); err != nil || retried {
		t.Fatalf("feedback retry=%v err=%v", retried, err)
	}
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err = d.WaitProcessed(waitCtx, 33); err != nil {
		t.Fatal(err)
	}
	probeAt := available.Add(time.Second)
	want, err := d.worker.Snapshot().Score(1, .4, 2, probeAt)
	if err != nil {
		t.Fatal(err)
	}
	if err = d.Close(); err != nil {
		t.Fatal(err)
	}
	if cold, err := OpenDurable(ctx, path, "tenant", "new-epoch", 2, 42); err == nil {
		cold.Close()
		t.Fatal("ordinary cold replay accepted a bound epoch")
	}
	changedLabels := append([]BoundLabel(nil), labels...)
	changedLabels[1].Feedback.Useful = !changedLabels[1].Feedback.Useful
	changedBootstrap := build(changedLabels)
	if changed, err := OpenBoundDurable(ctx, path, changedBootstrap, validateOriginal); err == nil {
		changed.Close()
		t.Fatal("changed transferred evidence accepted")
	}
	bootstrap = build(labels)
	key[0] ^= 1
	if invalid, err := OpenBoundDurable(ctx, path, bootstrap, validateOriginal); err == nil {
		invalid.Close()
		t.Fatal("replay accepted a wrong witness key")
	}
	key[0] ^= 1
	bootstrap = build(labels)
	reopened, err := OpenBoundDurable(ctx, path, bootstrap, validateOriginal)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	got, err := reopened.worker.Snapshot().Score(1, .4, 2, probeAt)
	if err != nil || got != want {
		t.Fatalf("restart score got=%g want=%g err=%v", got, want, err)
	}
	if n, failed, pending, queued := reopened.Counts(); n != 33 || failed != 0 || pending != 0 || queued != 0 {
		t.Fatalf("replayed counts %d/%d/%d/%d", n, failed, pending, queued)
	}
	if same, retry, err := reopened.AdmitSourceBoundWithWitness(ctx, 1, .4, firstAt, binding, witness); err != nil || !retry || same != p {
		t.Fatalf("restart source retry %+v retry=%v err=%v", same, retry, err)
	}
	if original, err := reopened.AdmissionBySource(ctx, binding.JournalID, binding.EventID); err != nil || original.Prediction != p || original.Witness == nil {
		t.Fatalf("source original changed on restart: %+v err=%v", original, err)
	}
	nextBinding := ServiceBinding{Tenant: "tenant", JournalID: "later-journal", EventID: "later-source", Snapshot: target}
	nextEvent := testutil.Event(nextBinding.EventID, "public fixture", probeAt.Add(-time.Second))
	nextEvent.TenantID = "tenant"
	nextWitness, err := NewSourceWitness("test-key", key, nextBinding, strings.Repeat("0", 64), nextEvent, 1, .4)
	if err != nil {
		t.Fatal(err)
	}
	next, retried, err := reopened.AdmitSourceBoundWithWitness(ctx, 1, .4, probeAt, nextBinding, nextWitness)
	if err != nil || retried || next.Probability != want {
		t.Fatalf("continuing forecast %+v retry=%v want=%g err=%v", next, retried, want, err)
	}
}

func TestBoundDurableCannotBindAnExistingColdLog(t *testing.T) {
	ctx := context.Background()
	path := t.TempDir() + "/cold.sqlite"
	log, err := researchledger.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = log.Append(ctx, researchledger.Key{Tenant: "tenant", Journal: "stream", Event: "1", Contract: RecordContract}, "admit", []byte(`{"fixture":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if err = log.Close(); err != nil {
		t.Fatal(err)
	}
	labels, target, cutoff := boundTransferFixture(64, true, false)
	key := make([]byte, 32)
	bootstrap, _, err := RebuildSealedBoundLabels(ctx, target, "tenant", "stream", 2, 42, cutoff, labels, retainSourceB, "test-key", key)
	if err != nil {
		t.Fatal(err)
	}
	if d, err := OpenBoundDurable(ctx, path, bootstrap, func(context.Context, RecordedPrediction) error { return nil }); err == nil {
		d.Close()
		t.Fatal("bound bootstrap attached to an existing cold log")
	}
}
