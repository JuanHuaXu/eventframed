// This post-run auditor uses the standard JSON decoder, not textual framing.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func decodePackages(r io.Reader) ([]json.RawMessage, error) {
	d := json.NewDecoder(r)
	out := []json.RawMessage{}
	for {
		var p json.RawMessage
		err := d.Decode(&p)
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
}

func main() {
	args := append([]string{"list", "-deps", "-test", "-json"}, os.Args[1:]...)
	c := exec.Command("go", args...)
	c.Stderr = os.Stderr
	raw, err := c.Output()
	if err == nil {
		var rows []json.RawMessage
		rows, err = decodePackages(bytes.NewReader(raw))
		if err == nil {
			err = json.NewEncoder(os.Stdout).Encode(rows)
		}
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
