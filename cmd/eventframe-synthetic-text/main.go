package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JuanHuaXu/eventframed/internal/synthetictext"
)

func main() {
	out := flag.String("out", "testdata/text-public-facts/corpus.jsonl", "output JSONL corpus")
	manifestPath := flag.String("manifest", "testdata/text-public-facts/manifest.json", "output manifest")
	flag.Parse()

	records, manifest := synthetictext.Build()
	var corpus bytes.Buffer
	encoder := json.NewEncoder(&corpus)
	encoder.SetEscapeHTML(false)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			fatal(err)
		}
	}
	digest := sha256.Sum256(corpus.Bytes())
	manifest.CorpusSHA256 = hex.EncodeToString(digest[:])

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*manifestPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, corpus.Bytes(), 0o644); err != nil {
		fatal(err)
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fatal(err)
	}
	manifestBytes = append(manifestBytes, '\n')
	if err := os.WriteFile(*manifestPath, manifestBytes, 0o644); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %d public-fact turns across %d sessions to %s\n", manifest.Turns, manifest.Sessions, *out)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
