package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/JuanHuaXu/eventframed/internal/observationgate"
)

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: eventframe-observation-gate NEW.json.gz NEW-summary.json")
	}
	for _, p := range os.Args[1:] {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			return fmt.Errorf("output exists or inaccessible: %s", p)
		}
	}
	runtime.GOMAXPROCS(1)
	hashes := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-gate-v1-protocol.md", "internal/observationgate/gate.go", "internal/observationgate/experiment.go", "cmd/eventframe-observation-gate/main.go"} {
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	o := observationgate.Run()
	o.Hashes = hashes
	f, e := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	z := gzip.NewWriter(f)
	if e = json.NewEncoder(z).Encode(o); e != nil {
		f.Close()
		return e
	}
	if e = z.Close(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	o.Records = nil
	f, e = os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return e
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if e = enc.Encode(o); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	fmt.Printf("Complete: grid/adaptive/hedge passes=%v\n", o.Pass)
	return nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
