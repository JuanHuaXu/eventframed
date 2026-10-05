package service

import (
	"os"
	"testing"
)

func TestResearchRecallWorkersV25FocusedRace(t *testing.T) {
	if os.Getenv("EVENTFRAME_RUN_RECALL_WORKERS_V25_RACE") != "1" {
		t.Skip("opt-in focused eight-worker race check")
	}
	r := runResearchRecallWorkersV25(t, true, 8)
	if len(r.offer) != 192 || len(r.gaps) != 191 || r.writes != 256 {
		t.Fatalf("incomplete focused race trial: offers=%d gaps=%d writes=%d", len(r.offer), len(r.gaps), r.writes)
	}
}
