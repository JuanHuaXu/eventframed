package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestCalibrationProjectionBoundary(t *testing.T) {
	p := projection{Partition: "calibration"}
	for i := 0; i < 180; i++ {
		p.Queries = append(p.Queries, query{ID: fmt.Sprint(i), Text: "source query"})
	}
	b, _ := json.Marshal(p)
	if _, e := decodeProjection(b); e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*projection){
		func(p *projection) { p.Partition = "fit" },
		func(p *projection) { p.Queries = p.Queries[:179] },
		func(p *projection) { p.Queries[1].ID = p.Queries[0].ID },
		func(p *projection) { p.Queries[0].Text = "" },
		func(p *projection) { p.Queries[0].Text = strings.Repeat("x", 4097) },
	} {
		copy := p
		copy.Queries = append([]query(nil), p.Queries...)
		mutate(&copy)
		bad, _ := json.Marshal(copy)
		if _, e := decodeProjection(bad); e == nil {
			t.Fatal("invalid projection accepted")
		}
	}
	for _, bad := range [][]byte{
		append(append([]byte(nil), b...), []byte(` {}`)...),
		[]byte(strings.Replace(string(b), `"partition":`, `"targets":[],"partition":`, 1)),
		[]byte(strings.Replace(string(b), `"text":`, `"positive":[],"text":`, 1)),
	} {
		if _, e := decodeProjection(bad); e == nil {
			t.Fatal("annotations or trailing JSON accepted")
		}
	}
}

func TestCalibrationModelBoundary(t *testing.T) {
	good := `{"primary":{"weights":[0,0,0,0,0,0,0,0]},"secondary":{"weights":[1,0,0,0,0,0,0,0]}}`
	if _, e := decodeModels([]byte(good)); e != nil {
		t.Fatal(e)
	}
	for _, bad := range []string{
		`{}`, strings.Replace(good, "[0,0,0,0,0,0,0,0]", "[0,0,0,0,0,0,0]", 1),
		strings.Replace(good, "[0,0,0,0,0,0,0,0]", "[0,0,0,0,0,0,0,0,1]", 1),
		strings.Replace(good, "[0,0,0,0,0,0,0,0]", "[0,0,0,0,0,0,1,0]", 1),
		strings.Replace(good, "[0,0,0,0,0,0,0,0]", "[5,0,0,0,0,0,0,0]", 1),
		strings.Replace(good, "[0,0,0,0,0,0,0,0]", "[1e999,0,0,0,0,0,0,0]", 1),
		strings.Replace(good, `"primary":`, `"targets":[],"primary":`, 1), good + `{}`,
	} {
		if _, e := decodeModels([]byte(bad)); e == nil {
			t.Fatal("invalid model accepted")
		}
	}
}
