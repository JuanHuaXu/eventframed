package libravdbstore

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"reflect"
	"testing"
)

func TestResearchBatchArchiveWorkloadV33(t *testing.T) {
	read := func(path, name string) *ast.FuncDecl {
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
				return fn
			}
		}
		t.Fatal("missing workload", name)
		return nil
	}
	base := read("research_combined_witness_v26_load_test.go", "runCombinedWitnessLoadV26")
	fork := read("research_batch_archive_load_v33_generated_test.go", "runBatchArchiveLoadV33")
	fork.Name.Name = base.Name.Name
	fork.Type.Results.List[0].Type = ast.NewIdent("joinedTrialV25")
	changes := map[string]int{}
	ast.Inspect(fork.Body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "attachLoadArchiveV33" {
			id.Name = "attachCombinedV26"
			changes["attachment"]++
		}
		if kv, ok := n.(*ast.KeyValueExpr); ok {
			if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Scheduled" {
				kv.Value = ast.NewIdent("combined")
				changes["diagnostic"]++
			}
		}
		if ret, ok := n.(*ast.ReturnStmt); ok && len(ret.Results) == 1 {
			if call, ok := ret.Results[0].(*ast.CallExpr); ok {
				if id, ok := call.Fun.(*ast.Ident); ok && id.Name == "finishArchiveTrialV33" {
					if len(call.Args) != 2 {
						t.Fatal("bad archive diagnostic wrapper")
					}
					ret.Results[0] = call.Args[1]
					changes["result"]++
				}
			}
		}
		return true
	})
	var statements []ast.Stmt
	for _, stmt := range fork.Body.List {
		if d, ok := stmt.(*ast.DeferStmt); ok {
			if sel, ok := d.Call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "closeV33" {
				changes["drain"]++
				continue
			}
		}
		statements = append(statements, stmt)
	}
	fork.Body.List = statements
	for _, key := range []string{"attachment", "diagnostic", "result", "drain"} {
		if changes[key] != 1 {
			t.Fatal("unexpected transformation", key, changes)
		}
	}
	normalize := func(fn *ast.FuncDecl) []string {
		var b bytes.Buffer
		if err := format.Node(&b, token.NewFileSet(), fn); err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		file := fset.AddFile("fn", -1, b.Len())
		var s scanner.Scanner
		s.Init(file, b.Bytes(), nil, 0)
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
	if !reflect.DeepEqual(normalize(base), normalize(fork)) {
		t.Fatal("workload/offers/gates differ beyond declared attachment/diagnostics")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(normalize(base), normalize(fork)) {
		t.Fatal("workload corruption escaped inverse checker")
	}
	t.Log("offer rates, independent producers, full request measurements and gates preserved")
}
