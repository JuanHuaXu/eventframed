package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/JuanHuaXu/eventframed/internal/observationrescueexperiment"
)

func main() {
	path := flag.String("output", "", "new .json.gz evidence path")
	summary := flag.String("summary", "", "new summary JSON path")
	flag.Parse()
	if *path == "" || *summary == "" || *path == *summary {
		fmt.Fprintln(os.Stderr, "distinct output and summary paths required")
		os.Exit(2)
	}
	if err := run(*path, *summary); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(path, summary string) error {
	for _, p := range []string{path, summary} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			return fmt.Errorf("refusing existing or inaccessible output: %s", p)
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-antipigeon-v2-protocol.md", "internal/observationrescue/state.go", "internal/observationrescueexperiment/experiment.go", "internal/observation/controller.go", "internal/observation/events.go", "internal/observationexperiment/experiment.go", "internal/bayes/revision.go", "internal/bayes/group.go", "internal/bayes/changepoint.go", "cmd/eventframe-observation-rescue/main.go"} {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	runtime.GOMAXPROCS(1)
	out, err := observationrescueexperiment.Run(func(s string) { fmt.Println(s) })
	if err != nil {
		return err
	}
	out.Hashes = hashes
	out.Runtime = runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(f)
	encodeErr := json.NewEncoder(gz).Encode(out)
	closeErr := gz.Close()
	fileErr := f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if fileErr != nil {
		return fileErr
	}
	for i := range out.Records {
		out.Records[i].Outcomes = nil
		out.Records[i].Audit = nil
		for j := range out.Records[i].Policies {
			out.Records[i].Policies[j].Ticks = nil
		}
	}
	f, err = os.OpenFile(summary, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	encodeErr = enc.Encode(out)
	fileErr = f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	if fileErr != nil {
		return fileErr
	}
	fmt.Printf("Completed %d streams; shift gain=%v; stable protection=%v; overall=%v\n", len(out.Records), out.ShiftGainPassed, out.StableProtectionPassed, out.OverallPassed)
	return nil
}
