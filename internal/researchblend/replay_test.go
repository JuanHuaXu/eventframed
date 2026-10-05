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

// Post-collection replay is separate from the sealed collector. It regenerates
// latent draws, policy RNG choices and all law/packet fields from original seeds.
func TestReplayV30(t *testing.T) {
	path := os.Getenv("EVENTFRAME_BLEND_V30_REPLAY")
	if path == "" {
		t.Skip("opt-in archived tape replay")
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
	if manifest.Kind != "manifest" || manifest.Worlds != 768 || (manifest.Split != "design" && manifest.Split != "confirmation") {
		t.Fatal("invalid manifest")
	}
	seed := int64(2026103003)
	if manifest.Split == "confirmation" {
		seed = 2026103004
	}
	if manifest.SeedBase != seed || len(manifest.Sources) != 8 {
		t.Fatal("seed or source count differs")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for source, hash := range manifest.Sources {
		b, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != hash {
			t.Fatalf("source changed: %s", source)
		}
	}
	clearTimings := func(w *worldV30) {
		for i := range w.Arms {
			a := &w.Arms[i]
			a.SetupNS, a.SelectionNS, a.UpdateNS, a.FinalNS, a.TotalNS = 0, 0, 0, 0, 0
		}
	}
	count := 0
	for g := 0; g < 2; g++ {
		for r := range regimesV30 {
			for w := 0; w < 32; w++ {
				var got worldV30
				if err = dec.Decode(&got); err != nil {
					t.Fatal(err)
				}
				want := makeWorldV30(seed, g, r, w)
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
	t.Logf("replayed %d worlds / %d arms; only five timing fields per arm excluded", count, count*len(armsV30))
}
