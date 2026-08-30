package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/rerankexperiment"
)

func main() {
	datasetPath := flag.String("dataset", "", "write generated design and confirmation data to this path")
	reportPath := flag.String("report", "", "write paired design and confirmation results to this path")
	flag.Parse()
	if *datasetPath == "" || *reportPath == "" {
		fail(fmt.Errorf("-dataset and -report are required"))
	}
	design, err := rerankexperiment.Generate(rerankexperiment.BlockConfig{Name: "design", Seed: 29082901, BidirectionalRepeats: 2, RetentionRepeats: 10, EnvelopeRepeats: 10})
	if err != nil {
		fail(err)
	}
	confirmation, err := rerankexperiment.Generate(rerankexperiment.BlockConfig{Name: "confirmation", Seed: 29082902, BidirectionalRepeats: 8, RetentionRepeats: 30, EnvelopeRepeats: 50})
	if err != nil {
		fail(err)
	}
	suite := rerankexperiment.SuiteDataset{SchemaVersion: rerankexperiment.SchemaVersion, Design: design, Confirmation: confirmation}
	if err := writeJSON(*datasetPath, suite); err != nil {
		fail(err)
	}
	designReport, err := rerankexperiment.Run(context.Background(), design)
	if err != nil {
		fail(err)
	}
	confirmationReport, err := rerankexperiment.Run(context.Background(), confirmation)
	if err != nil {
		fail(err)
	}
	report := rerankexperiment.SuiteReport{SchemaVersion: rerankexperiment.SchemaVersion, Design: designReport, Confirmation: confirmationReport}
	if err := writeJSON(*reportPath, report); err != nil {
		fail(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fail(err)
	}
}

func writeJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o644)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eventframe-rerank-experiment:", err)
	os.Exit(1)
}
