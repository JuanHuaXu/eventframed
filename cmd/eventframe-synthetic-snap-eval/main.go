package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JuanHuaXu/eventframed/internal/snapexperiment"
	"github.com/JuanHuaXu/eventframed/internal/synthetictext"
)

func main() {
	corpusPath := flag.String("corpus", "testdata/text-public-facts/corpus.jsonl", "public-fact text JSONL")
	casesPath := flag.String("cases", "testdata/sheaf-public-facts/cases.jsonl", "sheaf-inspired case JSONL")
	outPath := flag.String("out", "testdata/sheaf-public-facts/results.json", "experiment report")
	dimension := flag.Int("embedding-dimension", 256, "feature-hash embedding dimension")
	recallK := flag.Int("recall-k", 16, "retrieval frontier size")
	packK := flag.Int("pack-k", 3, "packed context size")
	graphWeight := flag.Float64("graph-weight", .05, "graph compatibility rank weight")
	flag.Parse()

	var records []synthetictext.Record
	if err := readJSONL(*corpusPath, &records); err != nil {
		fatal(err)
	}
	var cases []synthetictext.SnapCase
	if err := readJSONL(*casesPath, &cases); err != nil {
		fatal(err)
	}
	options := snapexperiment.Options{EmbeddingDimension: *dimension, RecallK: *recallK, PackK: *packK, GraphWeight: *graphWeight}
	report, err := snapexperiment.Run(context.Background(), records, cases, options)
	if err != nil {
		fatal(err)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fatal(err)
	}
	encoded = append(encoded, '\n')
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, encoded, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %d fixed-topology query outcomes to %s\n", len(report.Results), *outPath)
	fmt.Printf("rescue=%s relevant_delta=%+.3f obsolete_delta=%+.3f mrr_delta=%+.3f\n",
		report.RescueVerdict.Status, report.RescueVerdict.ObservedRelevantHitDelta,
		report.RescueVerdict.ObservedObsoleteHitDelta, report.RescueVerdict.ObservedMRRDelta)
	for _, value := range report.Aggregates {
		if value.Split == "confirmation" {
			fmt.Printf("%-18s hit=%.3f obsolete=%.3f mrr=%.3f graph=%.3f n=%d\n", value.Variant, value.HitRate, value.ObsoleteHitRate, value.MeanReciprocalRank, value.GraphApplicationRate, value.Queries)
		}
	}
}

func readJSONL[T any](path string, values *[]T) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var value T
		if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
		*values = append(*values, value)
	}
	return scanner.Err()
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
