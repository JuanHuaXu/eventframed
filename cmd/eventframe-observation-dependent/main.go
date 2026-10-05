package main

import (
	"compress/gzip"
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
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: eventframe-observation-dependent NEW.json.gz")
	}
	// Reserve the artifact before expensive work; never overwrite a prior run.
	f, err := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	paths := []string{
		"docs/experiments/mmm-dependent-v9-protocol.md",
		"internal/observationlearners/dependent.go",
		"internal/observationlearners/retained.go",
		"internal/observationlearners/partial_observer.go",
		"internal/observationlearners/learners.go",
		"internal/observationlearners/experiment.go",
		"internal/observation/controller.go",
		"internal/observation/events.go",
		"internal/observation/forecast.go",
		"internal/observationexperiment/experiment.go",
		"internal/bayes/forecast_mix.go",
		"cmd/eventframe-observation-dependent/main.go",
	}
	hashes, sources := map[string]string{}, map[string]string{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
		sources[p] = string(b)
	}
	runtime.GOMAXPROCS(1)
	records, err := exp.RunDependent()
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(f)
	err = json.NewEncoder(gz).Encode(struct {
		Hashes    map[string]string
		Sources   map[string]string
		GoVersion string
		Records   []exp.DependentRecord
	}{hashes, sources, runtime.Version(), records})
	closeErr := gz.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	fmt.Printf("Completed %d dependent-input streams\n", len(records))
	return f.Close()
}
