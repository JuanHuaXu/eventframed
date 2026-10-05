// Independent inverse syntax check for the generated research forks.
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
	"strconv"
)

func sha(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func parse(p string) (*ast.File, []byte) {
	b, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), p, b, 0)
	if err != nil {
		panic(err)
	}
	return f, b
}
func text(n ast.Node) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), n); err != nil {
		panic(err)
	}
	return b.String()
}
func tokens(n ast.Node) []string {
	b := []byte(text(n))
	var s scanner.Scanner
	s.Init(token.NewFileSet().AddFile("x", -1, len(b)), b, nil, 0)
	var out []string
	for {
		_, k, v := s.Scan()
		if k == token.EOF {
			return out
		}
		out = append(out, k.String()+":"+v)
	}
}
func fn(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if x, ok := d.(*ast.FuncDecl); ok && x.Name.Name == name {
			return x
		}
	}
	panic("missing " + name)
}
func expression(s string) ast.Expr {
	x, err := parser.ParseExpr(s)
	if err != nil {
		panic(err)
	}
	return x
}
func statements(s string) []ast.Stmt {
	f, err := parser.ParseFile(token.NewFileSet(), "x", "package p;func f(){"+s+"}", 0)
	if err != nil {
		panic(err)
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List
}
func rename(f *ast.File, names map[string]string) {
	selectors := map[*ast.Ident]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if x, ok := n.(*ast.SelectorExpr); ok {
			selectors[x.Sel] = true
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		if x, ok := n.(*ast.Ident); ok && !selectors[x] {
			if v, yes := names[x.Name]; yes {
				x.Name = v
			}
		}
		return true
	})
}
func equal(a, b ast.Node, name string) {
	if !reflect.DeepEqual(tokens(a), tokens(b)) {
		panic("inverse syntax mismatch " + name)
	}
}
func main() {
	model, modelBytes := parse("internal/researchswitch/scored_model_v51_generated.go")
	originalModel, originalModelBytes := parse("internal/researchswitch/compact_v48_generated.go")
	for _, d := range model.Decls {
		if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.TYPE {
			for _, spec := range g.Specs {
				s := spec.(*ast.TypeSpec)
				if s.Name.Name == "scoreModelV51" || s.Name.Name == "scoreConfigV51" {
					st := s.Type.(*ast.StructType)
					last := st.Fields.List[len(st.Fields.List)-1]
					if len(last.Names) != 1 || (last.Names[0].Name != "mode" && last.Names[0].Name != "Mode") {
						panic("contract field")
					}
					st.Fields.List = st.Fields.List[:len(st.Fields.List)-1]
				}
			}
		}
	}
	constructor := fn(model, "newScoreV51")
	if text(constructor.Body.List[0]) != "if !validScoreModeV51(cfg.Mode) {\n\treturn nil, errors.New(\"invalid score mode\")\n}" {
		panic("constructor guard")
	}
	constructor.Body.List = constructor.Body.List[1:]
	for _, method := range []string{"newScoreV51", "BeginEpoch"} {
		ast.Inspect(fn(model, method), func(n ast.Node) bool {
			if x, ok := n.(*ast.CompositeLit); ok && len(x.Elts) > 0 {
				if kv, ok := x.Elts[len(x.Elts)-1].(*ast.KeyValueExpr); ok {
					if id, ok := kv.Key.(*ast.Ident); ok && (id.Name == "mode" || id.Name == "Mode") {
						x.Elts = x.Elts[:len(x.Elts)-1]
					}
				}
			}
			return true
		})
	}
	predict := fn(model, "Predict")
	if text(predict.Body.List[len(predict.Body.List)-1]) != "return m.scoredForecastV51(advice, q)" {
		panic("forecast hook")
	}
	predict.Body.List[len(predict.Body.List)-1] = statements("return q,nil")[0]
	emissions := 0
	ast.Inspect(fn(model, "Resolve"), func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok && a.Tok == token.ADD_ASSIGN && len(a.Rhs) == 1 && text(a.Rhs[0]) == "scoreEmissionV51(m.mode, p)" {
			a.Rhs[0] = expression("math.Log(p)")
			emissions++
		}
		return true
	})
	if emissions != 1 {
		panic("emission hook count")
	}
	rename(model, map[string]string{"scoreModelV51": "compactModelV48", "scoreRowV51": "compactRowV48", "scoreTicketV51": "compactTicketV48", "newScoreV51": "newCompactV48", "scoreConfigV51": "compactConfigV48", "scoreReceiptV51": "compactReceiptV48", "scoreFiniteV51": "compactFiniteV48", "scoreExpertsV51": "compactExpertsV48", "scoreTrialsV51": "compactTrialsV48", "scoreFloorV51": "compactFloorV48"})
	equal(model, originalModel, "complete model")
	checks := []string{"complete model inverse"}
	sources := map[string]string{"internal/researchswitch/scored_model_v51_generated.go": sha(modelBytes), "internal/researchswitch/compact_v48_generated.go": sha(originalModelBytes)}
	for _, s := range []struct{ source, target, old, name string }{
		{"internal/researchswitch/memo_collector_v49_generated_test.go", "internal/researchswitch/scored_collector_v51_generated_test.go", "collectMemoV49", "collectScoredV51"},
		{"internal/researchswitch/hybrid_study_v48_test.go", "internal/researchswitch/scored_audit_v51_generated_test.go", "auditHybridV48", "auditScoredV51"},
	} {
		old, oldBytes := parse(s.source)
		new, newBytes := parse(s.target)
		x := fn(new, s.name)
		x.Name.Name = s.old
		x.Type.Params.List = x.Type.Params.List[:len(x.Type.Params.List)-1]
		laws := 0
		ast.Inspect(x, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Name == "ScoredHybridTicketV51" {
				id.Name = "HybridTicket"
			}
			if c, ok := n.(*ast.CallExpr); ok {
				if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "NewScoredHybridV51" {
					id.Name = "NewMemoHybridV49"
					c.Args = c.Args[:len(c.Args)-1]
				}
				if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "End" && text(sel.X) == "researchscoreref" {
					sel.X = ast.NewIdent("researchhybridref")
					c.Args = c.Args[:len(c.Args)-1]
				}
			}
			if b, ok := n.(*ast.BlockStmt); ok {
				var list []ast.Stmt
				for i := 0; i < len(b.List); i++ {
					v := b.List[i]
					if a, ok := v.(*ast.AssignStmt); ok && len(a.Rhs) == 1 && text(a.Rhs[0]) == "researchscoreref.Forecast(sw, heads[:], style)" {
						if i+1 >= len(b.List) || text(b.List[i+1]) != "if err != nil {\n\treturn err\n}" {
							panic("forecast guard")
						}
						list = append(list, statements("q:=sw[0]*heads[0]+sw[1]*heads[1]")[0])
						i++
						laws++
					} else {
						list = append(list, v)
					}
				}
				b.List = list
			}
			return true
		})
		if s.old == "auditHybridV48" && laws != 2 {
			panic("law inverse count")
		}
		equal(x, fn(old, s.old), s.old)
		sources[s.source], sources[s.target] = sha(oldBytes), sha(newBytes)
		checks = append(checks, s.old+" complete function inverse")
	}
	fixture, b := parse("internal/researchdispersion/scored_fixture_v51_generated_test.go")
	originalFixture, oldB := parse("internal/researchdispersion/hybrid_fixture_v48_test.go")
	cohort := fn(fixture, "scoredCohortV51")
	ast.Inspect(cohort, func(n ast.Node) bool {
		if x, ok := n.(*ast.BasicLit); ok {
			if v, yes := map[string]string{"2026105107": "2026104807", "2026105109": "2026104809", "2026105111": "2026104811"}[x.Value]; yes {
				x.Value = v
			}
		}
		return true
	})
	ast.Inspect(fn(fixture, "TestScoredV51SeedSeparation"), func(n ast.Node) bool {
		if x, ok := n.(*ast.CompositeLit); ok && len(x.Elts) == 20 {
			x.Elts = x.Elts[:17]
		}
		return true
	})
	ast.Inspect(fixture, func(n ast.Node) bool {
		if x, ok := n.(*ast.BasicLit); ok && x.Kind == token.STRING {
			if v, yes := map[string]string{strconv.Quote("EVENTFRAME_SCORED_V51_FIXTURE"): strconv.Quote("EVENTFRAME_HYBRID_V48_FIXTURE"), strconv.Quote("EVENTFRAME_SCORED_V51_SPLIT"): strconv.Quote("EVENTFRAME_HYBRID_V48_SPLIT")}[x.Value]; yes {
				x.Value = v
			}
		}
		return true
	})
	rename(fixture, map[string]string{"scoredCohortV51": "hybridCohortV48", "TestScoredV51SeedSeparation": "TestSpecialistActualSeedSeparationV48", "TestScoredV51Fixture": "TestSpecialistFixtureV48"})
	equal(fixture, originalFixture, "complete fixture bridge")
	sources["internal/researchdispersion/scored_fixture_v51_generated_test.go"], sources["internal/researchdispersion/hybrid_fixture_v48_test.go"] = sha(b), sha(oldB)
	checks = append(checks, "complete fixture bridge inverse")
	self, err := os.ReadFile("cmd/research-scored-inverse/main.go")
	if err != nil {
		panic(err)
	}
	sources["cmd/research-scored-inverse/main.go"] = sha(self)
	out, _ := json.MarshalIndent(map[string]any{"allInverseSyntaxEqual": true, "checks": checks, "sources": sources, "qualityClaim": false}, "", "  ")
	f, err := os.OpenFile("research/scored-v51-inverse-audit.json", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write(append(out, '\n')); err != nil {
		panic(err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
