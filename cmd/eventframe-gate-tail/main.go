package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/observationgate"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: eventframe-gate-tail NEW.json.gz NEW-summary.json")
	}
	for _, p := range os.Args[1:] {
		if _, e := os.Stat(p); !os.IsNotExist(e) {
			panic("output exists or inaccessible")
		}
	}
	hashes := map[string]string{}
	for _, p := range []string{"internal/observationgate/tail.go", "internal/observationgate/gate.go", "internal/observationgate/experiment.go", "cmd/eventframe-gate-tail/main.go", "docs/experiments/mmm-gate-tail-v2-protocol.md"} {
		b, e := os.ReadFile(p)
		if e != nil {
			panic(e)
		}
		h := sha256.Sum256(b)
		hashes[p] = hex.EncodeToString(h[:])
	}
	o := observationgate.RunTail()
	o.Hashes = hashes
	f, e := os.OpenFile(os.Args[1], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	z := gzip.NewWriter(f)
	if e = json.NewEncoder(z).Encode(o); e != nil {
		panic(e)
	}
	if e = z.Close(); e != nil {
		panic(e)
	}
	if e = f.Close(); e != nil {
		panic(e)
	}
	o.Records = nil
	f, e = os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		panic(e)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if e = enc.Encode(o); e != nil {
		panic(e)
	}
	if e = f.Close(); e != nil {
		panic(e)
	}
	fmt.Println("grid/anchored90/anchored50 passes:", o.Pass)
}
