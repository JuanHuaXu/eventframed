package researchmemory

import (
	"bytes"
	"context"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestMotionSealBindsOriginAndLabelsNotCurrentTarget(t *testing.T) {
	ctx := context.Background()
	labels, origin, cutoff := boundTransferFixture(64, true, false)
	retainB := func(_ context.Context, _ model.Snapshot, label BoundLabel) (bool, error) {
		return label.Prediction.Binding.EventID == "source-b", nil
	}
	key := []byte("0123456789abcdef0123456789abcdef")
	build := func(target, anchor model.Snapshot, source []BoundLabel) []byte {
		bootstrap, retained, err := RebuildMotionSealedBoundLabels(ctx, target, anchor, "tenant", "motion-stream", 2, 42, cutoff, source, retainB, "motion-key", key)
		if err != nil || retained != 32 {
			t.Fatalf("motion rebuild count=%d err=%v", retained, err)
		}
		return bootstrap.state.seal
	}
	first := build(origin, origin, labels)
	future := origin
	future.RuntimeVersion++
	if other := build(future, origin, labels); !bytes.Equal(first, other) {
		t.Fatal("harmless target motion changed origin-bound seal")
	}
	if other := build(future, future, labels); bytes.Equal(first, other) {
		t.Fatal("different origin kept the same motion seal")
	}
	changedKey := append([]byte(nil), key...)
	changedKey[0]++
	wrongKey, _, err := RebuildMotionSealedBoundLabels(ctx, future, origin, "tenant", "motion-stream", 2, 42, cutoff, labels, retainB, "motion-key", changedKey)
	if err != nil || bytes.Equal(first, wrongKey.state.seal) {
		t.Fatalf("changed key kept motion seal: %v", err)
	}
	retainA := func(_ context.Context, _ model.Snapshot, label BoundLabel) (bool, error) {
		return label.Prediction.Binding.EventID == "source-a", nil
	}
	changedDecision, retained, err := RebuildMotionSealedBoundLabels(ctx, future, origin, "tenant", "motion-stream", 2, 42, cutoff, labels, retainA, "motion-key", key)
	if err != nil || retained != 32 || bytes.Equal(first, changedDecision.state.seal) {
		t.Fatalf("changed retain decision kept motion seal: count=%d err=%v", retained, err)
	}
	changed := append([]BoundLabel(nil), labels...)
	changed[1].Feedback.Useful = !changed[1].Feedback.Useful
	if other := build(future, origin, changed); bytes.Equal(first, other) {
		t.Fatal("changed retained outcome kept the same motion seal")
	}
	legacy, _, err := RebuildSealedBoundLabels(ctx, origin, "tenant", "motion-stream", 2, 42, cutoff, labels, retainB, "motion-key", key)
	if err != nil || bytes.Equal(first, legacy.state.seal) {
		t.Fatal("legacy and motion seal domains collided")
	}
}
