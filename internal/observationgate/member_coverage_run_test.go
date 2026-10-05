package observationgate

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/JuanHuaXu/eventframed/internal/observationpreserved"
)

func TestMemberCoverageCollect(t *testing.T) {
	output, replay := os.Getenv("EVENTFRAME_MEMBER_COVERAGE_OUT"), os.Getenv("EVENTFRAME_MEMBER_COVERAGE_REPLAY")
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
	for _, p := range []string{"docs/experiments/mmm-member-coverage-v1-protocol.md", "go.mod", "go.sum"} {
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
	}{"member-coverage-v1", 512, [2]int64{2026091507, 2026091508}, hashes}); e != nil {
		t.Fatal(e)
	}
	for phase, split := range []string{"design", "confirmation"} {
		for scenario, name := range observationpreserved.Scenarios {
			base, e := observationpreserved.Base(name == "null")
			if e != nil {
				t.Fatal(e)
			}
			for i := 0; i < 512; i++ {
				r, e := memberRun(base, split, scenario, i, int64(2026091507+phase))
				if e != nil {
					t.Fatal(split, name, i, e)
				}
				if e = enc.Encode(r); e != nil {
					t.Fatal(e)
				}
			}
			if base.Support() != 4096 {
				t.Fatal("base changed")
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
