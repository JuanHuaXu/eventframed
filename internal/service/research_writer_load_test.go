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
)

func researchWriterLoadCell(t *testing.T, trial, requests, writes int, active, four bool) researchGuardLoadResult {
	return researchWriterGroupCell(t, trial, requests, writes, active, four, 0)
}

func researchWriterGroupCell(t *testing.T, trial, requests, writes int, active, four bool, cap int) researchGuardLoadResult {
	t.Helper()
	mode := "off"
	if active {
		mode = "group4postverify"
	}
	r := researchGuardLoadGroupCapArm(t, mode, trial, requests, writes, true, active, active, active, active, active, active, &researchLoadSchedule{5 * time.Millisecond, 10 * time.Millisecond}, false, four, cap)
	checkQueuedGuardLoad(t, r, writes)
	if err := researchCheckOffered(r, writes); err != nil {
		t.Fatal(err)
	}
	if r.FourNativeWriters != four || r.ResolvedSourceAdmissions != active {
		t.Fatal("wrong writer experiment arm")
	}
	if active && (r.Accepted == 0 || r.Admits == 0 || r.Admits != r.Discards) {
		t.Fatal("active source path missing")
	}
	if r.ResearchGroupCap != cap {
		t.Fatal("wrong group cap")
	}
	for _, n := range r.GroupSizes {
		if cap != 0 && n > cap {
			t.Fatal("group cap exceeded")
		}
	}
	return r
}

func TestResearchWriterLoadAccounting(t *testing.T) {
	for _, active := range []bool{false, true} {
		for _, four := range []bool{false, true} {
			researchWriterLoadCell(t, 0, 8, 4, active, four)
		}
	}
}

func TestResearchWriterLoadExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_WRITER_LOAD_ARTIFACT")
	if path == "" {
		t.Skip("opt-in scheduled native writer comparison")
	}
	runResearchWriterLoad(t, path, "")
}

func TestResearchWriterProfileExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_WRITER_PROFILE_ARTIFACT")
	if path == "" {
		t.Skip("opt-in arm-specific profiling")
	}
	arm := os.Getenv("EVENTFRAME_WRITER_PROFILE_ARM")
	if arm != "off4" && arm != "active4" {
		t.Fatal("profile arm must be off4 or active4")
	}
	runResearchWriterLoad(t, path, arm)
}

func runResearchWriterLoad(t *testing.T, path, profile string) {
	runResearchWriterGroupLoad(t, path, profile, false)
}

func TestResearchGroupCapAccounting(t *testing.T) {
	for _, cap := range []int{1, 2, 4} {
		researchWriterGroupCell(t, 0, 8, 4, true, true, cap)
	}
}

func TestResearchGroupCapExperiment(t *testing.T) {
	path := os.Getenv("EVENTFRAME_GROUP_CAP_ARTIFACT")
	if path == "" {
		t.Skip("opt-in group scheduling experiment")
	}
	runResearchWriterGroupLoad(t, path, "", true)
}

func runResearchWriterGroupLoad(t *testing.T, path, profile string, groups bool) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sources, hashes := map[string]string{}, map[string]string{}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	capture := func(name string) error {
		b, e := os.ReadFile(filepath.Join(root, name))
		if e != nil {
			return e
		}
		sources[name] = string(b)
		h := sha256.Sum256(b)
		hashes[name] = hex.EncodeToString(h[:])
		return nil
	}
	// Capture internal implementation and test sources, not private test corpora.
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
	for _, name := range []string{"go.mod", "go.sum", "docs/experiments/mmm-writer-load-v62-protocol.md"} {
		if err = capture(name); err != nil {
			t.Fatal(err)
		}
	}
	enc := json.NewEncoder(f)
	if groups {
		if profile != "" {
			t.Fatal("group study is not a profile arm")
		}
		if err = capture("docs/experiments/mmm-group-cap-v64-protocol.md"); err != nil {
			t.Fatal(err)
		}
	}
	cells := 12
	if profile != "" {
		cells = 3
		if err = capture("docs/experiments/mmm-writer-profile-v63-protocol.md"); err != nil {
			t.Fatal(err)
		}
	}
	if err = enc.Encode(map[string]any{"Kind": "header", "Cells": cells, "GroupCapStudy": groups, "ProfileArm": profile, "Go": runtime.Version(), "GOMAXPROCS": runtime.GOMAXPROCS(0), "Sources": sources, "Hashes": hashes}); err != nil {
		t.Fatal(err)
	}
	for trial := 0; trial < 3; trial++ {
		for order := 0; order < 4; order++ {
			arm := (trial + order) % 4
			if (profile == "off4" && arm != 1) || (profile == "active4" && arm != 3) {
				continue
			}
			var r researchGuardLoadResult
			if groups {
				r = researchWriterGroupCell(t, trial, 192, 96, arm != 0, true, []int{0, 4, 2, 1}[arm])
			} else {
				r = researchWriterLoadCell(t, trial, 192, 96, arm >= 2, arm%2 == 1)
			}
			if err = enc.Encode(r); err != nil {
				t.Fatal(err)
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			t.Logf("trial%d active%t four%t cap%d accepted%d dropped%d expired%d", trial, r.SourceOwner, r.FourNativeWriters, r.ResearchGroupCap, r.Accepted, r.Dropped, r.WaitExpired)
		}
	}
}
