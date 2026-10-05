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

// The separately added replay audit must not change a sealed V32 source.
func TestCrossfitReplayV32(t *testing.T) {
	path := os.Getenv("EVENTFRAME_CROSSFIT_V32_REPLAY")
	if path == "" {
		t.Skip("opt-in archived crossfit tape replay")
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
	seed := int64(2026103203)
	if manifest.Split == "confirmation" {
		seed = 2026103204
	} else if manifest.Split != "design" {
		t.Fatal("unknown split")
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 768 || manifest.SeedBase != seed || len(manifest.Sources) != len(sourcesV32) {
		t.Fatal("invalid manifest")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sourcesV32 {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != manifest.Sources[source] {
			t.Fatalf("missing or changed source: %s", source)
		}
	}
	clearTimings := func(w *worldV32) {
		for i := range w.Arms {
			a := &w.Arms[i]
			a.SetupNS, a.SelectionNS, a.UpdateNS, a.FinalNS, a.TotalNS = 0, 0, 0, 0, 0
		}
	}
	count := 0
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				var got worldV32
				if err = dec.Decode(&got); err != nil {
					t.Fatal(err)
				}
				want := makeWorldV32(seed, g, r, w)
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
	t.Logf("replayed %d worlds / %d arms including LOO rows; only five elapsed fields excluded", count, count*17)
}
