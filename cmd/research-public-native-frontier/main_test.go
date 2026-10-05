package main

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestNativeFitOnlyBoundary(t *testing.T) {
	f := fitInput{Partition: "fit", Queries: make([]query, 351)}
	for i := range f.Queries {
		f.Queries[i] = query{ID: fmt.Sprint(i), Text: "public query"}
	}
	if err := validateFit(f); err != nil {
		t.Fatal(err)
	}
	for _, part := range []string{"calibration", "confirmation", ""} {
		v := f
		v.Partition = part
		if validateFit(v) == nil {
			t.Fatal("accepted wrong partition")
		}
	}
	v := f
	v.Queries = append([]query(nil), f.Queries...)
	v.Queries[350] = v.Queries[0]
	if validateFit(v) == nil {
		t.Fatal("accepted duplicate")
	}
	v = f
	v.Queries = v.Queries[:350]
	if validateFit(v) == nil {
		t.Fatal("accepted missing")
	}
	// ID membership and exact query text are independently frozen by the runner;
	// the structural client check alone is not evidence of held-out separation.
}

func TestNativeFitOwnedPath(t *testing.T) {
	cwd := t.TempDir()
	if _, err := ownedRoot(filepath.Join(cwd, "research", "native-test"), cwd); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{cwd, filepath.Join(cwd, "research-other", "test"), filepath.Join(cwd, "research", "..", "outside")} {
		if _, err := ownedRoot(p, cwd); err == nil {
			t.Fatal("accepted outside root")
		}
	}
}
