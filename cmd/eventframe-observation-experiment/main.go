// eventframe-observation-experiment runs the frozen synthetic MMM pilot only.
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

	"github.com/JuanHuaXu/eventframed/internal/observationexperiment"
)

func main() {
	path := flag.String("output", "", "new .json.gz evidence file")
	summary := flag.String("summary", "", "new compact summary JSON file")
	flag.Parse()
	if *path == "" || *summary == "" {
		fmt.Fprintln(os.Stderr, "output and summary are required")
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
			return fmt.Errorf("refusing existing or inaccessible output %s", p)
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-attention-v1-protocol.md", "internal/observation/controller.go", "internal/observation/events.go", "internal/observationexperiment/experiment.go", "cmd/eventframe-observation-experiment/main.go"} {
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	runtime.GOMAXPROCS(1)
	out, err := observationexperiment.Run()
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
	count := len(out.Rows)
	out.Rows = nil
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
	fmt.Printf("Completed %d paired-policy predictions; primary passed=%v; confirmation harm flag=%v\n", count, out.PrimaryPassed, out.AnyConfirmationHarm)
	return nil
}
