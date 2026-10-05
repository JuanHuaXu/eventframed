package researchlineage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestLineageRestartAndTouchScope(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "lineage.sqlite")
	first := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
	l, err := Create(path, first)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if _, err := Open(path, first); err == nil {
		t.Fatal("second owner opened live sidecar")
	}
	second := first
	second.RuntimeVersion++
	second.EvidenceEpoch++
	if err := l.Record(ctx, first, second, []EventKey{{Tenant: "t", Event: "a"}}); err != nil {
		t.Fatal(err)
	}
	third := second
	third.RuntimeVersion++
	third.GraphVersion++
	if err := l.Record(ctx, second, third, nil); err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	l, err = Open(path, third)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	for _, tc := range []struct {
		from model.Snapshot
		key  EventKey
		want bool
	}{
		{first, EventKey{"t", "a"}, false},
		{second, EventKey{"t", "a"}, true},
		{first, EventKey{"t", "b"}, true},
	} {
		got, err := l.Unchanged(ctx, tc.from, third, tc.key)
		if err != nil || got != tc.want {
			t.Fatalf("continuity %+v: got=%v err=%v", tc.key, got, err)
		}
	}
	if same, err := l.Unchanged(ctx, first, second, EventKey{"t", "b"}); err == nil || same {
		t.Fatal("stale target accepted")
	}
	if err := l.Record(ctx, first, third, nil); err == nil {
		t.Fatal("stale predecessor accepted")
	}
	if _, err := Create(path, third); err == nil {
		t.Fatal("existing lineage reinitialized")
	}
}

func TestLineageWriteGapAndCorruptCheckpointFailClosed(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "lineage.sqlite")
	first := model.Snapshot{RuntimeVersion: 1}
	second := model.Snapshot{RuntimeVersion: 2}
	l, err := Create(path, first)
	if err != nil {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := l.Record(canceled, first, second, []EventKey{{Tenant: "t", Event: "a"}}); err == nil {
		t.Fatal("canceled sidecar write succeeded")
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path, second); err == nil {
		t.Fatal("backend commit without sidecar commit certified")
	}
	l, err = Open(path, first)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE lineage_meta SET committed='not-json' WHERE id=1"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path, first); err == nil {
		t.Fatal("corrupted checkpoint accepted")
	}
	missing := filepath.Join(t.TempDir(), "missing.sqlite")
	if _, err := Open(missing, first); !os.IsNotExist(err) {
		t.Fatalf("missing lineage was not rejected: %v", err)
	}
}
