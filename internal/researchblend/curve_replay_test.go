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

// Post-collection seed replay is separate from all nineteen sealed sources.
func TestCurveReplayV33(t *testing.T) {
	path := os.Getenv("EVENTFRAME_CURVE_V33_REPLAY")
	if path == "" {
		t.Skip("opt-in archived learning-curve replay")
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
	seed := int64(2026103303)
	if manifest.Split == "confirmation" {
		seed = 2026103304
	} else if manifest.Split != "design" {
		t.Fatal("unknown split")
	}
	if manifest.Kind != "manifest" || manifest.Worlds != 768 || manifest.SeedBase != seed || len(manifest.Sources) != len(sourcesV33) {
		t.Fatal("invalid manifest")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sourcesV33 {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != manifest.Sources[source] {
			t.Fatalf("missing or changed source: %s", source)
		}
	}
	count := 0
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				var got worldV33
				if err = dec.Decode(&got); err != nil {
					t.Fatal(err)
				}
				want := makeWorldV33(seed, g, r, w)
				got.InputNS, want.InputNS = 0, 0
				clearCurveTimingsV33(&got.curveV33)
				clearCurveTimingsV33(&want.curveV33)
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
	t.Logf("replayed %d worlds / %d snapshots / %d six-model issued steps; only elapsed fields excluded", count, count*35, count*150)
}
