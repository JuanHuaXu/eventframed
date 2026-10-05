package observationlearners

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
)

func spikeIndependentIdentity(n int, r softV120Record) bool {
	return n >= 0 && n < 672 && r.Schedule == n%2 && r.Index == n/2%8 && r.Case == n/16%21 && r.Phase == n/336
}

func TestSpikeIndependentIdentity(t *testing.T) {
	for n := 0; n < 672; n++ {
		r := softV120Record{Phase: n / 336, Case: n / 16 % 21, Index: n / 2 % 8, Schedule: n % 2}
		if !spikeIndependentIdentity(n, r) || spikeIndependentIdentity(n+1, r) {
			t.Fatal("identity", n)
		}
	}
	if spikeIndependentIdentity(-1, softV120Record{}) || spikeIndependentIdentity(672, softV120Record{}) {
		t.Fatal("bounds")
	}
}

// No cached fits are accepted: every block is fitted from the independent tape.
func TestSpikeIndependentFits(t *testing.T) {
	started := time.Now()
	source, prefix := os.Getenv("EVENTFRAME_SPIKE_INDEPENDENT_FIT_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_INDEPENDENT_FIT_PREFIX")
	if source == "" {
		t.Skip("explicit independent source required")
	}
	if prefix == "" {
		t.Fatal("output prefix required")
	}
	f, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var header struct {
		softV120Artifact
		Cohort string
	}
	if err := dec.Decode(&header); err != nil {
		t.Fatal(err)
	}
	if header.Version != "soft-learners-v120" || header.Cohort != "spike-independent-v1" || header.TransferBase != spikeIndependentTransfer || header.BooleanBase != spikeIndependentBoolean || header.Workers != 1 {
		t.Fatal("independent cohort contract")
	}
	sourceFile, err := os.Open(source)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, sourceFile); err != nil {
		sourceFile.Close()
		t.Fatal(err)
	}
	if err := sourceFile.Close(); err != nil {
		t.Fatal(err)
	}
	sourceHash := fmt.Sprintf("%x", hash.Sum(nil))
	codeHashes := transferV116Hashes(t)
	var outputs []*os.File
	create := func(path string) *json.Encoder {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		outputs = append(outputs, f)
		t.Cleanup(func() { f.Close() })
		return json.NewEncoder(f)
	}
	writers := make([]*json.Encoder, 8)
	for block := range writers {
		writers[block] = create(fmt.Sprintf("%s-clock%d.jsonl", prefix, block*32))
	}
	compact := create(prefix + ".jsonl")
	manifest := create(prefix + "-manifest.json")
	count, totalFits, caps := 0, 0, 0
	for {
		var r softV120Record
		err := dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if !spikeIndependentIdentity(count, r) || len(r.Steps) != 256 || len(r.Fits) != 8 {
			t.Fatal("source identity/order/shape", count)
		}
		out := spikeContinuousRecord{Phase: r.Phase, Case: r.Case, Index: r.Index, Schedule: r.Schedule}
		var last *spikeCadenceFit
		for block := 0; block < 8; block++ {
			v, err := runSpikeCadenceAt(r, block*32)
			if err != nil {
				t.Fatal(err)
			}
			if len(v.Fits) == 0 {
				t.Fatal("empty fits")
			}
			if last != nil && reflect.DeepEqual(last.Origins, v.Fits[0].Origins) {
				a, b := *last, v.Fits[0]
				a.Clock, b.Clock = 0, 0
				if !reflect.DeepEqual(a, b) {
					t.Fatal("unchanged boundary refit changed forecast")
				}
			}
			last = &v.Fits[len(v.Fits)-1]
			copy(out.P[block*32:], v.P[:])
			out.FitCounts[block] = len(v.Fits)
			for _, fit := range v.Fits {
				if fit.Stop == "iteration-cap" {
					out.Capped[block]++
					caps++
				}
			}
			totalFits += len(v.Fits)
			if err := writers[block].Encode(v); err != nil {
				t.Fatal(err)
			}
		}
		if err := compact.Encode(out); err != nil {
			t.Fatal(err)
		}
		count++
		if count%16 == 0 {
			t.Logf("records=%d fits=%d caps=%d", count, totalFits, caps)
		}
	}
	if count != 672 {
		t.Fatal("incomplete cohort", count)
	}
	if err := manifest.Encode(struct {
		Cohort, SourceSHA256 string
		CodeHashes           map[string]string
		Records, Fits, Caps  int
		Seconds              float64
	}{header.Cohort, sourceHash, codeHashes, count, totalFits, caps, time.Since(started).Seconds()}); err != nil {
		t.Fatal(err)
	}
	for _, f := range outputs {
		if err := f.Sync(); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("complete records=%d forecasts=%d fits=%d caps=%d", count, count*256, totalFits, caps)
}
