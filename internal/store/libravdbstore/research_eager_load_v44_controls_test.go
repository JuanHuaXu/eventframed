package libravdbstore

import (
	"context"
	"errors"
	"go/ast"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

func eagerPredicatesV44(t *testing.T, file, method string) []string {
	t.Helper()
	f := researchASTV34(t, file, method)
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if id.Name == "e" {
				id.Name = "err"
			}
			if id.Name == "eager" {
				id.Name = "enabled"
			}
		}
		return true
	})
	var predicates []string
	for _, stmt := range f.Body.List {
		if branch, ok := stmt.(*ast.IfStmt); ok {
			predicates = append(predicates, researchExprV34(branch.Cond))
		}
	}
	return predicates
}

func TestResearchEagerPredicatePreservationV44(t *testing.T) {
	base := eagerPredicatesV44(t, "research_eager_core_v42_test.go", "refresh")
	fork := eagerPredicatesV44(t, "research_eager_load_v44_test.go", "refresh")
	if len(base) < 8 || !reflect.DeepEqual(base, fork) {
		t.Fatal("publication enable/cancel/head/poison/certificate/reuse/encoding predicates changed", base, fork)
	}
	bad := append([]string(nil), fork...)
	bad[2] = "false"
	if reflect.DeepEqual(base, bad) {
		t.Fatal("predicate deletion escaped")
	}
}

func TestResearchEagerWorkloadGenerationV44(t *testing.T) {
	if _, err := os.Stat("research_eager_load_v44_generated_test.go"); errors.Is(err, os.ErrNotExist) {
		t.Skip("V44 generator not executed yet; no preservation claim")
	}
	base := researchASTV34(t, "research_warm_search_load_v37_generated_test.go", "runWarmSearchLoadV37")
	fork := researchASTV34(t, "research_eager_load_v44_generated_test.go", "runEagerLoadV44")
	back := map[string]string{"runEagerLoadV44": "runWarmSearchLoadV37", "eagerTrialV44": "warmTrialV37", "attachLoadEagerV44": "attachLoadWarmV37", "finishEagerTrialV44": "finishWarmTrialV37"}
	if len(fork.Type.Params.List) != len(base.Type.Params.List)+1 {
		t.Fatal("extra signature drift")
	}
	fork.Type.Params.List = fork.Type.Params.List[:len(fork.Type.Params.List)-1]
	attachments := 0
	ast.Inspect(fork, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "attachLoadEagerV44" {
				if len(c.Args) != 4 || researchExprV34(c.Args[3]) != "eager" {
					t.Fatal("eager attachment")
				}
				c.Args = c.Args[:3]
				attachments++
			}
		}
		if id, ok := n.(*ast.Ident); ok {
			if name, ok := back[id.Name]; ok {
				id.Name = name
			}
		}
		return true
	})
	if attachments != 1 || !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("undeclared offered workload change")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("workload deletion escaped")
	}
}

func TestResearchEagerTracedOwnerAndDurabilityV44(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_EAGER_LOAD_V44") != "1" {
		t.Skip("explicit isolated native control")
	}
	for _, mode := range []int{-2, 4} {
		for _, enabled := range []bool{false, true} {
			t.Run(map[int]string{-2: "joined", 4: "archive"}[mode]+map[bool]string{false: "-lazy", true: "-eager"}[enabled], func(t *testing.T) {
				ctx := context.Background()
				f := createWitnessFixtureV23(t, true)
				defer f.close()
				s := attachLoadEagerV44(t, f, mode, enabled)
				defer s.closeV37()
				prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
				if err != nil {
					t.Fatal(err)
				}
				j, err := s.eager.authority.gate.store.GetBayesianJournal(ctx, "tenant-a", prime.BayesianShadow.JournalID)
				if err != nil || durableWireV42(j, prime) != nil {
					t.Fatal("prime full durable wire", err)
				}
				if enabled {
					if len(s.eager.preparations) != 1 || !s.eager.preparations[0].Built || s.eager.preparations[0].Kind != "journal" || !reflect.DeepEqual(s.eager.preparations[0].OperationIDs, []string{j.ID}) {
						t.Fatal("prime preparation trace")
					}
				} else if len(s.eager.preparations) != 0 {
					t.Fatal("disabled hidden preparation")
				}
				entered := make(chan struct{}, 1)
				resume := make(chan struct{})
				s.metadata.afterProjection = func(context.Context) { entered <- struct{}{}; <-resume }
				s.eager.authority.owner.Lock()
				held := true
				released := false
				release := func() {
					if held {
						s.eager.authority.owner.Unlock()
						held = false
					}
					if !released {
						close(resume)
						released = true
					}
				}
				type response struct {
					packet model.ContextPacket
					err    error
				}
				done := make(chan response, 1)
				go func() { p, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "held")); done <- response{p, e} }()
				terminal := false
				defer func() {
					release()
					if !terminal {
						select {
						case <-done:
						case <-time.After(5 * time.Second):
							t.Error("held request not drained")
						}
					}
				}()
				if enabled {
					select {
					case <-entered:
					case <-time.After(time.Second):
						t.Fatal("eager nomination blocked")
					}
				} else {
					select {
					case <-entered:
						t.Fatal("lazy unexpectedly warmed first head")
					case <-time.After(10 * time.Millisecond):
					}
				}
				select {
				case r := <-done:
					terminal = true
					t.Fatal("packet before full owner release", r.err)
				case <-time.After(10 * time.Millisecond):
				}
				release()
				select {
				case r := <-done:
					terminal = true
					if r.err != nil {
						t.Fatal(r.err)
					}
					j, e := s.eager.authority.gate.store.GetBayesianJournal(ctx, "tenant-a", r.packet.BayesianShadow.JournalID)
					if e != nil || durableWireV42(j, r.packet) != nil {
						t.Fatal("held durable wire", e)
					}
				case <-time.After(5 * time.Second):
					t.Fatal("held request unfinished")
				}
				s.metadata.afterProjection = nil
				if enabled {
					if len(s.eager.preparations) != 2 || !s.eager.preparations[1].Reused {
						t.Fatal("same-head preparation missing")
					}
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					if !errors.Is(s.eager.refresh(cancelled, "control"), context.Canceled) {
						t.Fatal("cancelled preparation accepted")
					}
					last := s.eager.preparations[len(s.eager.preparations)-1]
					if last.Owner || !last.OwnerAt.IsZero() || last.Built || last.Error == "" {
						t.Fatal("cancelled work attribution")
					}
					for _, trace := range s.eager.preparations[:len(s.eager.preparations)-1] {
						if !trace.Owner || trace.OwnerAt.Before(trace.Begin) || trace.OwnerAt.After(trace.End) || (trace.Built && trace.BuiltAt.Before(trace.OwnerAt)) {
							t.Fatal("owner-time attribution", trace)
						}
					}
				}
				s.eager.authority.poison.Store(true)
				if _, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "poison")); !errors.Is(e, store.ErrStaleSnapshot) {
					t.Fatal("poison delivered", e)
				}
				s.eager.authority.poison.Store(false)
			})
		}
	}
}
