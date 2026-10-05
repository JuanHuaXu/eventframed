package researchpublicfrontierunion

import (
	"context"
	"fmt"
	fb "github.com/JuanHuaXu/eventframed/internal/researchpublicfeedback"
	r "github.com/JuanHuaXu/eventframed/internal/researchpublicpairrank"
	"testing"
	"time"
)

func TestUnionScoresFullSetAndDuplicates(t *testing.T) {
	ctx := context.Background()
	at := time.Unix(100, 0)
	docs := make([]fb.Document, 400)
	a, b := []fb.Hit{}, []fb.Hit{}
	for j := range docs {
		docs[j] = fb.Document{ID: fmt.Sprint(j), Text: "needle", AvailableAt: at}
		if j < 200 {
			a = append(a, fb.Hit{ID: docs[j].ID, Score: 1})
		} else {
			b = append(b, fb.Hit{ID: docs[j].ID, Score: 1})
		}
	}
	x, e := fb.New(ctx, docs, at)
	if e != nil {
		t.Fatal(e)
	}
	v, e := OriginalScores(ctx, x, "needle", a, b)
	if e != nil || len(v) != 400 {
		t.Fatal(e, len(v))
	}
	for _, h := range v {
		if h.Score <= 0 {
			t.Fatal("missing score")
		}
	}
	bad := append([]fb.Hit(nil), b...)
	bad[1] = bad[0]
	if _, e = OriginalScores(ctx, x, "needle", a, bad); e == nil {
		t.Fatal("duplicate accepted")
	}
	bad[1] = fb.Hit{ID: "unknown", Score: 1}
	if _, e = OriginalScores(ctx, x, "needle", a, bad); e == nil {
		t.Fatal("unknown accepted")
	}
}
func TestUnionScoreBeforeCap(t *testing.T) {
	rows := make([]r.Row, 400)
	for j := range rows {
		rows[j] = r.Row{ID: fmt.Sprint(j), Base: .5}
	}
	rows[399].Base = 1
	v, e := Rank(context.Background(), r.Model{}, rows, 200)
	if e != nil || len(v) != 200 || v[0].ID != "399" {
		t.Fatal(e)
	}
	rows[300] = rows[0]
	if _, e = Rank(context.Background(), r.Model{}, rows, 200); e == nil {
		t.Fatal("tail duplicate escaped cap")
	}
	c, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = Rank(c, r.Model{}, nil, 200); e == nil {
		t.Fatal("cancel accepted")
	}
	if _, e = Rank(context.Background(), r.Model{}, make([]r.Row, 401), 200); e == nil {
		t.Fatal("cap exceeded")
	}
}
