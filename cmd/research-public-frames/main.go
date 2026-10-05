// research-public-frames is a label-free local conversion, not a ranking run.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicframe"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: research-public-frames CORPUS.json NEW.json")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	info, err := f.Stat()
	if err != nil || info.Size() > researchpublicframe.MaxInputBytes {
		panic("public source file bound")
	}
	f.Close()
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	r, err := researchpublicframe.Decode(b)
	if err != nil {
		panic(err)
	}
	start := time.Now()
	docs, err := researchpublicframe.Convert(context.Background(), r, researchpublicframe.Config{TenantID: "research-scifact", SessionID: "public-import", ImportedAt: time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		panic(err)
	}
	elapsed := time.Since(start)
	out, err := os.OpenFile(os.Args[2], os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	e := json.NewEncoder(out)
	e.SetEscapeHTML(false)
	if err = e.Encode(docs); err != nil {
		panic(err)
	}
	if err = out.Close(); err != nil {
		panic(err)
	}
	frames := 0
	for _, d := range docs {
		frames += len(d.Spans)
	}
	fmt.Printf("documents=%d frames=%d convert_ns=%d no_labels_no_models=true\n", len(docs), frames, elapsed.Nanoseconds())
}
