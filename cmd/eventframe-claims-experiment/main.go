package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/claimsexperiment"
)

func main() {
	output := flag.String("output", "", "optional path for the JSON report")
	seedBase := flag.Int64("seed-base", 982451653, "base seed for stochastic experiment trajectories")
	runRole := flag.String("run-role", claimsexperiment.RunRoleExploratory, "run role: exploratory, design, or confirmation")
	designSeedBase := flag.Int64("design-seed-base", 0, "design seed base required for confirmation runs")
	flag.Parse()
	options := claimsexperiment.RunOptions{SeedBase: *seedBase, Role: *runRole}
	designSeedProvided := false
	flag.Visit(func(used *flag.Flag) {
		if used.Name == "design-seed-base" {
			designSeedProvided = true
		}
	})
	if designSeedProvided {
		options.DesignSeedBase = designSeedBase
	}
	report, err := claimsexperiment.RunSuiteWithOptions(context.Background(), options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded = append(encoded, '\n')
	if *output != "" {
		if err := os.WriteFile(*output, encoded, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if _, err := os.Stdout.Write(encoded); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
