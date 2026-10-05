package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	exp "github.com/JuanHuaXu/eventframed/internal/observationfalsification"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 3 || os.Args[1] == os.Args[2] {
		return fmt.Errorf("usage: eventframe-observation-falsification NEW.json.gz NEW-summary.json")
	}
	for _, p := range os.Args[1:] {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			return fmt.Errorf("output exists or inaccessible: %s", p)
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-falsification-v4-protocol.md", "internal/observationfalsification/experiment.go", "internal/observationpreserved/state.go", "internal/observationpreserved/experiment.go", "internal/observation/controller.go", "internal/observation/events.go", "internal/observation/forecast.go", "internal/observationexperiment/experiment.go", "internal/observationrescue/state.go", "internal/bayes/forecast_mix.go", "internal/bayes/revision.go", "internal/bayes/group.go", "internal/bayes/changepoint.go", "cmd/eventframe-observation-falsification/main.go"} {
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
	fmt.Printf("Completed %d streams; added shift benefit=%v stable=%v overall=%v\n", len(out.Records), out.ShiftPass, out.StablePass, out.OverallPass)
	return nil
}
