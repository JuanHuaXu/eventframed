package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/productioneval"
)

func main() {
	root := flag.String("corpus", ".eventframed/rc2-chatgpt", "private ChatGPT corpus directory")
	output := flag.String("output", ".eventframed/rc2-chatgpt/results", "private derived artifact directory")
	ablations := flag.Bool("ablations", false, "run the frozen ablation family")
	flag.Parse()
	result, err := productioneval.RunChatGPT(context.Background(), productioneval.ChatGPTConfig{
		RawDir: filepath.Join(*root, "raw"), ManifestPath: filepath.Join(*root, "manifest.local.json"),
		RuleFrozenAt: time.Now().UTC(), Ablations: *ablations,
	})
	if err != nil {
		fail(err)
	}
	artifacts := map[string]any{
		"protocol.json": result.Protocol, "design-dataset.json": result.Design, "design-report.json": result.DesignReport,
		"confirmation-dataset.json": result.Confirmation, "confirmation-report.json": result.ConfirmReport,
		"design-control-report.json": result.DesignControlReport, "confirmation-control-report.json": result.ConfirmControlReport,
	}
	for name, value := range artifacts {
		if err := productioneval.WriteJSON(filepath.Join(*output, name), value); err != nil {
			fail(err)
		}
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"output": *output, "design_cases": len(result.Design.Cases), "confirmation_cases": len(result.Confirmation.Cases),
	})
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eventframe-chatgpt-eval:", err)
	os.Exit(1)
}
