package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/JuanHuaXu/eventframed/internal/translationexperiment"
)

func main() {
	split := flag.String("split", "confirmation", "experiment split: design or confirmation")
	casesOnly := flag.Bool("cases", false, "emit public-grounded input cases instead of results")
	flag.Parse()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if *casesOnly {
		cases, err := translationexperiment.Cases(*split)
		if err != nil {
			fail(err)
		}
		if err := encoder.Encode(cases); err != nil {
			fail(err)
		}
		return
	}
	result, err := translationexperiment.Run(*split)
	if err != nil {
		fail(err)
	}
	if err := encoder.Encode(result); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
