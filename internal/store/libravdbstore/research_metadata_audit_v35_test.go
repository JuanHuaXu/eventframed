package libravdbstore

import (
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/store"
)

func TestResearchMetadataGenerationV35(t *testing.T) {
	for _, name := range []string{"GetSelectionCertificate", "GetOmittedInfluenceCertificate", "GetBayesianPosterior", "compatibleV23", "runBatchArchiveLoadV34"} {
		basePath, forkPath, forkName := "research_durable_witness_v23_test.go", "research_metadata_getters_v35_generated_test.go", name
		if name == "compatibleV23" {
			forkPath, forkName = "research_metadata_projection_v35_test.go", "compatibleV35"
		}
		if name == "runBatchArchiveLoadV34" {
			basePath, forkPath, forkName = "research_archive_cost_load_v34_generated_test.go", "research_metadata_projection_load_v35_generated_test.go", "runMetadataProjectionLoadV35"
		}
		base := researchASTV34(t, basePath, name)
		fork := researchASTV34(t, forkPath, forkName)
		switch name {
		case "runBatchArchiveLoadV34":
			back := map[string]string{"runMetadataProjectionLoadV35": "runBatchArchiveLoadV34", "projectionTrialV35": "archiveTrialV34", "attachLoadProjectionV35": "attachLoadArchiveV34", "closeV35": "closeV34", "finishProjectionTrialV35": "finishArchiveTrialV34"}
			ast.Inspect(fork, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					if v, ok := back[id.Name]; ok {
						id.Name = v
					}
				}
				return true
			})
		default:
			fork.Recv.List[0].Type = &ast.StarExpr{X: ast.NewIdent("witnessStoreV23")}
			if name == "compatibleV23" {
				fork.Name.Name = name
				fork.Type.Params.List = append(fork.Type.Params.List[:1], fork.Type.Params.List[2:]...)
			} else {
				if len(fork.Body.List) < 4 {
					t.Fatal("projection prefix missing")
				}
				first, ok := fork.Body.List[0].(*ast.AssignStmt)
				if !ok || len(first.Rhs) != 1 {
					t.Fatal("projection initializer missing")
				}
				if call, ok := first.Rhs[0].(*ast.CallExpr); !ok || researchExprV34(call.Fun) != "s.pinV35" {
					t.Fatal("wrong projection initializer")
				}
				fork.Body.List = fork.Body.List[4:]
				var kept []ast.Stmt
				counter := 0
				for _, stmt := range fork.Body.List {
					if x, ok := stmt.(*ast.ExprStmt); ok {
						if c, ok := x.X.(*ast.CallExpr); ok && researchExprV34(c.Fun) == "s.traceGetterV35" {
							continue
						}
					}
					if x, ok := stmt.(*ast.ExprStmt); ok && researchExprV34(x.X) == "s.nativeSuccess.Add(1)" {
						lock, err := parser.ParseFile(token.NewFileSet(), "snippet", "package p\nfunc f(){s.owner.Lock();defer s.owner.Unlock()}", 0)
						if err != nil {
							t.Fatal(err)
						}
						kept = append(kept, lock.Decls[0].(*ast.FuncDecl).Body.List...)
						counter++
						continue
					}
					kept = append(kept, stmt)
				}
				if counter != 1 {
					t.Fatal("native success marker", name, counter)
				}
				fork.Body.List = kept
			}
			ast.Inspect(fork.Body, func(n ast.Node) bool {
				if sel, ok := n.(*ast.SelectorExpr); ok {
					if researchExprV34(sel.X) == "s.authority" {
						sel.X = ast.NewIdent("s")
					}
					if researchExprV34(sel.X) == "pin" && sel.Sel.Name == "state" {
						sel.X = ast.NewIdent("s")
					}
				}
				if c, ok := n.(*ast.CallExpr); ok && researchExprV34(c.Fun) == "s.compatibleV35" {
					c.Fun = &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("compatibleV23")}
					c.Args = append(c.Args[:1], c.Args[2:]...)
				}
				return true
			})
		}
		if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
			t.Fatal("undeclared getter/validity/workload change", name)
		}
		fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
		if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
			t.Fatal("body corruption escaped", name)
		}
	}
}

func TestResearchMetadataCoverageV35(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "../store.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == "EventStore" {
			for _, m := range ts.Type.(*ast.InterfaceType).Methods.List {
				want[m.Names[0].Name] = true
			}
		}
		return true
	})
	classes := map[string]string{}
	for group, names := range map[string][]string{
		"projected":                    []string{"Search", "GetSelectionCertificate", "GetOmittedInfluenceCertificate", "GetBayesianPosterior"},
		"expiry-boundary":              []string{"PutBayesianJournal"},
		"sealed-native-read":           []string{"GetCompositionTombstone", "GetEvents", "GetBayesianJournal", "GetAntiPigeonCertificate", "GetResidualCandidates", "GetPredictiveGraph", "Stats", "Snapshot"},
		"sealed-maintenance":           []string{"Backup"},
		"sealed-mutation-gap-rejected": []string{"BindBayesianPolicy", "Put", "PutComposition", "Delete", "DeleteComposition", "DeleteBefore", "Compact", "PublishSelectionCertificate", "PublishAntiPigeonCertificate", "PublishOmittedInfluenceCertificate", "ApplyBayesianOutcome", "PublishPredictiveSnap", "RollbackPredictiveSnap", "PutAgencyProposal", "ClaimAgencyProposals", "ResolveAgencyProposal"},
		"sealed-lifecycle":             []string{"Close"},
	} {
		for _, name := range names {
			if classes[name] != "" {
				t.Fatal("duplicate classification", name)
			}
			classes[name] = group
		}
	}
	complete := func() bool {
		if len(classes) != len(want) {
			return false
		}
		for name := range want {
			if classes[name] == "" {
				return false
			}
		}
		return true
	}
	if len(want) != 31 || !complete() {
		t.Fatal("classification gap", len(want), len(classes))
	}
	delete(classes, "Snapshot")
	if complete() {
		t.Fatal("missing Snapshot escaped")
	}
	t.Log("all31methods classified; not a claim that arbitrary delegated mutations have projection authority")
}

func TestResearchMetadataControlsV35(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_METADATA_PROJECTION_V35") != "1" {
		t.Skip("isolated native projection controls")
	}
	ctx := context.Background()
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadProjectionV35(t, f, mode)
			defer s.closeV35()
			prime, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "prime"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err = f.feedback(ctx, prime.BayesianShadow.JournalID, "past100", "first", false); err != nil {
				t.Fatal(err)
			}
			w := pinnedWrite("future", f.origin.Add(time.Hour))
			w.Vector = denseRowV6(f.query, 7, .0003)
			if _, err = s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			var leaked context.Context
			s.metadata.afterProjection = func(live context.Context) {
				leaked = live
				p, _, err := s.metadata.pinV35(live)
				if err != nil {
					t.Fatal(err)
				}
				left, err := s.metadata.GetBayesianPosterior(live, "tenant-a", "past100")
				if err != nil {
					t.Fatal(err)
				}
				right, err := s.metadata.authority.GetBayesianPosterior(live, "tenant-a", "past100")
				if err != nil || !reflect.DeepEqual(left, right) {
					t.Fatal("posterior differential", err)
				}
				lc, le := s.metadata.GetSelectionCertificate(live, "tenant-a")
				rc, re := s.metadata.authority.GetSelectionCertificate(live, "tenant-a")
				if le != nil || re != nil || !reflect.DeepEqual(lc, rc) {
					t.Fatal("selection differential")
				}
				lo, le := s.metadata.GetOmittedInfluenceCertificate(live, "tenant-a")
				ro, re := s.metadata.authority.GetOmittedInfluenceCertificate(live, "tenant-a")
				if le != nil || re != nil || !reflect.DeepEqual(lo, ro) {
					t.Fatal("omitted differential")
				}
				// Mutating our private copy must neither alias durable witness state nor
				// bypass full native equality. Current epoch transport must disappear.
				original := p.state.Sources["past100"]
				tampered := original
				tampered.Posterior.Alpha++
				p.state.Sources["past100"] = tampered
				bad, e := s.metadata.GetBayesianPosterior(live, "tenant-a", "past100")
				if e != nil || bad.EvidenceEpoch == p.state.Head.EvidenceEpoch || s.metadata.authority.state.Sources["past100"].Posterior.Alpha != original.Posterior.Alpha {
					t.Fatal("native inequality or copy ownership bypassed", e)
				}
				p.state.Sources["past100"] = original
				// Getter remains callable while journal/publication owner is held.
				s.metadata.authority.owner.Lock()
				var once sync.Once
				unlock := func() { once.Do(s.metadata.authority.owner.Unlock) }
				defer unlock()
				done := make(chan error, 1)
				go func() { _, e := s.metadata.GetBayesianPosterior(live, "tenant-a", "past100"); done <- e }()
				select {
				case e := <-done:
					if e != nil {
						t.Fatal(e)
					}
				case <-time.After(time.Second):
					unlock()
					<-done
					t.Fatal("projected getter acquired owner")
				}
				unlock()
			}
			r := f.request(time.Now().UTC(), "learned")
			projected, _, err := s.recall(ctx, f.svc, r)
			if err != nil || witnessBeliefsV23(projected) == 0 {
				t.Fatal("positive projected law", err)
			}
			s.metadata.afterProjection = nil
			if _, err = s.metadata.GetBayesianPosterior(leaked, "tenant-a", "past100"); err == nil {
				t.Fatal("projection survived journal boundary")
			}
			p := leaked.Value(metadataKeyV35{}).(*metadataPinV35)
			oldOwner := p.owner
			p.owner = &projectionStoreV35{}
			p.active.Store(true)
			if _, _, err = s.metadata.pinV35(leaked); err == nil {
				t.Fatal("foreign projection accepted")
			}
			p.owner = oldOwner
			p.active.Store(false)
			s.metadata.projected = false
			control, _, err := s.recall(ctx, f.svc, r)
			if err != nil || !reflect.DeepEqual(projected.BayesianShadow.Decisions, control.BayesianShadow.Decisions) {
				t.Fatal("law/selection differential", err)
			}
			s.metadata.projected = true
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
					r.AsOf = originalSourceAtV35(t, s).Add(-time.Nanosecond)
				}
				packet, _, err := s.recall(ctx, f.svc, r)
				if err != nil || witnessBeliefsV23(packet) != 0 {
					t.Fatal("request overreach", kind, err)
				}
			}
			w = pinnedWrite("visible", f.origin)
			w.Vector = denseRowV6(f.query, 3, .0002)
			if _, err = s.append(ctx, []ResearchEventWrite{w}, false); err != nil {
				t.Fatal(err)
			}
			visible, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "visible"))
			if err != nil || witnessBeliefsV23(visible) != 0 {
				t.Fatal("visible overreach", err)
			}
			if err = s.closeV35(); err != nil {
				t.Fatal(err)
			}
			f.reopen(t)
			s = attachLoadProjectionV35(t, f, mode)
			defer s.closeV35()
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
			t.Log("native differential, owner-held lookup, copy isolation, native inequality, expiry/foreign, future/visible/request controls, reopen and gap PASS")
		})
	}
}
func originalSourceAtV35(t *testing.T, s *loadProjectionV35) time.Time {
	t.Helper()
	p, err := s.gate.store.GetBayesianPosterior(context.Background(), "tenant-a", "past100")
	if err != nil {
		t.Fatal(err)
	}
	return p.UpdatedAt
}

func TestResearchMetadataCancellationV35(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_METADATA_PROJECTION_V35") != "1" {
		t.Skip("isolated projected handoff cancellation")
	}
	for _, mode := range []int{-2, 4} {
		t.Run(map[int]string{-2: "joined", 4: "archive"}[mode], func(t *testing.T) {
			f := createWitnessFixtureV23(t, true)
			defer f.close()
			s := attachLoadProjectionV35(t, f, mode)
			defer s.closeV35()
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
			var leaked context.Context
			s.metadata.afterProjection = func(ctx context.Context) { leaked = ctx }
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "held")); done <- err }()
			select {
			case <-entered:
			case <-time.After(5 * time.Second):
				t.Fatal("handoff not accepted")
			}
			if _, _, err := s.metadata.pinV35(leaked); err == nil {
				t.Fatal("projection live after forwarded handoff")
			}
			cancel()
			closing := make(chan error, 1)
			go func() { closing <- s.closeV35() }()
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
				t.Fatal("caller did not finish")
			}
			select {
			case e := <-closing:
				if e != nil {
					t.Fatal(e)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("close did not finish")
			}
			if len(s.state.Journals) != 1 {
				t.Fatal("accepted canceled journal lost")
			}
			t.Log("projection expired before accepted handoff; canceled caller and Close waited for durable terminal")
		})
	}
}

// Keep the saved owned wire readable without trusting projected convenience fields.
func metadataDecodedV35(t *testing.T, p metadataProofV35) *witnessStateV23 {
	t.Helper()
	var s witnessStateV23
	if err := json.Unmarshal([]byte(p.Encoded), &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

var _ store.EventStore = (*projectionStoreV35)(nil)
