package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/researchpublicationstore"
)

type researchGateTimingCell struct {
	Trial            int
	Active, Measured bool
	Result           researchGuardLoadResult
	Spans            []researchpublicationstore.GateSpan
	Dropped          uint64
}

func researchGateTimingRunCell(t *testing.T, trial, requests, writes int, active, measured bool) researchGateTimingCell {
	t.Helper()
	var recorder *researchpublicationstore.GateRecorder
	if measured {
		var err error
		recorder, err = researchpublicationstore.NewGateRecorder(4096)
		if err != nil {
			t.Fatal(err)
		}
	}
	mode := "off"
	if active {
		mode = "group4postverify"
	}
	r := researchGuardLoadMeasuredArm(t, mode, trial, requests, writes, true, active, active, active, active, active, active, &researchLoadSchedule{5 * time.Millisecond, 10 * time.Millisecond}, !active, true, 0, recorder)
	checkQueuedGuardLoad(t, r, writes)
	if err := researchCheckOffered(r, writes); err != nil {
		t.Fatal(err)
	}
	if r.ResolvedSourceAdmissions != active || !r.FourNativeWriters {
		t.Fatal("wrong timing arm")
	}
	out := researchGateTimingCell{Trial: trial, Active: active, Measured: measured, Result: r}
	if measured {
		out.Spans, out.Dropped = recorder.Snapshot()
		if out.Dropped != 0 {
			t.Fatal("trace overflow")
		}
		ingestions, guards := 0, 0
		for _, s := range out.Spans {
			if s.WaitNS < 0 || s.HeldNS < 0 || (!s.Entered && s.HeldNS != 0) {
				t.Fatal("invalid span", s)
			}
			if s.Kind == "ingestion" && s.Entered {
				ingestions++
			}
			if s.Kind == "asof" && s.Entered {
				guards++
			}
		}
		if ingestions != writes {
			t.Fatalf("ingestion coverage %d/%d", ingestions, writes)
		}
		if active && guards == 0 {
			t.Fatal("missing active guard trace")
		}
	}
	return out
}

func TestResearchGateTimingAccounting(t *testing.T) {
	for _, active := range []bool{false, true} {
		for _, measured := range []bool{false, true} {
			researchGateTimingRunCell(t, 0, 8, 4, active, measured)
		}
	}
}

func TestResearchGateTimingExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_GATE_TIMING_ARTIFACT")
	if path == "" {
		t.Skip("opt-in isolated gate timing")
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	sources, hashes := map[string]string{}, map[string]string{}
	capture := func(name string) error {
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			return e
		}
		sum := sha256.Sum256(b)
		sources[name] = string(b)
		hashes[name] = hex.EncodeToString(sum[:])
		return nil
	}
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			name, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			return capture(name)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "go.sum", "docs/experiments/mmm-gate-timing-v1-contract.md"} {
		if err = capture(name); err != nil {
			t.Fatal(err)
		}
	}
	enc := json.NewEncoder(f)
	if err = enc.Encode(map[string]any{"Kind": "header", "Sources": sources, "Hashes": hashes, "Go": runtime.Version(), "GOMAXPROCS": runtime.GOMAXPROCS(0), "Cells": 12}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 4; order++ {
			arm := (trial + order) % 4
			r := researchGateTimingRunCell(t, trial, 192, 96, arm >= 2, arm%2 == 1)
			if err = enc.Encode(r); err != nil {
				t.Fatal(err)
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Logf("trial%d active%t measured%t accepted%d dropped%d traces%d", trial, r.Active, r.Measured, r.Result.Accepted, r.Result.Dropped, len(r.Spans))
		}
	}
}
