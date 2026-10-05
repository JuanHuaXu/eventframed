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
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 3 || os.Args[1] == os.Args[2] {
		return fmt.Errorf("usage: eventframe-observation-learners NEW.json.gz NEW-summary.json")
	}
	for _, p := range os.Args[1:] {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			return fmt.Errorf("output exists or inaccessible: %s", p)
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-learners-v5-protocol.md", "internal/observationlearners/learners.go", "internal/observationlearners/experiment.go", "internal/observation/controller.go", "internal/observation/events.go", "internal/observation/forecast.go", "internal/observationexperiment/experiment.go", "internal/bayes/forecast_mix.go", "cmd/eventframe-observation-learners/main.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	runtime.GOMAXPROCS(1)
	out, e := exp.Run(func(s string) { fmt.Println(s) })
	if e != nil {
		return e
	}
	out.Hashes = hashes
	f, e := os.OpenFile(os.Args[1], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	gz := gzip.NewWriter(f)
	err := json.NewEncoder(gz).Encode(out)
	a, b := gz.Close(), f.Close()
	if err != nil {
		return err
	}
	if a != nil {
		return a
	}
	if b != nil {
		return b
	}
	for i := range out.Records {
		out.Records[i].Ticks = nil
	}
	f, e = os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(out)
	e = f.Close()
	if err != nil {
		return err
	}
	if e != nil {
		return e
	}
	fmt.Printf("Completed %d streams; verdicts=%+v\n", len(out.Records), out.Verdicts)
	return nil
}
