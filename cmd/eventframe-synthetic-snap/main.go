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
	out := flag.String("out", "testdata/sheaf-public-facts/cases.jsonl", "output JSONL snap cases")
	manifestPath := flag.String("manifest", "testdata/sheaf-public-facts/manifest.json", "output manifest")
	textManifestPath := flag.String("text-manifest", "testdata/text-public-facts/manifest.json", "public-fact text manifest")
	flag.Parse()

	records, _ := synthetictext.Build()
	cases, manifest, err := synthetictext.BuildSnapCases(records)
	if err != nil {
		fatal(err)
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	for _, testCase := range cases {
		if err := encoder.Encode(testCase); err != nil {
			fatal(err)
		}
	}
	digest := sha256.Sum256(encoded.Bytes())
	manifest.CasesSHA256 = hex.EncodeToString(digest[:])

	textManifestBytes, err := os.ReadFile(*textManifestPath)
	if err != nil {
		fatal(err)
	}
	var textManifest synthetictext.Manifest
	if err := json.Unmarshal(textManifestBytes, &textManifest); err != nil {
		fatal(err)
	}
	manifest.TextCorpusSHA256 = textManifest.CorpusSHA256

	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(*manifestPath), 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(*out, encoded.Bytes(), 0o644); err != nil {
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
	fmt.Printf("wrote %d sheaf-inspired public-fact cases to %s\n", manifest.Cases, *out)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
