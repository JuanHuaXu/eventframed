// Structured research fork. Unchanged method bodies have inverse token checks.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"reflect"
)

var names = map[string]string{"compactModelV48": "scoreModelV51", "compactRowV48": "scoreRowV51", "compactTicketV48": "scoreTicketV51", "newCompactV48": "newScoreV51", "compactConfigV48": "scoreConfigV51", "compactReceiptV48": "scoreReceiptV51", "compactFiniteV48": "scoreFiniteV51", "compactExpertsV48": "scoreExpertsV51", "compactTrialsV48": "scoreTrialsV51", "compactFloorV48": "scoreFloorV51"}

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func syntax(b []byte) []string {
	f := token.NewFileSet().AddFile("tokens", -1, len(b))
	var s scanner.Scanner
	s.Init(f, b, nil, 0)
	var out []string
	for {
		_, t, v := s.Scan()
		if t == token.EOF {
			return out
		}
		out = append(out, t.String()+":"+v)
	}
}
func rename(f *ast.File, mapping map[string]string) {
	selectors := map[*ast.Ident]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if x, ok := n.(*ast.SelectorExpr); ok {
			selectors[x.Sel] = true
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		if x, ok := n.(*ast.Ident); ok && !selectors[x] {
			if v, found := mapping[x.Name]; found {
				x.Name = v
			}
		}
		return true
	})
}
func statements(src string) []ast.Stmt {
	f, err := parser.ParseFile(token.NewFileSet(), "fragment", "package p;func fragment(){"+src+"}", 0)
	if err != nil {
		panic(err)
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List
}
func expression(src string) ast.Expr {
	x, err := parser.ParseExpr(src)
	if err != nil {
		panic(err)
	}
	return x
}
func exclusive(p string, b []byte) error {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(b); err != nil {
		return err
	}
	return f.Sync()
}
func body(f *ast.File, name string) []string {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
			var b bytes.Buffer
			if err := format.Node(&b, token.NewFileSet(), fn.Body); err != nil {
				panic(err)
			}
			return syntax(b.Bytes())
		}
	}
	panic("missing method " + name)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	const source, target = "internal/researchswitch/compact_v48_generated.go", "internal/researchswitch/scored_model_v51_generated.go"
	original, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, source, original, parser.ParseComments)
	if err != nil {
		return err
	}
	rename(f, names)
	fields, newCount, predictCount, emissionCount, epochCount := 0, 0, 0, 0, 0
	for _, d := range f.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.TYPE {
			for _, s := range g.Specs {
				t := s.(*ast.TypeSpec)
				if t.Name.Name == "scoreConfigV51" || t.Name.Name == "scoreModelV51" {
					name := "mode"
					if t.Name.Name == "scoreConfigV51" {
						name = "Mode"
					}
					st := t.Type.(*ast.StructType)
					st.Fields.List = append(st.Fields.List, &ast.Field{Names: []*ast.Ident{ast.NewIdent(name)}, Type: ast.NewIdent("string")})
					fields++
				}
			}
		}
		fn, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		switch fn.Name.Name {
		case "newScoreV51":
			fn.Body.List = append(statements(`if !validScoreModeV51(cfg.Mode){return nil,errors.New("invalid score mode")}`), fn.Body.List...)
			ast.Inspect(fn, func(n ast.Node) bool {
				if x, ok := n.(*ast.CompositeLit); ok {
					if id, ok := x.Type.(*ast.Ident); ok && id.Name == "scoreModelV51" {
						x.Elts = append(x.Elts, &ast.KeyValueExpr{Key: ast.NewIdent("mode"), Value: expression("cfg.Mode")})
						newCount++
					}
				}
				return true
			})
		case "Predict":
			fn.Body.List[len(fn.Body.List)-1] = statements(`return m.scoredForecastV51(advice,q)`)[0]
			predictCount++
		case "Resolve":
			ast.Inspect(fn, func(n ast.Node) bool {
				if x, ok := n.(*ast.AssignStmt); ok && x.Tok == token.ADD_ASSIGN && len(x.Rhs) == 1 {
					var b bytes.Buffer
					format.Node(&b, fset, x.Rhs[0])
					if b.String() == "math.Log(p)" {
						x.Rhs[0] = expression("scoreEmissionV51(m.mode,p)")
						emissionCount++
					}
				}
				return true
			})
		case "BeginEpoch":
			ast.Inspect(fn, func(n ast.Node) bool {
				if x, ok := n.(*ast.CompositeLit); ok {
					if id, ok := x.Type.(*ast.Ident); ok && id.Name == "scoreConfigV51" {
						x.Elts = append(x.Elts, &ast.KeyValueExpr{Key: ast.NewIdent("Mode"), Value: expression("m.mode")})
						epochCount++
					}
				}
				return true
			})
		}
	}
	if fields != 2 || newCount != 1 || predictCount != 1 || emissionCount != 1 || epochCount != 1 {
		return fmt.Errorf("unexpected edit counts %d/%d/%d/%d/%d", fields, newCount, predictCount, emissionCount, epochCount)
	}
	var out bytes.Buffer
	if err := format.Node(&out, fset, f); err != nil {
		return err
	}
	generated, err := format.Source(append([]byte("// Research generated fork: strategy-loss messages, not a joint Bayesian posterior.\n"), out.Bytes()...))
	if err != nil {
		return err
	}
	back, err := parser.ParseFile(token.NewFileSet(), target, generated, 0)
	if err != nil {
		return err
	}
	inverse := map[string]string{}
	for a, b := range names {
		inverse[b] = a
	}
	rename(back, inverse)
	orig, err := parser.ParseFile(token.NewFileSet(), source, original, 0)
	if err != nil {
		return err
	}
	allowed := map[string]bool{"newCompactV48": true, "Predict": true, "Resolve": true, "BeginEpoch": true}
	unchanged := []string{}
	for _, d := range orig.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && !allowed[fn.Name.Name] {
			if !reflect.DeepEqual(body(orig, fn.Name.Name), body(back, fn.Name.Name)) {
				return fmt.Errorf("untouched body differs: %s", fn.Name.Name)
			}
			unchanged = append(unchanged, fn.Name.Name)
		}
	}
	if err := exclusive(target, generated); err != nil {
		return err
	}
	meta, err := json.MarshalIndent(map[string]any{"source": source, "target": target, "sourceSHA256": sha(original), "targetSHA256": sha(generated), "renames": names, "untouchedBodiesInverseEqual": unchanged, "changedBodies": allowed, "contractFieldsAdded": 2}, "", "  ")
	if err != nil {
		return err
	}
	if err := exclusive("research/scored-v51-generation.json", append(meta, '\n')); err != nil {
		return err
	}
	fmt.Println(string(meta))
	return nil
}
