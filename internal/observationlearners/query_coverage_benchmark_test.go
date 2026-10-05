package observationlearners

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"
)

func BenchmarkCoverageQuery(b *testing.B) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		b.Skip("explicit source required")
	}
	f, err := os.Open(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		b.Fatal(err)
	}
	var input softV120Record
	for {
		if err := d.Decode(&input); err == io.EOF {
			b.Fatal("fixture missing")
		} else if err != nil {
			b.Fatal(err)
		}
		if input.Phase == 1 && input.Case == 0 && input.Index == 0 && input.Schedule == 1 {
			break
		}
	}
	reference := runCoverageQuery(context.Background(), input)
	if reference.Error != "" {
		b.Fatal(reference.Error)
	}
	b.Run("Original8Batch", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := decideRegimeOutcome(input); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("FullResponseExplicitRefits", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := runCoverageQuery(context.Background(), input)
			if r.Error != "" || r.Selected != reference.Selected {
				b.Fatal("reference changed", r.Error)
			}
		}
	})
	b.Run("TwoHistogramsPrecomputedResponses", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, branch := range reference.Branches {
				for _, w := range reference.Weights {
					if _, err := coverageQueryGain(reference.Base, branch, w); err != nil {
						b.Fatal(err)
					}
				}
			}
		}
	})
}
