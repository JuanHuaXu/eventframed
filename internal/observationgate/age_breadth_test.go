package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type ageBreadthRecord struct {
	ageRecord
	Config    subsetBreadthCase
	TrainSeed int64
}

func TestAgeBreadthContracts(t *testing.T) {
	seen := map[int64]bool{}
	add := func(seed int64) {
		if seen[seed] {
			t.Fatal("seed collision", seed)
		}
		seen[seed] = true
	}
	for phase := 0; phase < 2; phase++ {
		for j := range subsetBreadthCases {
			for i := 0; i < 64; i++ {
				add(int64(2026118800)*1000000 + int64(j*10000+phase*1000+i))
				for role := 0; role < 5; role++ {
					add(observationpreserved.Seed(int64(2026118801+10*j+phase), j, i, role))
				}
			}
		}
	}
	cfg := subsetBreadthCases[1]
	base, e := breadthBase(cfg, 2026118899)
	if e != nil {
		t.Fatal(e)
	}
	for _, schedule := range []feedbackSchedule{learningSchedules[0], learningSchedules[3]} {
		a, e := ageLearningRun(base, "unit", 1, 0, 2026118899, schedule)
		if e != nil {
			t.Fatal(e)
		}
		b, e := ageBreadthRun(base, "unit", 1, 0, 2026118899, schedule, cfg)
		if e != nil {
			t.Fatal(e)
		}
		b.Scenario = a.Scenario
		if a != b {
			t.Fatal("unchanged bit integration parity")
		}
	}
}
func TestAgeBreadthV88(t *testing.T) {
	out, replay := os.Getenv("EVENTFRAME_AGE_BREADTH_OUT"), os.Getenv("EVENTFRAME_AGE_BREADTH_REPLAY")
	if out == "" && replay == "" {
		t.Skip("opt-in")
	}
	if out != "" && replay != "" {
		t.Fatal("choose one")
	}
	root, e := filepath.Abs("../..")
	if e != nil {
		t.Fatal(e)
	}
	var header struct {
		Version string
		Hashes  map[string]string
	}
	var dec *json.Decoder
	if replay != "" {
		f, e := os.Open(replay)
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		dec = json.NewDecoder(f)
		dec.DisallowUnknownFields()
		if e := dec.Decode(&header); e != nil || header.Version != "v88" {
			t.Fatal("header", e)
		}
	} else {
		header.Version = "v88"
		header.Hashes = map[string]string{}
		var paths []string
		for _, dir := range []string{"observationgate", "observation", "observationlearners", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
			ps, e := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
			if e != nil {
				t.Fatal(e)
			}
			paths = append(paths, ps...)
		}
		for _, p := range []string{"docs/experiments/mmm-age-breadth-v88-protocol.md", "research/generate-age-breadth-v88.mjs", "research/age-breadth-v88-summary.mjs", "go.mod", "go.sum"} {
			paths = append(paths, filepath.Join(root, p))
		}
		for _, p := range paths {
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			h := sha256.Sum256(b)
			rel, e := filepath.Rel(root, p)
			if e != nil {
				t.Fatal(e)
			}
			header.Hashes[rel] = hex.EncodeToString(h[:])
		}
	}
	for p, want := range header.Hashes {
		if !filepath.IsLocal(p) {
			t.Fatal("source path")
		}
		b, e := os.ReadFile(filepath.Join(root, p))
		if e != nil {
			t.Fatal(e)
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != want {
			t.Fatal("source changed", p)
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(header); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, cfg := range subsetBreadthCases {
			name := cfg.Name

			var latent [64]string
			for _, schedule := range []feedbackSchedule{learningSchedules[0], learningSchedules[3]} {
				for i := 0; i < 64; i++ {
					trainSeed := int64(2026118800)*1000000 + int64(scenario*10000+phase*1000+i)
					base, e := breadthBase(cfg, trainSeed)
					if e != nil {
						t.Fatal(e)
					}
					result, e := ageBreadthRun(base, split, scenario, i, int64(2026118801+10*scenario+phase), schedule, cfg)
					r := ageBreadthRecord{ageRecord: result, Config: cfg, TrainSeed: trainSeed}
					if e != nil {
						t.Fatal(e)
					}
					if latent[i] == "" {
						latent[i] = r.LatentTape
					} else if latent[i] != r.LatentTape {
						t.Fatal("latent mismatch")
					}
					if dec != nil {
						var want ageBreadthRecord
						if e := dec.Decode(&want); e != nil || want != r {
							t.Fatal("replay", split, name, schedule.Name, i, e)
						}
					} else {
						if e := enc.Encode(r); e != nil {
							t.Fatal(e)
						}
					}
				}
				t.Log(split, name, schedule.Name, "complete")
			}
		}
	}
	if dec != nil {
		var extra any
		if e := dec.Decode(&extra); e != io.EOF {
			t.Fatal("trailing rows", e)
		}
		return
	}
	f, e := os.OpenFile(out, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e := f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e := f.Sync(); e != nil {
		t.Fatal(e)
	}
}
