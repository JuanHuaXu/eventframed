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

// Independent inverse transformation proves persistence did not quietly lose
// a native receipt/readback, stale-LSN gate, or failure transition in the fork.
func TestResearchArchiveGenerationV29(t *testing.T) {
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
		t.Fatal("missing function", name)
		return nil
	}
	base := read("research_joined_journal_v25_test.go", "appendJoinedJournalV25")
	fork := read("research_archive_journal_v29_generated_test.go", "appendArchiveJournalV29")
	fork.Name.Name = base.Name.Name
	last := fork.Type.Params.List[len(fork.Type.Params.List)-1]
	if len(last.Names) != 1 || last.Names[0].Name != "admit" {
		t.Fatal("missing admission parameter")
	}
	fork.Type.Params.List = fork.Type.Params.List[:len(fork.Type.Params.List)-1]
	n := 0
	ast.Inspect(fork.Body, func(node ast.Node) bool {
		c, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := c.Fun.(*ast.Ident)
		if !ok || id.Name != "admit" {
			return true
		}
		if len(c.Args) != 2 {
			t.Fatal("wrong admission arguments")
		}
		e, err := parser.ParseExpr("store.JournalSnapshotCompatible(entry.Snapshot, beforeSnapshot, entry.AsOf, g.store.ingestMotion)")
		if err != nil {
			t.Fatal(err)
		}
		replacement := e.(*ast.CallExpr)
		c.Fun, c.Args = replacement.Fun, replacement.Args
		n++
		return true
	})
	if n != 1 {
		t.Fatal("wrong admission count", n)
	}
	normalize := func(fn *ast.FuncDecl) []string {
		var b bytes.Buffer
		if err := format.Node(&b, token.NewFileSet(), fn); err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		file := fset.AddFile("fn", -1, b.Len())
		var scan scanner.Scanner
		scan.Init(file, b.Bytes(), nil, 0)
		var tokens []string
		for {
			_, tok, lit := scan.Scan()
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
		t.Fatal("native persistence differs beyond declared admission change")
	}
	fork.Body.List = fork.Body.List[:len(fork.Body.List)-1]
	if reflect.DeepEqual(normalize(base), normalize(fork)) {
		t.Fatal("persistence corruption control escaped checker")
	}
}
