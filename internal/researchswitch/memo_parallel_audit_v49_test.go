package researchswitch

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// Only reference verification is parallel. Every job owns decoded tapes; no
// candidate timing, predictor state, score or hypothesis is changed here.
func parallelMemoAuditV49(input, raw io.Reader, workers int, audit func(studyFixture, hybridRecordV48) error) (int, error) {
	if workers < 1 || workers > 4 || audit == nil {
		return 0, errors.New("invalid audit bound")
	}
	fd, rd := json.NewDecoder(bufio.NewReader(input)), json.NewDecoder(bufio.NewReader(raw))
	rd.DisallowUnknownFields()
	var fm, rm map[string]any
	if e := fd.Decode(&fm); e != nil {
		return 0, e
	}
	if e := rd.Decode(&rm); e != nil {
		return 0, e
	}
	if fm["Kind"] != "fixture_manifest" || rm["Kind"] != "hybrid_study_manifest" {
		return 0, errors.New("manifest kinds")
	}
	rm["Kind"] = fm["Kind"]
	if !reflect.DeepEqual(fm, rm) {
		return 0, errors.New("manifest identity")
	}
	expected, ok := fm["Worlds"].(float64)
	if !ok || expected <= 0 || expected != float64(int(expected)) || expected > 448 {
		return 0, errors.New("world count bound")
	}
	type job struct {
		f studyFixture
		r hybridRecordV48
	}
	jobs := make(chan job, 1)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var first error
	completed := 0
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				e := audit(j.f, j.r)
				mu.Lock()
				completed++
				if e != nil && first == nil {
					first = e
				}
				mu.Unlock()
			}
		}()
	}
	count := 0
	seen := map[int64]bool{}
	decode := func() error {
		for {
			var f studyFixture
			e := fd.Decode(&f)
			if e == io.EOF {
				break
			}
			if e != nil {
				return e
			}
			var r hybridRecordV48
			if e := rd.Decode(&r); e != nil {
				return e
			}
			if seen[r.Seed] || r.Seed != f.World.Population.Seed || len(r.Arms) != 9 {
				return errors.New("record identity/coverage")
			}
			seen[r.Seed] = true
			for j, a := range r.Arms {
				if a.Mode != hybridModesV48[j%3] || a.Schedule != studySchedules[j/3] {
					return errors.New("arm order")
				}
			}
			count++
			if count > int(expected) {
				return errors.New("extra fixture")
			}
			jobs <- job{f, r}
		}
		var extra any
		if e := rd.Decode(&extra); e != io.EOF {
			return errors.New("extra or invalid raw record")
		}
		if count != int(expected) {
			return errors.New("missing worlds")
		}
		return nil
	}
	e := decode()
	close(jobs)
	wg.Wait()
	if e != nil {
		return completed, e
	}
	if first != nil {
		return completed, first
	}
	if completed != count {
		return completed, errors.New("unfinished reference job")
	}
	return completed, nil
}

func auditMemoRecordV49(f studyFixture, r hybridRecordV48) error {
	for j, a := range r.Arms {
		if e := auditHybridV48(f, a, j/3); e != nil {
			return fmt.Errorf("seed %d arm %d: %w", r.Seed, j, e)
		}
	}
	return nil
}

func TestMemoV49ParallelAuditLifecycle(t *testing.T) {
	fixture := studyFixture{}
	fixture.World.Population.Seed = 7
	record := hybridRecordV48{Seed: 7}
	for s := 0; s < 3; s++ {
		for _, m := range hybridModesV48 {
			record.Arms = append(record.Arms, hybridArmV48{Mode: m, Schedule: studySchedules[s]})
		}
	}
	encode := func(kind string, x any) string {
		var b strings.Builder
		e := json.NewEncoder(&b)
		if err := e.Encode(map[string]any{"Kind": kind, "Worlds": 1}); err != nil {
			t.Fatal(err)
		}
		if err := e.Encode(x); err != nil {
			t.Fatal(err)
		}
		return b.String()
	}
	f, r := encode("fixture_manifest", fixture), encode("hybrid_study_manifest", record)
	for _, n := range []int{1, 4} {
		calls := 0
		var mu sync.Mutex
		count, e := parallelMemoAuditV49(strings.NewReader(f), strings.NewReader(r), n, func(x studyFixture, y hybridRecordV48) error {
			mu.Lock()
			calls++
			mu.Unlock()
			if x.World.Population.Seed != y.Seed {
				return errors.New("wrong tape")
			}
			return nil
		})
		if e != nil || count != 1 || calls != 1 {
			t.Fatal("positive coverage", count, calls, e)
		}
		count, e = parallelMemoAuditV49(strings.NewReader(f), strings.NewReader(r), n, func(studyFixture, hybridRecordV48) error { return errors.New("injected worker rejection") })
		if e == nil || count != 1 {
			t.Fatal("worker error lost", count, e)
		}
		for _, bad := range []string{r + r, r[:len(r)/2], strings.Replace(r, "static", "bad", 1), strings.Replace(r, "\"Seed\":7", "\"Seed\":8", 1)} {
			if _, e = parallelMemoAuditV49(strings.NewReader(f), strings.NewReader(bad), n, func(studyFixture, hybridRecordV48) error { return nil }); e == nil {
				t.Fatal("corruption accepted")
			}
		}
	}
	for _, n := range []int{0, 5} {
		if _, e := parallelMemoAuditV49(strings.NewReader(f), strings.NewReader(r), n, func(studyFixture, hybridRecordV48) error { return nil }); e == nil {
			t.Fatal("bound accepted")
		}
	}
}

func TestMemoV49ParallelAuditReference(t *testing.T) {
	input, raw := os.Getenv("EVENTFRAME_HYBRID_V48_FIXTURE"), os.Getenv("EVENTFRAME_HYBRID_V48_AUDIT")
	if raw == "" {
		t.Skip("explicit full reference inputs required")
	}
	f, e := os.Open(input)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	r, e := os.Open(raw)
	if e != nil {
		t.Fatal(e)
	}
	defer r.Close()
	count, e := parallelMemoAuditV49(f, r, 4, auditMemoRecordV49)
	if e != nil {
		t.Fatal(e)
	}
	t.Log("unchanged dense reference verified all worlds", count)
}
