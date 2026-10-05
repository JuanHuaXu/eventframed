package main

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/researchfusion"
)

func TestIsolatedServiceContract(t *testing.T) {
	base, _ := embed.NewHashEmbedder(32)
	em, _ := researchfusion.NewMemo(base, 64)
	facts := []fact{{"one", "A test control records a red square."}, {"two", "A test control records a green triangle."}, {"three", "A test control records a blue circle."}}
	q := query{"unit", "Which test control records a green triangle?"}
	var incumbent record
	for _, arm := range arms {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		r, err := runCase(ctx, em, facts, q, arm, true)
		cancel()
		if err != nil {
			t.Fatal(arm, err)
		}
		if len(r.Order) != 3 || !r.JournalStored || len(r.Frames) != 3 || r.MemoAfter.Misses != r.MemoBefore.Misses {
			t.Fatal("integration/cold leak", arm, r)
		}
		if arm == "baseline" {
			incumbent = r
			continue
		}
		if r.HookCalls != 1 || len(r.HookScores) != 3 || len(r.Lexical) != 3 {
			t.Fatal("hook coverage", arm, r)
		}
		if arm == "incumbent" {
			for i, v := range incumbent.Order {
				if v.ID != r.Order[i].ID {
					t.Fatal("incumbent order moved")
				}
			}
		}
		laws := map[string]any{}
		for _, v := range incumbent.Order {
			laws[v.ID] = v.Law
		}
		for _, v := range r.Order {
			if !reflect.DeepEqual(laws[v.ID], v.Law) {
				t.Fatal("rank changed law")
			}
		}
	}
}
