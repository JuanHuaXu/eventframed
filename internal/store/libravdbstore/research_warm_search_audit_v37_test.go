package libravdbstore

import (
	"context"
	"errors"
	"go/ast"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
)

func TestResearchWarmGenerationV37(t *testing.T) {
	base := researchASTV34(t, "research_metadata_core_load_v36_generated_test.go", "runMetadataCoreLoadV36")
	fork := researchASTV34(t, "research_warm_search_load_v37_generated_test.go", "runWarmSearchLoadV37")
	back := map[string]string{"runWarmSearchLoadV37": "runMetadataCoreLoadV36", "warmTrialV37": "coreTrialV36", "attachLoadWarmV37": "attachLoadCoreV36", "finishWarmTrialV37": "finishCoreTrialV36", "closeV37": "closeV36"}
	ast.Inspect(fork, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if name, ok := back[id.Name]; ok {
				id.Name = name
			}
		}
		return true
	})
	native := func(f *ast.FuncDecl) string {
		found := ""
		ast.Inspect(f.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && researchExprV34(c.Fun) == "s.publishedRecallLoadStoreV16.Search" {
				if found != "" {
					t.Fatal("duplicate native Search")
				}
				found = researchExprV34(c)
			}
			return true
		})
		if found == "" {
			t.Fatal("native Search missing")
		}
		return found
	}
	if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("undeclared workload change")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("body corruption escaped")
	}
	base = researchASTV34(t, "research_durable_witness_v23_test.go", "Search")
	fork = researchASTV34(t, "research_warm_search_v37_test.go", "Search")
	ast.Inspect(fork, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && researchExprV34(sel.X) == "s.authority" {
			sel.X = ast.NewIdent("s")
		}
		return true
	})
	if native(base) != native(fork) {
		t.Fatal("native Search arguments changed")
	}
	binding := func(f *ast.FuncDecl) []ast.Stmt {
		start, end := -1, -1
		for i, stmt := range f.Body.List {
			if a, ok := stmt.(*ast.AssignStmt); ok && len(a.Lhs) == 1 {
				switch researchExprV34(a.Lhs[0]) {
				case "scope.Vector":
					start = i
				case "scope.Frontier":
					end = i
				}
			}
		}
		if start < 0 || end < start {
			t.Fatal("binding block missing")
		}
		return f.Body.List[start : end+1]
	}
	base.Body.List, fork.Body.List = binding(base), binding(fork)
	fork.Recv = base.Recv
	if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("sealed binding changed")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("binding corruption escaped")
	}
}

func TestResearchWarmControlsV37(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_WARM_SEARCH_V37") != "1" {
		t.Skip("isolated native warm controls")
	}
	ctx := context.Background()
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadWarmV37(t, f, mode)
			defer s.closeV37()
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil || s.shared.core.Load() != nil {
				t.Fatal("nil certificate cached", err)
			}
			if _, err = f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "first", false); err != nil {
				t.Fatal(err)
			}
			w := pinnedWrite("future", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err = s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			req := f.request(time.Now().UTC(), "prime-core")
			packet, _, err := s.recall(ctx, f.svc, req)
			if err != nil || witnessBeliefsV23(packet) == 0 {
				t.Fatal("future beliefs missing", err)
			}
			core := s.shared.core.Load()
			if core == nil {
				t.Fatal("no core")
			}
			var live context.Context
			s.metadata.afterProjection = func(c context.Context) {
				live = c
				l, le := s.warm.GetBayesianPosterior(c, "tenant-a", "past100")
				r, re := s.metadata.authority.GetBayesianPosterior(c, "tenant-a", "past100")
				if le != nil || re != nil || !reflect.DeepEqual(l, r) {
					t.Fatal("native differential", le, re)
				}
			}
			warm, _, err := s.recall(ctx, f.svc, req)
			if err != nil || s.warm.fast.Load() == 0 || !reflect.DeepEqual(packet.BayesianShadow.Decisions, warm.BayesianShadow.Decisions) {
				t.Fatal("warm law/nomination", err)
			}
			if _, _, err = s.metadata.pinV35(live); err == nil {
				t.Fatal("handoff token lives")
			}
			if _, err = s.warm.Search(live, "tenant-a", f.query, req.AsOf, 150); err == nil {
				t.Fatal("expired Search rearmed")
			}
			foreign := context.WithValue(live, metadataKeyV35{}, &metadataPinV35{owner: &projectionStoreV35{}})
			if _, err = s.warm.Search(foreign, "tenant-a", f.query, req.AsOf, 150); err == nil {
				t.Fatal("foreign Search accepted")
			}
			s.warm.enabled = false
			s.metadata.afterProjection = nil
			control, _, err := s.recall(ctx, f.svc, req)
			if err != nil || !reflect.DeepEqual(warm.BayesianShadow.Decisions, control.BayesianShadow.Decisions) {
				t.Fatal("sealed differential", err)
			}
			s.warm.enabled = true
			// Nomination is allowed with owner held; returned packet is not. The sealed
			// capture/commit still blocks and proves this is not an early-ack shortcut.
			entered, done := make(chan struct{}), make(chan error, 1)
			var once sync.Once
			s.metadata.afterProjection = func(c context.Context) { once.Do(func() { close(entered) }) }
			s.metadata.authority.owner.Lock()
			var unlocked sync.Once
			unlock := func() { unlocked.Do(s.metadata.authority.owner.Unlock) }
			defer unlock()
			go func() { _, _, e := s.recall(ctx, f.svc, req); done <- e }()
			select {
			case <-entered:
			case <-time.After(time.Second):
				unlock()
				t.Fatal("warm nomination blocked by owner")
			}
			select {
			case e := <-done:
				unlock()
				t.Fatal("packet before durable owner release", e)
			case <-time.After(10 * time.Millisecond):
			}
			unlock()
			select {
			case e := <-done:
				if e != nil {
					t.Fatal(e)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("durable packet unfinished")
			}
			s.metadata.afterProjection = nil
			for _, kind := range []string{"query", "vector", "selection", "old"} {
				r := f.request(time.Now().UTC(), "changed")
				switch kind {
				case "query":
					r.Query = "different"
				case "vector":
					r.Embedding = denseRowV6(f.query, 11, .1)
				case "selection":
					r.PackK = 9
				case "old":
					r.AsOf = originalSourceAtV35(t, s.loadProjectionV35).Add(-time.Nanosecond)
				}
				p, _, e := s.recall(ctx, f.svc, r)
				if e != nil || witnessBeliefsV23(p) != 0 {
					t.Fatal("overreach", kind, e)
				}
			}
			w = pinnedWrite("visible", f.origin)
			w.Vector = denseRowV6(f.query, 3, .0002)
			if _, err = s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			visible, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "visible"))
			if err != nil || witnessBeliefsV23(visible) != 0 || s.shared.core.Load() == core {
				t.Fatal("motion/visible overreach", err)
			}
			s.metadata.authority.poison.Store(true)
			if _, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "poison")); !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatal("poison delivered", err)
			}
			s.metadata.authority.poison.Store(false)
			if err = s.closeV37(); err != nil {
				t.Fatal(err)
			}
			f.reopen(t)
			s = attachLoadWarmV37(t, f, mode)
			defer s.closeV37()
			if _, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen")); err != nil {
				t.Fatal(err)
			}
			w = pinnedWrite("gap", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 8, .0001)
			if err = appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); err != nil {
				t.Fatal(err)
			}
			if _, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatal("gap delivered", err)
			}
			t.Log("warm native/nomination/law, owner-held stage without early ack, cold/motion/expiry/request/poison/gap/reopen PASS")
		})
	}
}

func TestResearchWarmCancellationV37(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_WARM_SEARCH_V37") != "1" {
		t.Skip("isolated warm handoff cancellation")
	}
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadWarmV37(t, f, mode)
			defer s.closeV37()
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			defer unblock()
			barrier := func() { close(entered); <-release }
			if s.archive != nil {
				s.archive.barrier = barrier
			} else {
				s.joinedWitnessV25.barrier = barrier
			}
			var live context.Context
			s.metadata.afterProjection = func(c context.Context) { live = c }
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, _, e := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "held")); done <- e }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("handoff absent")
			}
			if _, _, e := s.metadata.pinV35(live); e == nil {
				t.Fatal("token not expired")
			}
			cancel()
			closing := make(chan error, 1)
			go func() { closing <- s.closeV37() }()
			select {
			case e := <-closing:
				t.Fatal("close before durability", e)
			case <-time.After(5 * time.Millisecond):
			}
			select {
			case e := <-done:
				t.Fatal("canceled caller returned early", e)
			default:
			}
			unblock()
			select {
			case e := <-done:
				if !errors.Is(e, context.Canceled) {
					t.Fatal(e)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("caller unfinished")
			}
			select {
			case e := <-closing:
				if e != nil {
					t.Fatal(e)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("close unfinished")
			}
			if len(s.state.Journals) != 1 {
				t.Fatal("accepted journal lost")
			}
		})
	}
}
