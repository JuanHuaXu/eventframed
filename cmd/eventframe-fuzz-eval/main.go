package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/fuzzexperiment"
)

func main() {
	input := flag.String("input", "testdata/text-public-facts/corpus.jsonl", "public-fact JSONL corpus")
	split := flag.String("split", "confirmation", "dataset split to evaluate")
	output := flag.String("output", "", "optional JSON result path")
	flag.Parse()
	result, err := fuzzexperiment.Run(*input, *split)
	if err != nil {
		log.Fatal(err)
	}
	if *output != "" {
		if err := fuzzexperiment.Write(*output, result); err != nil {
			log.Fatal(err)
		}
	}
	encoder := jsonEncoder(os.Stdout)
	if err := encoder(result); err != nil {
		log.Fatal(err)
	}
}

func jsonEncoder(file *os.File) func(any) error {
	return func(value any) error {
		payload, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(file, string(payload))
		return err
	}
}
