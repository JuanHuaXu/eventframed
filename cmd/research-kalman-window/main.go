package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	exp "github.com/JuanHuaXu/eventframed/internal/observationlearners"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 3 || (os.Args[1] != "design" && os.Args[1] != "confirmation") {
		return fmt.Errorf("usage: research-kalman-window design|confirmation NEW.jsonl")
	}
	f, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	paths := []string{
		"docs/experiments/mmm-kalman-window-v1-protocol.md",
		"internal/observationlearners/kalman_window.go",
		"internal/observationlearners/experiment.go",
		"internal/observationlearners/learners.go",
		"internal/observation/controller.go",
		"internal/observation/forecast.go",
		"cmd/research-kalman-window/main.go",
		"research/kalman-window-v1-summary.mjs",
	}
	hashes := make(map[string]string, len(paths))
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(body)
		hashes[path] = hex.EncodeToString(sum[:])
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(map[string]any{"kind": "manifest", "split": os.Args[1], "hashes": hashes}); err != nil {
		return err
	}
	runtime.GOMAXPROCS(1)
	for _, scenario := range []int{0, 2, 5, 7, 8} {
		for fit := 0; fit < 4; fit++ {
			fitID := fit
			if os.Args[1] == "confirmation" {
				fitID += 100
			}
			base, err := exp.Base(exp.Scenarios[scenario], scenario, fitID)
			if err != nil {
				return err
			}
			for stream := 0; stream < 8; stream++ {
				r, err := exp.RunKalmanWindowWithBase(base, os.Args[1], scenario, fit, stream)
				if err != nil {
					return err
				}
				if err := enc.Encode(r); err != nil {
					return err
				}
			}
		}
	}
	return f.Sync()
}
