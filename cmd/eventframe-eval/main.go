package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/productioneval"
)

func main() {
	input := flag.String("input", "", "evaluation dataset JSON")
	baseline := flag.String("baseline", "", "override baseline variant for a production-evaluation artifact")
	output := flag.String("output", "", "optional report JSON output path")
	flag.Parse()
	if *input == "" {
		fail("-input is required")
	}
	payload, err := os.ReadFile(*input)
	if err != nil {
		fail(err.Error())
	}
	var report evaluation.Report
	if *baseline != "" {
		var artifact productioneval.Artifact
		if err := json.Unmarshal(payload, &artifact); err != nil {
			fail(err.Error())
		}
		report, err = productioneval.EvaluateArtifactWithBaseline(artifact, *baseline)
	} else {
		var dataset evaluation.Dataset
		if err := json.Unmarshal(payload, &dataset); err != nil {
			fail(err.Error())
		}
		report, err = evaluation.Evaluate(dataset)
	}
	if err != nil {
		fail(err.Error())
	}
	if *output != "" {
		if err := productioneval.WriteJSON(*output, report); err != nil {
			fail(err.Error())
		}
		return
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fail(err.Error())
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, "eventframe-eval:", message)
	os.Exit(1)
}
