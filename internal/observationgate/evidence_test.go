package observationgate

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestEvidence(t *testing.T) {
	f, e := os.Open("../../docs/experiments/mmm-gate-v1.json.gz")
	if os.IsNotExist(e) {
		t.Skip("experiment not run")
	}
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	z, e := gzip.NewReader(f)
	if e != nil {
		t.Fatal(e)
	}
	defer z.Close()
	var o Output
	if e = json.NewDecoder(z).Decode(&o); e != nil {
		t.Fatal(e)
	}
	if len(o.Records) != 10240 {
		t.Fatal("incomplete matrix")
	}
	for p, want := range o.Hashes {
		b, e := os.ReadFile("../../" + p)
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatalf("source drift: %s", p)
		}
	}
	seeds := map[int64]bool{}
	for _, r := range o.Records {
		if seeds[r.Seed] || len(r.Differences) != 512 {
			t.Fatal("duplicate seed or truncated record")
		}
		seeds[r.Seed] = true
	}
	// Regenerate outcomes from recorded seed domains as well as the full gates.
	// This is independent of the stored summaries but uses the same model code.
	replay := Run()
	if !reflect.DeepEqual(o.Records, replay.Records) || !reflect.DeepEqual(o.Summary, replay.Summary) || o.Pass != replay.Pass {
		t.Fatal("replay mismatch")
	}
}
