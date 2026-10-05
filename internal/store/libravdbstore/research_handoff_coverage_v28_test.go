package libravdbstore

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

// Independently check AST coverage rather than trusting embedded-interface
// compile success, which alone could silently forward uninstrumented methods.
func TestResearchHandoffCoverageV28(t *testing.T) {
	set := token.NewFileSet()
	api, err := parser.ParseFile(set, "../store.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{}
	ast.Inspect(api, func(n ast.Node) bool {
		if s, ok := n.(*ast.TypeSpec); ok && s.Name.Name == "EventStore" {
			in, ok := s.Type.(*ast.InterfaceType)
			if !ok {
				t.Fatal("interface missing")
			}
			for _, f := range in.Methods.List {
				if len(f.Names) != 1 {
					t.Fatal("embedded method")
				}
				want[f.Names[0].Name] = true
			}
		}
		return true
	})
	got := map[string]bool{}
	for _, path := range []string{"research_handoff_observer_v27_generated_test.go", "research_handoff_boundary_v27_test.go"} {
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) != 1 {
				continue
			}
			ptr, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			id, ok := ptr.X.(*ast.Ident)
			if !ok || id.Name != "handoffObserverV27" || !want[fn.Name.Name] {
				continue
			}
			if len(fn.Body.List) == 0 {
				t.Fatal("empty forwarder")
			}
			first, ok := fn.Body.List[0].(*ast.ExprStmt)
			if !ok {
				t.Fatal("no observation first", fn.Name.Name)
			}
			call, ok := first.X.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				t.Fatal("wrong observer", fn.Name.Name)
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "observe" {
				t.Fatal("wrong observer call", fn.Name.Name)
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				t.Fatal("nonliteral tag")
			}
			tag, err := strconv.Unquote(literal.Value)
			if err != nil || tag != fn.Name.Name {
				t.Fatal("wrong observer tag", fn.Name.Name)
			}
			delegated := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					if method, ok := c.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == fn.Name.Name {
						if receiver, ok := method.X.(*ast.SelectorExpr); ok && receiver.Sel.Name == "EventStore" {
							delegated = true
						}
					}
				}
				return true
			})
			if !delegated {
				t.Fatal("missing actual forwarding", fn.Name.Name)
			}
			got[fn.Name.Name] = true
		}
	}
	complete := func(have map[string]bool) bool {
		if len(have) != len(want) {
			return false
		}
		for name := range want {
			if !have[name] {
				return false
			}
		}
		return true
	}
	if len(want) != handoffMethodCountV27 || !complete(got) {
		t.Fatal("interface coverage gap", len(want), len(got))
	}
	delete(got, "Snapshot")
	if complete(got) {
		t.Fatal("missing-method negative control not detected")
	}
	t.Logf("all%d interface methods observed first and actually delegated; missing Snapshot rejected", len(want))
}
