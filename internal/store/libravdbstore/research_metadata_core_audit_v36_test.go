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

func TestResearchMetadataCoreGenerationV36(t *testing.T) {
	base := researchASTV34(t, "research_metadata_projection_load_v35_generated_test.go", "runMetadataProjectionLoadV35")
	fork := researchASTV34(t, "research_metadata_core_load_v36_generated_test.go", "runMetadataCoreLoadV36")
	back := map[string]string{"runMetadataCoreLoadV36": "runMetadataProjectionLoadV35", "coreTrialV36": "projectionTrialV35", "attachLoadCoreV36": "attachLoadProjectionV35", "finishCoreTrialV36": "finishProjectionTrialV35", "closeV36": "closeV35"}
	ast.Inspect(fork, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if name, ok := back[id.Name]; ok {
				id.Name = name
			}
		}
		return true
	})
	if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("undeclared load change")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
		t.Fatal("body corruption escaped")
	}
}

func TestResearchMetadataCoreControlsV36(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_METADATA_CORE_V36") != "1" {
		t.Skip("isolated native head-core controls")
	}
	ctx := context.Background()
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadCoreV36(t, f, mode)
			defer s.closeV36()
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
			var live context.Context
			s.metadata.afterProjection = func(c context.Context) {
				live = c
				l, le := s.shared.GetBayesianPosterior(c, "tenant-a", "past100")
				r, re := s.metadata.authority.GetBayesianPosterior(c, "tenant-a", "past100")
				if le != nil || re != nil || !reflect.DeepEqual(l, r) {
					t.Fatal("native differential", le, re)
				}
				lc, le := s.shared.GetSelectionCertificate(c, "tenant-a")
				rc, re := s.metadata.authority.GetSelectionCertificate(c, "tenant-a")
				if le != nil || re != nil || !reflect.DeepEqual(lc, rc) {
					t.Fatal("selection differential")
				}
				lo, le := s.shared.GetOmittedInfluenceCertificate(c, "tenant-a")
				ro, re := s.metadata.authority.GetOmittedInfluenceCertificate(c, "tenant-a")
				if le != nil || re != nil || !reflect.DeepEqual(lo, ro) {
					t.Fatal("omitted differential")
				}
			}
			req := f.request(time.Now().UTC(), "cached")
			packet, _, err := s.recall(ctx, f.svc, req)
			if err != nil || witnessBeliefsV23(packet) == 0 {
				t.Fatal("future beliefs missing", err)
			}
			core := s.shared.core.Load()
			bytes := s.shared.physicalBytes.Load()
			if core == nil || core.state.Certificate == nil || core.state.Hash == s.state.Hash {
				t.Fatal("core not historical after journal")
			}
			if _, _, err = s.metadata.pinV35(live); err == nil {
				t.Fatal("token outlived handoff")
			}
			if _, _, err = s.recall(ctx, f.svc, req); err != nil || s.shared.core.Load() != core || s.shared.physicalBytes.Load() != bytes || s.shared.hits.Load() == 0 {
				t.Fatal("same-head copy repeated", err)
			}
			s.metadata.afterProjection = nil
			s.shared.enabled = false
			control, _, err := s.recall(ctx, f.svc, req)
			if err != nil || !reflect.DeepEqual(packet.BayesianShadow.Decisions, control.BayesianShadow.Decisions) {
				t.Fatal("law/decision differential", err)
			}
			s.shared.enabled = true
			// Mutate an independent decoder, never the published core itself.
			isolated := metadataDecodedV35(t, metadataProofV35{Encoded: core.encoded})
			x := isolated.Sources["past100"]
			x.Posterior.Alpha++
			isolated.Sources["past100"] = x
			if core.state.Sources["past100"].Posterior.Alpha == x.Posterior.Alpha || s.state.Sources["past100"].Posterior.Alpha == x.Posterior.Alpha {
				t.Fatal("decoder aliases shared/durable maps")
			}
			var wg sync.WaitGroup
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if _, _, e := s.recall(ctx, f.svc, req); e != nil {
						t.Error(e)
					}
				}()
			}
			wg.Wait()
			if s.shared.core.Load() != core || s.shared.physicalBytes.Load() != bytes {
				t.Fatal("concurrent warm core changed")
			}
			for _, kind := range []string{"query", "vector", "selection", "old"} {
				r := f.request(time.Now().UTC(), "invalid")
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
					t.Fatal("request overreach", kind, e)
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
			if err = s.closeV36(); err != nil {
				t.Fatal(err)
			}
			f.reopen(t)
			s = attachLoadCoreV36(t, f, mode)
			defer s.closeV36()
			if _, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "reopen")); err != nil {
				t.Fatal(err)
			}
			w = pinnedWrite("gap", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 8, .0001)
			if err = appendDenseV6(ctx, s.gate, []ResearchEventWrite{w}); err != nil {
				t.Fatal(err)
			}
			if _, _, err = s.recall(ctx, f.svc, f.request(time.Now().UTC(), "gap")); !errors.Is(err, store.ErrStaleSnapshot) {
				t.Fatal("unaccounted motion delivered", err)
			}
			t.Log("nil transition, native differential, historical root/head reuse, concurrent immutability, request/as-of controls, motion, reopen and gap PASS")
		})
	}
}

func TestResearchMetadataCoreCancellationV36(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_METADATA_CORE_V36") != "1" {
		t.Skip("isolated core handoff cancellation")
	}
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadCoreV36(t, f, mode)
			defer s.closeV36()
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
			go func() { _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "held")); done <- err }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("handoff not accepted")
			}
			if _, _, err := s.metadata.pinV35(live); err == nil {
				t.Fatal("token live after handoff")
			}
			cancel()
			closing := make(chan error, 1)
			go func() { closing <- s.closeV36() }()
			select {
			case e := <-closing:
				t.Fatal("close before durability", e)
			case <-time.After(5 * time.Millisecond):
			}
			select {
			case e := <-done:
				t.Fatal("canceled handoff returned early", e)
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
				t.Fatal("accepted canceled journal lost")
			}
		})
	}
}

var _ store.EventStore = (*coreStoreV36)(nil)
