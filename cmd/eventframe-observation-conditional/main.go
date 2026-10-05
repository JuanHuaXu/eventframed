package main

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	exp "github.com/JuanHuaXu/eventframed/internal/observationlearners"
	"os"
	"runtime"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: eventframe-observation-conditional V9.json.gz NEW.json.gz")
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		return err
	}
	inputHash := sha256.Sum256(b)
	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	var input struct {
		Hashes, Sources map[string]string
		Records         []exp.DependentRecord
	}
	if err = json.NewDecoder(gz).Decode(&input); err != nil {
		return err
	}
	for p, source := range input.Sources {
		h := sha256.Sum256([]byte(source))
		if hex.EncodeToString(h[:]) != input.Hashes[p] {
			return fmt.Errorf("source integrity: %s", p)
		}
	}
	sources := map[string]string{}
	for _, p := range []string{"docs/experiments/mmm-conditional-v10-protocol.md", "internal/observationlearners/conditional.go", "internal/observationlearners/conditional_experiment.go", "internal/observationlearners/learners.go", "internal/observationlearners/partial_observer.go", "cmd/eventframe-observation-conditional/main.go"} {
		v, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		sources[p] = string(v)
	}
	out, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer out.Close()
	runtime.GOMAXPROCS(1)
	var records []exp.ConditionalRecord
	for _, r := range input.Records {
		v, e := exp.RunConditionalRecord(r)
		if e != nil {
			return e
		}
		records = append(records, v)
	}
	z := gzip.NewWriter(out)
	err = json.NewEncoder(z).Encode(struct {
		InputSHA256 string
		Sources     map[string]string
		Records     []exp.ConditionalRecord
	}{hex.EncodeToString(inputHash[:]), sources, records})
	ce := z.Close()
	if err != nil {
		return err
	}
	if ce != nil {
		return ce
	}
	fmt.Printf("Computed %d conditional diagnostics\n", len(records))
	return out.Close()
}
