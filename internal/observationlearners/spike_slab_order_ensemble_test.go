package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

type spikeOrderRecord struct {
	Order string
	spikePilotRecord
}

func TestSpikeReversePilotCollect(t *testing.T) {
	src, dst := os.Getenv("EVENTFRAME_SPIKE_REVERSE_SOURCE"), os.Getenv("EVENTFRAME_SPIKE_REVERSE_OUTPUT")
	if src == "" {
		t.Skip("explicit research source")
	}
	if dst == "" {
		t.Fatal("output required")
	}
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	dec, enc := json.NewDecoder(in), json.NewEncoder(out)
	var h softV120Artifact
	if err = dec.Decode(&h); err != nil || h.Version != "soft-learners-v120" {
		t.Fatal("header", err)
	}
	count := 0
	for {
		var r softV120Record
		err = dec.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Index != 0 {
			continue
		}
		v, err := runSpikePilotOrdered(r, 128, true)
		if err != nil {
			t.Fatalf("fit%d: %v", count, err)
		}
		if err = enc.Encode(spikeOrderRecord{"descending", v}); err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 84 {
		t.Fatal("count", count)
	}
	if err = out.Sync(); err != nil {
		t.Fatal(err)
	}
	t.Logf("fits=%d forecasts=%d", count, count*32)
}

func BenchmarkSpikeTwoOrderPrediction(b *testing.B) {
	s, masks, _, _ := spikeLargeFixture()
	f, err := fitSpikeFixed(s, masks, 1.0/255, 1, 1024)
	if err != nil {
		b.Fatal(err)
	}
	for i, j := 0, len(masks)-1; i < j; i, j = i+1, j-1 {
		masks[i], masks[j] = masks[j], masks[i]
	}
	r, err := fitSpikeFixed(s, masks, 1.0/255, 1, 1024)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _, err := f.predict(255)
		if err != nil {
			b.Fatal(err)
		}
		q, _, err := r.predict(255)
		if err != nil {
			b.Fatal(err)
		}
		if (p+q)/2 <= 0 {
			b.Fatal("ensemble probability")
		}
	}
}

func BenchmarkSpikeTwoOrderFit(b *testing.B) {
	s, masks, _, _ := spikeLargeFixture()
	reverse := append([]uint16(nil), masks...)
	for i, j := 0, len(reverse)-1; i < j; i, j = i+1, j-1 {
		reverse[i], reverse[j] = reverse[j], reverse[i]
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := fitSpikeFixed(s, masks, 1.0/255, 1, 1024); err != nil {
			b.Fatal(err)
		}
		if _, err := fitSpikeFixed(s, reverse, 1.0/255, 1, 1024); err != nil {
			b.Fatal(err)
		}
	}
}
