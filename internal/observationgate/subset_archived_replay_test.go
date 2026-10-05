package observationgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

// Archive verification uses the frozen file manifest, not a new directory glob.
// Adding an unrelated benchmark must not rewrite an earlier artifact's header.
func TestSubsetV82ArchivedReplay(t *testing.T) {
	path := os.Getenv("EVENTFRAME_SUBSET_ARCHIVED_REPLAY")
	if path == "" {
		t.Skip("opt-in archive verification")
	}
	f, e := os.Open(path)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var header struct {
		Version string
		Hashes  map[string]string
	}
	if e := dec.Decode(&header); e != nil {
		t.Fatal(e)
	}
	if header.Version != "v82" || len(header.Hashes) != 112 {
		t.Fatal("wrong archive manifest")
	}
	for p, want := range header.Hashes {
		if !filepath.IsLocal(p) {
			t.Fatal("nonlocal source path")
		}
		b, e := os.ReadFile(filepath.Join("../..", p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("archived source changed", p)
		}
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 64; i++ {
				var want subsetRecord
				if e := dec.Decode(&want); e != nil {
					t.Fatal(e)
				}
				trial := newSubsetTrial(base, scenario)
				r, e := subsetIntegrationRun(base, split, scenario, i, int64(2026118201+phase), trial)
				if e != nil {
					t.Fatal(e)
				}
				got := subsetRecord{r, trial.metrics, trial.fits, trial.subsetGuides, hex.EncodeToString(trial.hash.Sum(nil))}
				if got != want || trial.state.next != 512 || trial.state.pending != nil {
					t.Fatal("archived outcome mismatch", split, name, i)
				}
			}
			t.Log(split, name, "verified")
		}
	}
	var extra any
	if e := dec.Decode(&extra); e != io.EOF {
		t.Fatal("extra archive records", e)
	}
}
