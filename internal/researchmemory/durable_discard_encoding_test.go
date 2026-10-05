package researchmemory

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestDurableBatchDiscardPreservesOriginalTimeEncoding(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		for _, snapshot := range []bool{false, true} {
			t.Run(fmt.Sprintf("prepared%t/snapshot%t", prepared, snapshot), func(t *testing.T) {
				ctx := context.Background()
				path := t.TempDir() + "/encoding.sqlite"
				open := OpenDurable
				if prepared {
					open = OpenDurablePreparedBatches
				}
				d, err := open(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { d.Close() }()
				at := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
				if _, err = d.AdmitBatch(ctx, []AdmissionRequest{{ID: 1, Features: 1, Baseline: .6, At: at}, {ID: 2, Features: 2, Baseline: .6, At: at}}); err != nil {
					t.Fatal(err)
				}
				if _, err = d.Discard(ctx, 1, at); err != nil {
					t.Fatal(err)
				}
				original, err := d.log.Get(ctx, d.key(1), "feedback")
				if err != nil {
					t.Fatal(err)
				}
				retryAt := at.In(time.FixedZone("same-instant", 7200))
				if retry, err := d.Discard(ctx, 1, retryAt); err != nil || !retry {
					t.Fatal("single control", retry, err)
				}
				discard := d.DiscardBatch
				if snapshot {
					discard = d.DiscardBatchWithSnapshotReads
				}
				if got, err := discard(ctx, []DiscardRequest{{1, retryAt}, {2, at}}); err != nil || !reflect.DeepEqual(got, []bool{true, false}) {
					t.Fatal("equivalent instant batch retry", got, err)
				}
				stored, err := d.log.Get(ctx, d.key(1), "feedback")
				if err != nil || !bytes.Equal(original.Payload, stored.Payload) {
					t.Fatal("original terminal rewritten", err)
				}
				if err = d.Close(); err != nil {
					t.Fatal(err)
				}
				d, err = open(ctx, path, "tenant", "stream", 1, 42)
				if err != nil {
					t.Fatal(err)
				}
				if got, err := d.DiscardBatchWithSnapshotReads(ctx, []DiscardRequest{{1, at}, {2, retryAt}}); err != nil || !reflect.DeepEqual(got, []bool{true, true}) {
					t.Fatal("reopen retry", got, err)
				}
			})
		}
	}
}
