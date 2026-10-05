package libravdbstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestResearchCohortSourcePreservationV46(t *testing.T) {
	data, err := os.ReadFile("../../../research/cohort-batch-v46-generation.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Renames   map[string]string
		Functions []struct {
			Source, Name, Fork, Receiver string
			Worker                       bool
			Native                       bool
		}
		WorkerBoundsChanged int
		NativeBoundsChanged int
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.WorkerBoundsChanged != 2 || manifest.NativeBoundsChanged != 2 || len(manifest.Functions) < 58 {
		t.Fatal("incomplete closure", len(manifest.Functions))
	}
	inverse := map[string]string{}
	for a, b := range manifest.Renames {
		inverse[b] = a
	}
	fork, err := parser.ParseFile(token.NewFileSet(), "research_cohort_batch_v46_generated_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(fork, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if old, ok := inverse[id.Name]; ok {
				id.Name = old
			}
		}
		return true
	})
	changed := 0
	nativeChanged := 0
	for _, record := range manifest.Functions {
		base := researchASTV34(t, record.Source, record.Name)
		var actual *ast.FuncDecl
		for _, d := range fork.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == record.Name {
				if record.Receiver == "" && fn.Recv == nil || record.Receiver != "" && fn.Recv != nil && researchExprV34(fn.Recv.List[0].Type) == record.Receiver {
					if actual != nil {
						t.Fatal("duplicate function", record)
					}
					actual = fn
				}
			}
		}
		if actual == nil {
			t.Fatal("missing function", record)
		}
		if record.Worker {
			count := 0
			ast.Inspect(actual, func(n ast.Node) bool {
				if b, ok := n.(*ast.BinaryExpr); ok && b.Op == token.LSS && researchExprV34(b.X) == "len(batch)" {
					if lit, ok := b.Y.(*ast.BasicLit); ok && lit.Value == "8" {
						lit.Value = "4"
						count++
					}
				}
				return true
			})
			if count != 1 {
				t.Fatal("undeclared worker change", record, count)
			}
			changed += count
		}
		if record.Native {
			count, messages := 0, 0
			ast.Inspect(actual, func(n ast.Node) bool {
				if b, ok := n.(*ast.BinaryExpr); ok && b.Op == token.GTR && researchExprV34(b.X) == "len(entries)" {
					if lit, ok := b.Y.(*ast.BasicLit); ok && lit.Value == "8" {
						lit.Value = "4"
						count++
					}
				}
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == `"journal batch size must be in [1,8]"` {
					lit.Value = `"journal batch size must be in [1,4]"`
					messages++
				}
				return true
			})
			if count != 1 || messages != 1 {
				t.Fatal("undeclared native change", record, count, messages)
			}
			nativeChanged += count
		}
		if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(actual)) {
			t.Fatal("undeclared behavior change", record)
		}
		// Negative control: deleting a statement must not remain equivalent.
		if len(actual.Body.List) > 0 {
			bad := *actual
			body := *actual.Body
			bad.Body = &body
			bad.Body.List = bad.Body.List[:len(bad.Body.List)-1]
			if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(&bad)) {
				t.Fatal("deletion escaped", record)
			}
		}
	}
	if changed != 2 || nativeChanged != 2 {
		t.Fatal("worker coverage", changed)
	}
	t.Logf("independently normalized all %d functions; exactly two worker and two native 4->8 bounds", len(manifest.Functions))
}

func TestResearchCohortEightJournalDurabilityV46(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_JOINED_WITNESS_V25") != "1" {
		t.Skip("isolated native eight-entry regression")
	}
	ctx := context.Background()
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachJoinedV25_CohortV46(t, f, true)
	packet, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "eight-prime"))
	if err != nil {
		t.Fatal(err)
	}
	entry, err := s.gate.store.GetBayesianJournal(ctx, "tenant-a", packet.BayesianShadow.JournalID)
	if err != nil {
		t.Fatal(err)
	}
	jobs := make([]*joinedJobV25, 8)
	entries := make([]model.BayesianJournalEntry, 8)
	for i := range jobs {
		e := entry
		e.ID = fmt.Sprintf("eight-commit-%d", i)
		e.Report.JournalID = e.ID
		entries[i] = e
		jobs[i] = &joinedJobV25{entry: e, binding: s.state.Journals[entry.ID]}
	}
	if err := s.gate.appendJoinedJournalV25(ctx, entries, "", 0, nil); err == nil {
		t.Fatal("original four-entry contract silently widened")
	}
	if err := s.gate.appendArchiveJournalV29(ctx, entries, "", 0, nil, nil); err == nil {
		t.Fatal("original archive four-entry contract silently widened")
	}
	if err := s.apply(jobs); err != nil {
		t.Fatal("eight-entry full commit rejected", err)
	}
	for _, e := range entries {
		actual, err := s.gate.store.GetBayesianJournal(ctx, e.TenantID, e.ID)
		if err != nil || !reflect.DeepEqual(actual, e) {
			t.Fatal("lost original durable wire", e.ID, err)
		}
	}
	if len(s.state.Journals) != 9 || s.poison.Load() || s.current.Load() == nil {
		t.Fatal("incomplete eight-entry publication")
	}
	nine := append(append([]model.BayesianJournalEntry(nil), entries...), entry)
	if err := s.gate.appendJoinedJournalV25_CohortV46(ctx, nine, "", 0, nil); err == nil {
		t.Fatal("unbounded joined native cap")
	}
	if err := s.gate.appendArchiveJournalV29_CohortV46(ctx, nine, "", 0, nil, nil); err == nil {
		t.Fatal("unbounded archive native cap")
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	f.reopen(t)
	restored := attachJoinedV25_CohortV46(t, f, true)
	defer restored.Close()
	if restored.poison.Load() || len(restored.state.Journals) != 9 {
		t.Fatal("eight-entry durability lost on reopen")
	}
}

// The complete old workload runs in both arms; only cohort commit capacity differs.
func TestResearchCohortBatchLoadV46(t *testing.T) {
	output := os.Getenv("EVENTFRAME_COHORT_BATCH_V46_OUTPUT")
	if output == "" {
		t.Skip("exclusive isolated mixed load")
	}
	freeze := os.Getenv("EVENTFRAME_COHORT_BATCH_V46_FREEZE")
	if freeze == "" {
		freeze = "../../../research/cohort-batch-v46/freeze.json"
	}
	manifest, err := os.ReadFile(freeze)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.Sum256(manifest)
	f, err := os.OpenFile(output, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"Type": "header", "Trials": 32, "FreezeSHA256": hex.EncodeToString(h[:]), "Time": time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	for trial := 1; trial <= 2; trial++ {
		for _, visible := range []bool{false, true} {
			for _, mode := range []int{-2, 4} {
				for _, eager := range []bool{false, true} {
					caps := []int{4, 8}
					if trial == 2 {
						caps = []int{8, 4}
					}
					for _, cap := range caps {
						var value any
						if cap == 4 {
							value = runEagerLoadV44(t, trial, mode, visible, eager)
						} else {
							value = runEagerLoadV44_CohortV46(t, trial, mode, visible, eager)
						}
						b, err := json.Marshal(value)
						if err != nil {
							t.Fatal(err)
						}
						var row map[string]json.RawMessage
						if err := json.Unmarshal(b, &row); err != nil {
							t.Fatal(err)
						}
						row["BatchCap"] = json.RawMessage([]byte("4"))
						if cap == 8 {
							row["BatchCap"] = json.RawMessage([]byte("8"))
						}
						if err := enc.Encode(row); err != nil {
							t.Fatal(err)
						}
						if err := f.Sync(); err != nil {
							t.Fatal(err)
						}
						t.Logf("trial=%d cap=%d mode=%d visible=%v eager=%v pass=%s metrics=%s", trial, cap, mode, visible, eager, row["Pass"], row["Metrics"])
					}
				}
			}
		}
	}
	if err := enc.Encode(map[string]any{"Type": "footer", "Trials": 32}); err != nil {
		t.Fatal(err)
	}
	if err := f.Sync(); err != nil {
		t.Fatal(err)
	}
}
