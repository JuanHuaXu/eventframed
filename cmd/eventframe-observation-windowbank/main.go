package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	exp "github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"os"
	"runtime"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: eventframe-observation-windowbank NEW.jsonl.gz")
	}
	f, e := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	paths := []string{"docs/experiments/mmm-windowbank-v15-protocol.md", "internal/observationlearners/windowbank_experiment.go", "internal/observationlearners/breadth_experiment.go", "internal/observationlearners/subset.go", "internal/observationlearners/conditional.go", "internal/observationlearners/conditional_observer.go", "internal/observationlearners/dependent.go", "internal/observationlearners/experiment.go", "internal/observationlearners/learners.go", "internal/observationlearners/partial_observer.go", "internal/observation/controller.go", "internal/observation/forecast.go", "internal/observation/events.go", "internal/observationexperiment/experiment.go", "internal/bayes/forecast_mix.go", "cmd/eventframe-observation-windowbank/main.go"}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		sources[p] = string(b)
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	z := gzip.NewWriter(f)
	enc := json.NewEncoder(z)
	if e = enc.Encode(struct {
		Sources, Hashes map[string]string
		ExpectedRecords int
	}{sources, hashes, 1728}); e != nil {
		return e
	}
	runtime.GOMAXPROCS(1)
	count := 0
	e = exp.RunWindowBank(func(r exp.BreadthRecord) error { count++; return enc.Encode(r) })
	ce := z.Close()
	if e != nil {
		return e
	}
	if ce != nil {
		return ce
	}
	fmt.Printf("Completed %d window-bank streams\n", count)
	return f.Close()
}
