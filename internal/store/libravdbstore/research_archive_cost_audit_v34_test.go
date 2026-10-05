package libravdbstore

import (
	"bytes"
	"context"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func researchASTV34(t *testing.T, path, name string) *ast.FuncDecl {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	t.Fatal("missing function", name)
	return nil
}
func researchExprV34(n ast.Node) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), n); err != nil {
		panic(err)
	}
	return b.String()
}
func researchTokensV34(n ast.Node) []string {
	data := []byte(researchExprV34(n))
	fset := token.NewFileSet()
	file := fset.AddFile("fn", -1, len(data))
	var s scanner.Scanner
	s.Init(file, data, nil, 0)
	var tokens []string
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			lit = ";"
		}
		tokens = append(tokens, tok.String()+":"+lit)
	}
	return tokens
}

// Independent inverse comparison covers the ENTIRE apply, worker, attachment,
// and offered workload bodies, not only matching a favorable gate expression.
func TestResearchArchiveCostGenerationV34(t *testing.T) {
	for _, name := range []string{"apply", "worker", "attachBatchArchive", "runBatchArchiveLoad"} {
		basePath, forkPath := "research_batch_archive_v33_test.go", "research_archive_cost_v34_generated_test.go"
		if name == "runBatchArchiveLoad" {
			basePath, forkPath = "research_batch_archive_load_v33_generated_test.go", "research_archive_cost_load_v34_generated_test.go"
		}
		base := researchASTV34(t, basePath, name+"V33")
		fork := researchASTV34(t, forkPath, name+"V34")
		ast.Inspect(fork, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && strings.HasSuffix(id.Name, "V34") {
				id.Name = strings.TrimSuffix(id.Name, "V34") + "V33"
			}
			return true
		})
		switch name {
		case "apply":
			var kept []ast.Stmt
			removed := 0
			for _, stmt := range fork.Body.List {
				if a, ok := stmt.(*ast.AssignStmt); ok && len(a.Rhs) == 1 {
					if c, ok := a.Rhs[0].(*ast.CallExpr); ok && researchExprV34(c.Fun) == "s.validatorV33" {
						removed++
						continue
					}
				}
				if removed == 1 {
					if _, ok := stmt.(*ast.IfStmt); !ok {
						t.Fatal("initializer error check missing")
					}
					removed++
					continue
				}
				kept = append(kept, stmt)
			}
			if removed != 2 {
				t.Fatal("initializer count", removed)
			}
			fork.Body.List = kept
			changed := 0
			ast.Inspect(fork.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok && researchExprV34(c.Fun) == "validate" {
					c.Fun = &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("validate")}
					c.Args = []ast.Expr{ast.NewIdent("ctx"), c.Args[0], &ast.SelectorExpr{X: &ast.SelectorExpr{X: ast.NewIdent("s"), Sel: ast.NewIdent("state")}, Sel: ast.NewIdent("Head")}}
					changed++
				}
				return true
			})
			if changed != 2 {
				t.Fatal("capture validations", changed)
			}
		case "worker":
			changed := 0
			ast.Inspect(fork.Body, func(n ast.Node) bool {
				if b, ok := n.(*ast.BlockStmt); ok {
					for _, stmt := range b.List {
						loop, ok := stmt.(*ast.RangeStmt)
						if !ok || researchExprV34(loop.X) != "batch" || len(loop.Body.List) != 1 {
							continue
						}
						cond, ok := loop.Body.List[0].(*ast.IfStmt)
						if !ok || researchExprV34(cond.Cond) != "s.deferredProofs" {
							continue
						}
						old := cond.Else.(*ast.BlockStmt)
						if researchExprV34(old.List[len(old.List)-1]) != "s.proofBeforeAck++" {
							t.Fatal("proof counter missing")
						}
						old.List = old.List[:len(old.List)-1]
						loop.Body = old
						changed++
					}
				}
				return true
			})
			if changed != 1 {
				t.Fatal("proof branch count", changed)
			}
		case "attachBatchArchive":
			fork.Type.Params.List = fork.Type.Params.List[:len(fork.Type.Params.List)-1]
			changed := 0
			ast.Inspect(fork.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CompositeLit); ok && researchExprV34(c.Type) == "batchArchiveV33" {
					c.Elts = c.Elts[:len(c.Elts)-2]
					changed++
				}
				return true
			})
			if changed != 1 {
				t.Fatal("attachment factor count", changed)
			}
		case "runBatchArchiveLoad":
			ast.Inspect(fork, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok && id.Name == "mode" {
					id.Name = "combined"
				}
				return true
			})
			fork.Type.Params.List = base.Type.Params.List
		}
		if !reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
			t.Fatal("undeclared transformation", name)
		}
		fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
		if reflect.DeepEqual(researchTokensV34(base), researchTokensV34(fork)) {
			t.Fatal("body corruption escaped", name)
		}
	}
}

func TestResearchArchiveCostAuthorityV34(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_ARCHIVE_COST_V34") != "1" {
		t.Skip("isolated paired authority controls")
	}
	f := createWitnessFixtureV23(t, true)
	defer f.close()
	s := attachBatchArchiveV34(t, f, 128, true, true)
	defer s.Close()
	ctx := context.Background()
	if _, _, err := s.recall(ctx, f.svc, f.request(time.Now().UTC(), "capture")); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	c := s.pendingProofs[0].capture
	s.mu.Unlock()
	s.owner.Lock()
	defer s.owner.Unlock()
	check := func(c *archiveCaptureV29, current model.Snapshot, valid bool) {
		t.Helper()
		oldErr := s.validate(ctx, c, current)
		validator, newErr := s.validatorV34(ctx, current)
		if newErr == nil {
			newErr = validator(c)
		}
		if (oldErr == nil) != valid || (newErr == nil) != valid {
			t.Fatalf("validator mismatch want=%v old=%v new=%v", valid, oldErr, newErr)
		}
	}
	check(c, s.state.Head, true)
	for _, kind := range []string{"nil", "owner", "wire", "seal", "root", "snapshot", "asof", "current"} {
		clone := *c
		wire, _ := json.Marshal(c.entry)
		if err := json.Unmarshal(wire, &clone.entry); err != nil {
			t.Fatal(err)
		}
		current := s.state.Head
		switch kind {
		case "nil":
			check(nil, current, false)
			continue
		case "owner":
			clone.owner = &archiveWitnessV29{}
		case "wire":
			clone.entry.ID += "forged"
		case "seal":
			clone.seal = "forged"
		case "root":
			clone.root = "unknown"
			clone.seal = witnessHashV23([]any{clone.entry, clone.binding, clone.root})
		case "snapshot":
			clone.entry.Snapshot.EvidenceEpoch++
			clone.binding.Snapshot = clone.entry.Snapshot
			clone.wire = witnessHashV23(clone.entry)
			clone.seal = witnessHashV23([]any{clone.entry, clone.binding, clone.root})
		case "asof":
			clone.entry.AsOf = clone.entry.AsOf.Add(time.Second)
			clone.wire = witnessHashV23(clone.entry)
			clone.seal = witnessHashV23([]any{clone.entry, clone.binding, clone.root})
		case "current":
			current.EvidenceEpoch++
		}
		check(&clone, current, false)
	}
	var seq int
	var payload, prior, digest string
	if err := s.gate.sidecar.QueryRow("SELECT seq,payload,prior,digest FROM witness_v23 ORDER BY seq LIMIT 1").Scan(&seq, &payload, &prior, &digest); err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"payload", "prior", "digest"} {
		if _, err := s.gate.sidecar.Exec("UPDATE witness_v23 SET "+column+"='corrupt' WHERE seq=?", seq); err != nil {
			t.Fatal(err)
		}
		check(c, s.state.Head, false)
		if _, err := s.gate.sidecar.Exec("UPDATE witness_v23 SET payload=?,prior=?,digest=? WHERE seq=?", payload, prior, digest, seq); err != nil {
			t.Fatal(err)
		}
	}
	oldHash := s.state.Hash
	s.state.Hash = "forged"
	check(c, s.state.Head, false)
	s.state.Hash = oldHash
	s.poison.Store(true)
	check(c, s.state.Head, false)
	s.poison.Store(false)
	check(c, s.state.Head, true)
	t.Log("original/shared reject eight authority defects, three chain corruptions, state-hash and poison; valid historical capture retained")
}
