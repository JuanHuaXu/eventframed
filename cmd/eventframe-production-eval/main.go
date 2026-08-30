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
	home, err := os.UserHomeDir()
	if err != nil {
		fail(err)
	}
	sessions := flag.String("sessions", filepath.Join(home, ".openclaw", "agents", "main", "sessions"), "read-only OpenClaw main-agent session directory")
	output := flag.String("output", "docs/experiments/production-replay-2026-08-28", "derived artifact directory")
	flag.Parse()
	config := productioneval.Config{
		SessionDir: *sessions, ConfirmationStart: mustTime("2026-08-01T00:00:00Z"),
		DataEnd: mustTime("2026-08-28T00:00:00Z"), RuleFrozenAt: time.Now().UTC(),
	}
	result, err := productioneval.Run(context.Background(), config)
	if err != nil {
		fail(err)
	}
	artifacts := map[string]any{
		"protocol.json": result.Protocol, "design-dataset.json": result.Design, "design-report.json": result.DesignReport,
		"confirmation-dataset.json": result.Confirmation, "confirmation-report.json": result.ConfirmReport,
	}
	for name, value := range artifacts {
		if err := productioneval.WriteJSON(filepath.Join(*output, name), value); err != nil {
			fail(err)
		}
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(struct {
		Output       string `json:"output"`
		DesignCases  int    `json:"design_cases"`
		ConfirmCases int    `json:"confirmation_cases"`
	}{Output: *output, DesignCases: len(result.Design.Cases), ConfirmCases: len(result.Confirmation.Cases)}); err != nil {
		fail(err)
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eventframe-production-eval:", err)
	os.Exit(1)
}
