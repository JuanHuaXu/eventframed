// research-public-pool builds label-free document-level native index inputs.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
	"github.com/JuanHuaXu/eventframed/internal/researchpublicpool"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: research-public-pool CORPUS.json NEW.json")
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	records, err := researchpublicframe.Decode(b)
	if err != nil {
		panic(err)
	}
	start := time.Now()
	r, err := researchpublicpool.New(context.Background(), records, researchpublicframe.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)}, "public-scientific")
	if err != nil {
		panic(err)
	}
	elapsed := time.Since(start)
	f, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	if err = enc.Encode(r.Entries()); err != nil {
		panic(err)
	}
	if err = f.Close(); err != nil {
		panic(err)
	}
	fmt.Printf("documents=%d build_ns=%d no_labels=true\n", len(r.Entries()), elapsed.Nanoseconds())
}
