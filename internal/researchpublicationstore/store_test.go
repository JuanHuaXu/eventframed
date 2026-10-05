package researchpublicationstore

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

var generalMethods = strings.Fields("BindBayesianPolicy PutComposition Delete DeleteComposition DeleteBefore Backup Compact PublishSelectionCertificate PublishAntiPigeonCertificate PublishOmittedInfluenceCertificate ApplyBayesianOutcome PublishPredictiveSnap RollbackPredictiveSnap PutAgencyProposal ClaimAgencyProposals ResolveAgencyProposal")
var readMethods = strings.Fields("GetCompositionTombstone Search GetEvents GetBayesianJournal GetSelectionCertificate GetAntiPigeonCertificate GetOmittedInfluenceCertificate GetBayesianPosterior GetResidualCandidates GetPredictiveGraph Stats Snapshot")

func TestEveryCoreMethodClassifiedAndGuarded(t *testing.T) {
	classes := map[string]string{"Put": "Ingestion", "Close": "lifecycle", "PutBayesianJournal": "journal"}
	for _, n := range generalMethods {
		classes[n] = "General"
	}
	for _, n := range readMethods {
		classes[n] = "read"
	}
	typ := reflect.TypeOf((*store.EventStore)(nil)).Elem()
	if typ.NumMethod() != len(classes) {
		t.Fatal("store interface changed; audit publication coverage")
	}
	for i := 0; i < typ.NumMethod(); i++ {
		if classes[typ.Method(i).Name] == "" {
			t.Fatal("unclassified interface method", typ.Method(i).Name)
		}
	}
	methods := map[string]*ast.FuncDecl{}
	for _, name := range []string{"store.go", "mutations.go"} {
		f, e := parser.ParseFile(token.NewFileSet(), name, nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv != nil {
				methods[fn.Name.Name] = fn
			}
		}
	}
	for n, class := range classes {
		if class == "read" {
			continue
		}
		fn := methods[n]
		if fn == nil {
			t.Fatal("mutation inherited without wrapper", n)
		}
		if class != "General" && class != "Ingestion" {
			continue
		}
		guarded := false
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || (id.Name != "mutate" && id.Name != "mutateWithTouches") || len(call.Args) < 2 {
				return true
			}
			kind, ok := call.Args[1].(*ast.SelectorExpr)
			guarded = guarded || (ok && kind.Sel.Name == class)
			return true
		})
		if !guarded {
			t.Fatal("wrong mutation guard", n, class)
		}
	}
}

type panicBackend struct {
	store.EventStore
	snapshot model.Snapshot
}

func (b *panicBackend) Snapshot(context.Context) model.Snapshot { return b.snapshot }

func TestEveryWriterQuarantinesBackendPanic(t *testing.T) {
	for _, name := range append(append([]string{}, generalMethods...), "Put", "Close") {
		t.Run(name, func(t *testing.T) {
			base := model.Snapshot{RuntimeVersion: 1, EvidenceEpoch: 1}
			s, e := New(&panicBackend{snapshot: base})
			if e != nil {
				t.Fatal(e)
			}
			method := reflect.ValueOf(s).MethodByName(name)
			args := make([]reflect.Value, method.Type().NumIn())
			for i := range args {
				typ := method.Type().In(i)
				args[i] = reflect.Zero(typ)
				if typ == reflect.TypeOf((*context.Context)(nil)).Elem() {
					args[i] = reflect.ValueOf(context.Background())
				}
				if typ == reflect.TypeOf(model.Event{}) {
					args[i] = reflect.ValueOf(model.Event{AvailableAt: time.Now().Add(time.Hour)})
				}
			}
			panicked := false
			func() { defer func() { panicked = recover() != nil }(); method.Call(args) }()
			if !panicked || s.ResearchPublicationCompatible(context.Background(), base, time.Now()) {
				t.Fatal("backend panic escaped quarantine")
			}
		})
	}
}

func TestRealStoreTransitions(t *testing.T) {
	for _, persistent := range []bool{false, true} {
		name := "memory"
		if persistent {
			name = "persistent"
		}
		t.Run(name, func(t *testing.T) {
			var backend store.EventStore = memorystore.New()
			var e error
			if persistent {
				backend, e = libravdbstore.Open(libravdbstore.Config{Path: t.TempDir() + "/publication.libravdb", Dimension: 8, EmbeddingModel: "fixture", Quantization: "none", MemoryMapping: true})
				if e != nil {
					t.Fatal(e)
				}
			}
			s, e := New(backend)
			if e != nil {
				t.Fatal(e)
			}
			defer s.Close()
			ctx := context.Background()
			now := time.Now()
			old := s.Snapshot(ctx)
			current, e := s.BindBayesianPolicy(ctx, "policy-v1")
			if e != nil {
				t.Fatal(e)
			}
			if s.ResearchPublicationCompatible(ctx, old, now) || !s.ResearchPublicationCompatible(ctx, current, now) {
				t.Fatal("policy publication mismatch")
			}
			ev := testutil.Event("future-fixture", "public fixture", now.Add(time.Hour))
			r, e := s.Put(ctx, ev, make([]float32, 8), "digest")
			if e != nil {
				t.Fatal(e)
			}
			if !s.ResearchPublicationCompatible(ctx, current, now) {
				t.Fatal("future ingestion lost proof")
			}
			if _, e = s.Delete(ctx, ev.TenantID, ev.ID); e != nil {
				t.Fatal(e)
			}
			if s.ResearchPublicationCompatible(ctx, r.Snapshot, now) {
				t.Fatal("deletion failed invalidation")
			}
			if _, e = s.Put(ctx, ev, make([]float32, 8), "digest"); e != nil {
				t.Fatal(e)
			}
			if _, e = s.Put(ctx, ev, make([]float32, 8), "digest"); e == nil {
				t.Fatal("ambiguous duplicate was silently accepted")
			}
			if s.ResearchPublicationCompatible(ctx, s.Snapshot(ctx), now) {
				t.Fatal("duplicate did not quarantine")
			}
		})
	}
}
