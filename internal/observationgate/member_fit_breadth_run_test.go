package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observation"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestMemberFitBreadthCollect(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_MEMBER_FIT_BREADTH_OUT"), os.Getenv("EVENTFRAME_MEMBER_FIT_BREADTH_REPLAY")
	if output == "" && replay == "" {
		t.Skip("explicit output or replay required")
	}
	if output != "" && replay != "" {
		t.Fatal("choose output or replay")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, dir := range []string{"observationgate", "observation", "observationpreserved", "observationrescue", "observationexperiment", "bayes", "model"} {
		ps, err := filepath.Glob(filepath.Join(root, "internal", dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, ps...)
	}
	for _, p := range []string{"docs/experiments/mmm-member-fit-breadth-v1-protocol.md", "go.mod", "go.sum"} {
		paths = append(paths, filepath.Join(root, p))
	}
	hashes := map[string]string{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			t.Fatal(e)
		}
		hashes[rel] = fmt.Sprintf("%x", sha256.Sum256(b))
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if e := enc.Encode(struct {
		Version string
		PerCell int
		Seeds   [2]int64
		Hashes  map[string]string
	}{"member-fit-breadth-v1", 512, [2]int64{2026091511, 2026091512}, hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			for i := 0; i < 512; i++ {
				fitSeed := observationpreserved.Seed(int64(2026091511+phase), scenario, i, 4)
				base, fitHash, e := memberFreshBase(name == "null", fitSeed)
				if e != nil {
					t.Fatal(e)
				}
				r, e := memberRun(base, split, scenario, i, int64(2026091511+phase))
				if e != nil {
					t.Fatal(split, name, i, e)
				}
				if e = enc.Encode(struct {
					memberRecord
					FitSeed int64
					FitHash string
				}{r, fitSeed, fitHash}); e != nil {
					t.Fatal(e)
				}
				if base.Support() != 4096 {
					t.Fatal("base changed")
				}
			}
			t.Log(split, name, "512 complete")
		}
	}
	for rel, h := range hashes {
		b, e := os.ReadFile(filepath.Join(root, rel))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != h {
			t.Fatal("source changed", rel, e)
		}
	}
	if replay != "" {
		want, e := os.ReadFile(replay)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(want, buf.Bytes()) {
			t.Fatal("full replay mismatch")
		}
		t.Log("5120 trajectories: exact source/tape/metric replay PASS")
		return
	}
	f, e := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = f.Write(buf.Bytes()); e != nil {
		t.Fatal(e)
	}
	if e = f.Sync(); e != nil {
		t.Fatal(e)
	}
	t.Log("5120 trajectories written; source immutability PASS")
}

func memberFreshBase(null bool, seed int64) (*observation.Model, string, error) {
	rng := rand.New(rand.NewSource(seed))
	samples := make([]observation.Sample, 4096)
	for i := range samples {
		x := uint16(rng.Intn(512))
		samples[i] = observation.Sample{Bits: x, Outcome: integrationLabel(x, false, null, rng)}
	}
	data, err := json.Marshal(samples)
	if err != nil {
		return nil, "", err
	}
	model, err := observation.Fit(samples)
	return model, fmt.Sprintf("%x", sha256.Sum256(data)), err
}

func TestMemberFreshBaseContracts(t *testing.T) {
	for _, null := range []bool{false, true} {
		legacy, err := observationpreserved.Base(null)
		if err != nil {
			t.Fatal(err)
		}
		replica, h, err := memberFreshBase(null, observationpreserved.FitSeed)
		if err != nil || !reflect.DeepEqual(legacy, replica) {
			t.Fatal("legacy base mismatch", null, err)
		}
		other, oh, err := memberFreshBase(null, observationpreserved.FitSeed+1)
		if err != nil || other.Support() != 4096 || h == oh || reflect.DeepEqual(other, replica) {
			t.Fatal("fresh fit ineffective", err)
		}
	}
	seen := map[int64]bool{}
	for phase := 0; phase < 2; phase++ {
		for scenario := 0; scenario < 5; scenario++ {
			for i := 0; i < 512; i++ {
				for role := 0; role < 5; role++ {
					seed := observationpreserved.Seed(int64(2026091511+phase), scenario, i, role)
					if seen[seed] {
						t.Fatal("seed collision", seed)
					}
					seen[seed] = true
				}
			}
		}
	}
}
