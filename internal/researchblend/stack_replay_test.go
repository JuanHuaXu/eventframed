package researchblend

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// This post-collection audit leaves the sealed collector and checker untouched.
// Only elapsed timings are excluded; issued rows and RNG choices must replay.
func TestStackReplayV31(t *testing.T) {
	path := os.Getenv("EVENTFRAME_STACK_V31_REPLAY")
	if path == "" {
		t.Skip("opt-in archived stacking tape replay")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var manifest struct {
		Kind, Split string
		SeedBase    int64
		Worlds      int
		Sources     map[string]string
	}
	if err = dec.Decode(&manifest); err != nil {
		t.Fatal(err)
	}
	seed := int64(2026103103)
	if manifest.Split == "confirmation" {
		seed = 2026103104
	} else if manifest.Split != "design" {
		t.Fatal("unknown split")
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 768 || manifest.SeedBase != seed || len(manifest.Sources) != len(sourcesV31) {
		t.Fatal("invalid manifest")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sourcesV31 {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != manifest.Sources[source] {
			t.Fatalf("missing or changed source: %s", source)
		}
	}
	clearTimings := func(w *worldV31) {
		for i := range w.Arms {
			a := &w.Arms[i]
			a.SetupNS, a.SelectionNS, a.UpdateNS, a.FinalNS, a.TotalNS = 0, 0, 0, 0, 0
		}
	}
	count := 0
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				var got worldV31
				if err = dec.Decode(&got); err != nil {
					t.Fatal(err)
				}
				want := makeWorldV31(seed, g, r, w)
				clearTimings(&got)
				clearTimings(&want)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("seed replay differs: geometry=%d regime=%d world=%d", g, r, w)
				}
				count++
			}
		}
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF || count != 768 {
		t.Fatal("extra or missing worlds")
	}
	t.Logf("replayed %d worlds / %d arms including issued forecasts; only five elapsed fields excluded", count, count*15)
}
