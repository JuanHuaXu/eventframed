package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/transfergenerator"
)

const transferV116Base int64 = 2180111700

type transferV116Artifact struct {
	Version, Go, OS, Arch string
	SeedBase              int64
	PerCase               int
	Hashes                map[string]string
	Records               []transferRecord
}

func transferV116Hashes(t *testing.T) map[string]string {
	t.Helper()
	paths := []string{"go.mod", "go.sum", "docs/experiments/mmm-transfer-v116-protocol.md", "research/independent-generator-contract.md", "research/transfer-v116-summary.mjs"}
	for _, dir := range []string{"observation", "observationlearners", "transfergenerator"} {
		found, err := filepath.Glob(filepath.Join("../..", "internal", dir, "*.go"))
		if err != nil || len(found) == 0 {
			t.Fatal("source discovery", dir, err)
		}
		for _, path := range found {
			rel, err := filepath.Rel("../..", path)
			if err != nil {
				t.Fatal(err)
			}
			paths = append(paths, rel)
		}
	}
	out := map[string]string{}
	for _, path := range paths {
		raw, err := os.ReadFile(filepath.Join("../..", path))
		if err != nil {
			t.Fatal(err)
		}
		out[path] = fmt.Sprintf("%x", sha256.Sum256(raw))
	}
	return out
}

func TestTransferV116Seeds(t *testing.T) {
	used := map[int64]bool{}
	for phase := int64(0); phase < 2; phase++ {
		for family := int64(0); family < 3; family++ {
			for mode := int64(0); mode < 3; mode++ {
				for index := int64(0); index < 32; index++ {
					for role := int64(0); role < 5; role++ {
						seed := (transferV116Base + phase*1000000 + family*100000 + mode*10000 + index*10 + role) % 2147483647
						if used[seed] {
							t.Fatal("quality seed collision")
						}
						used[seed] = true
					}
				}
			}
		}
	}
	// Include the QA block conservatively as thirty contiguous case offsets.
	for _, base := range []int64{2026119000, 2030119100, 2034119200, 2038119300, 2042119400, 2046119500, 2050119600, 2054119700, 2062119900, 2066110000, 2078110100, 2082110200, 2090110300, 2100110400, 2110110600, 2120110800, 2130110900, 2140111100, 2144111200, 2150111300, 2160111500, 2176111600} {
		for phase := int64(0); phase < 2; phase++ {
			for c := int64(0); c < 30; c++ {
				for index := int64(0); index < 128; index++ {
					for role := int64(0); role < 5; role++ {
						if used[(base+phase*1000000+c*10000+index*10+role)%2147483647] {
							t.Fatal("archived/QA seed collision", base)
						}
					}
				}
			}
		}
	}
	for _, base := range []int64{3070119900, 3090110000, 3110110100, 3130110200} {
		for mode := int64(0); mode < 2; mode++ {
			for family := int64(0); family < 4096; family++ {
				for test := int64(0); test < 64; test++ {
					if used[(base+mode*10000000+family*1000+test)%2147483647] {
						t.Fatal("archived null seed collision", base)
					}
				}
			}
		}
	}
}

func TestTransferV116(t *testing.T) {
	path := os.Getenv("EVENTFRAME_TRANSFER_V116")
	if path == "" {
		t.Skip("explicit exclusive artifact required")
	}
	hashes := transferV116Hashes(t)
	a := transferV116Artifact{Version: "transfer-v116", Go: runtime.Version(), OS: runtime.GOOS, Arch: runtime.GOARCH, SeedBase: transferV116Base, PerCase: 32, Hashes: hashes}
	replay := os.Getenv("EVENTFRAME_TRANSFER_V116_REPLAY") == "1"
	var old transferV116Artifact
	if replay {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(raw, &old); err != nil {
			t.Fatal(err)
		}
		if old.Version != a.Version || old.SeedBase != a.SeedBase || old.PerCase != 32 || len(old.Records) != 1152 || !reflect.DeepEqual(old.Hashes, hashes) {
			t.Fatal("artifact contract/source mismatch")
		}
	} else if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("output already exists or cannot be inspected", err)
	}
	for phase := 0; phase < 2; phase++ {
		for family := transfergenerator.Additive; family <= transfergenerator.LocalTable; family++ {
			for mode := transfergenerator.Stationary; mode <= transfergenerator.Gradual; mode++ {
				for index := 0; index < 32; index++ {
					spec := transfergenerator.Spec{SeedBase: a.SeedBase, Family: family, Mode: mode, Phase: phase, Index: index}
					data, teacher, err := transfergenerator.Generate(spec)
					if err != nil {
						t.Fatal(err)
					}
					for schedule := 0; schedule < 2; schedule++ {
						r, err := transferRunEvidence(data, teacher, schedule)
						if err != nil {
							t.Fatal(err)
						}
						if replay && !reflect.DeepEqual(old.Records[len(a.Records)], r) {
							t.Fatal("replay mismatch", spec, schedule)
						}
						a.Records = append(a.Records, r)
					}
				}
			}
		}
	}
	if replay {
		t.Log("all1152 trajectories replay exactly", len(hashes), "source hashes")
		return
	}
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(raw); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote", len(a.Records), "trajectories", len(raw), "bytes", len(hashes), "source hashes")
}
