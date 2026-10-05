package main

import (
	"strings"
	"testing"
)

func TestDecoder(t *testing.T) {
	raw := "{\"Dir\":\"a } { \\\"\"}\n{\"Dir\":\"b\"}\n"
	rows, err := decodePackages(strings.NewReader(raw))
	if err != nil || len(rows) != 2 {
		t.Fatal(len(rows), err)
	}
	for _, bad := range []string{raw + "{", raw + "junk", `{"Dir":"unterminated}`} {
		if _, err := decodePackages(strings.NewReader(bad)); err == nil {
			t.Fatal("truncated/invalid input accepted")
		}
	}
}
