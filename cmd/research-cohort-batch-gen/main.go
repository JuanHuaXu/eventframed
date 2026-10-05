// Fork only isolated test implementations; production and earlier studies remain unchanged.
package main

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

type record struct {
	Source, Name, Fork string
	Receiver           string
	Worker             bool
	Native             bool
}

func main() {
	base := "internal/store/libravdbstore/"
	inputs := []string{
		"research_joined_witness_v25_test.go", "research_combined_witness_v26_test.go",
		"research_joined_journal_v25_test.go", "research_archive_journal_v29_generated_test.go",
		"research_archive_cost_v34_generated_test.go", "research_archive_cost_v34_test.go",
		"research_archive_cost_checks_v34_generated_test.go",
		"research_metadata_projection_v35_test.go", "research_metadata_core_v36_test.go",
		"research_metadata_getters_v35_generated_test.go",
		"research_warm_search_v37_test.go", "research_eager_load_v44_test.go",
		"research_eager_load_v44_generated_test.go",
		"research_eager_load_v44_controls_test.go",
	}
	files := make([]*ast.File, 0, len(inputs))
	renames := map[string]string{}
	renames["appendJoinedJournalV25"] = "appendJoinedJournalV25_CohortV46"
	renames["appendArchiveJournalV29"] = "appendArchiveJournalV29_CohortV46"
	keepType := map[string]bool{"joinedJobV25": true, "joinedBatchV25": true}
	var records []record
	for _, input := range inputs {
		f, err := parser.ParseFile(token.NewFileSet(), base+input, nil, 0)
		if err != nil {
			panic(err)
		}
		var kept []ast.Decl
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok {
				if input == "research_eager_load_v44_controls_test.go" && fn.Name.Name != "TestResearchEagerTracedOwnerAndDurabilityV44" {
					continue
				}
				if input == "research_combined_witness_v26_test.go" && fn.Name.Name != "attachCombinedV26" {
					continue
				}
				if input == "research_eager_load_v44_generated_test.go" && fn.Name.Name != "runEagerLoadV44" {
					continue
				}
				if strings.HasPrefix(fn.Name.Name, "Test") && input != "research_joined_witness_v25_test.go" && input != "research_archive_cost_checks_v34_generated_test.go" && input != "research_eager_load_v44_controls_test.go" {
					continue
				}
				if fn.Recv == nil {
					renames[fn.Name.Name] = fn.Name.Name + "_CohortV46"
				}
				receiver := ""
				if fn.Recv != nil {
					receiver = researchCall(fn.Recv.List[0].Type)
				}
				records = append(records, record{input, fn.Name.Name, fn.Name.Name, receiver, fn.Name.Name == "worker" || fn.Name.Name == "workerV34", fn.Name.Name == "appendJoinedJournalV25" || fn.Name.Name == "appendArchiveJournalV29"})
			}
			if gen, ok := d.(*ast.GenDecl); ok && gen.Tok != token.IMPORT {
				var specs []ast.Spec
				for _, spec := range gen.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if keepType[s.Name.Name] {
							continue
						}
						renames[s.Name.Name] = s.Name.Name + "_CohortV46"
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.Name != "_" {
								renames[name.Name] = name.Name + "_CohortV46"
							}
						}
					}
					specs = append(specs, spec)
				}
				gen.Specs = specs
				if len(specs) == 0 {
					continue
				}
			}
			kept = append(kept, d)
		}
		f.Decls = kept
		files = append(files, f)
	}
	imports := map[string]*ast.ImportSpec{}
	out := &ast.File{Name: ast.NewIdent("libravdbstore")}
	changes := 0
	nativeChanges, messages := 0, 0
	for _, f := range files {
		for _, d := range f.Decls {
			if gen, ok := d.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				for _, spec := range gen.Specs {
					s := spec.(*ast.ImportSpec)
					imports[s.Path.Value] = s
				}
				continue
			}
			if fn, ok := d.(*ast.FuncDecl); ok && (fn.Name.Name == "worker" || fn.Name.Name == "workerV34") {
				ast.Inspect(fn, func(n ast.Node) bool {
					b, ok := n.(*ast.BinaryExpr)
					if ok && b.Op == token.LSS {
						call, ok := b.X.(*ast.CallExpr)
						lit, isLit := b.Y.(*ast.BasicLit)
						if ok && isLit && researchCall(call) == "len(batch)" && lit.Value == "4" {
							lit.Value = "8"
							changes++
						}
					}
					return true
				})
			}
			if fn, ok := d.(*ast.FuncDecl); ok && (fn.Name.Name == "appendJoinedJournalV25" || fn.Name.Name == "appendArchiveJournalV29") {
				ast.Inspect(fn, func(n ast.Node) bool {
					if b, ok := n.(*ast.BinaryExpr); ok && b.Op == token.GTR && researchCall(b.X) == "len(entries)" {
						if lit, ok := b.Y.(*ast.BasicLit); ok && lit.Value == "4" {
							lit.Value = "8"
							nativeChanges++
						}
					}
					if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == strconv.Quote("journal batch size must be in [1,4]") {
						lit.Value = strconv.Quote("journal batch size must be in [1,8]")
						messages++
					}
					return true
				})
			}
			ast.Inspect(d, func(n ast.Node) bool {
				if id, ok := n.(*ast.Ident); ok {
					if name, exists := renames[id.Name]; exists {
						id.Name = name
					}
				}
				return true
			})
			out.Decls = append(out.Decls, d)
		}
	}
	if changes != 2 {
		panic("exactly two worker collection bounds must change")
	}
	if nativeChanges != 2 || messages != 2 {
		panic("exactly two native bounds and diagnostic messages must change")
	}
	used := map[string]bool{}
	ast.Inspect(out, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	keys := make([]string, 0, len(imports))
	for k := range imports {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	decl := &ast.GenDecl{Tok: token.IMPORT, Lparen: 1}
	for _, k := range keys {
		s := imports[k]
		p, _ := strconv.Unquote(k)
		alias := path.Base(p)
		if s.Name != nil {
			alias = s.Name.Name
		}
		if used[alias] {
			decl.Specs = append(decl.Specs, s)
		}
	}
	out.Decls = append([]ast.Decl{decl}, out.Decls...)
	for i := range records {
		if name, ok := renames[records[i].Name]; ok {
			records[i].Fork = name
		}
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, token.NewFileSet(), out); err != nil {
		panic(err)
	}
	data, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	write(base+"research_cohort_batch_v46_generated_test.go", append([]byte("// Code generated by research-cohort-batch-gen; DO NOT EDIT.\n"), data...))
	manifest, err := json.MarshalIndent(struct {
		Renames             map[string]string
		Functions           []record
		WorkerBoundsChanged int
		NativeBoundsChanged int
	}{renames, records, changes, nativeChanges}, "", "  ")
	if err != nil {
		panic(err)
	}
	write("research/cohort-batch-v46-generation.json", append(manifest, '\n'))
}

func researchCall(n ast.Node) string {
	var b bytes.Buffer
	if err := format.Node(&b, token.NewFileSet(), n); err != nil {
		panic(err)
	}
	return b.String()
}
func write(p string, b []byte) {
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write(b); err != nil {
		panic(err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
}
