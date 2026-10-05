package productioneval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCodexUsesCompletedToolInputsWithoutExportingThem(t *testing.T) {
	dir := t.TempDir()
	for sessionIndex := 0; sessionIndex < 2; sessionIndex++ {
		start := time.Date(2026, 5, 1+sessionIndex, 0, 0, 0, 0, time.UTC)
		path := filepath.Join(dir, fmt.Sprintf("session-%d.jsonl", sessionIndex))
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		writeCodexRecord(t, file, start, "session_meta", map[string]any{"type": "session_meta", "id": fmt.Sprintf("session-%d", sessionIndex)})
		for turn := 0; turn < 12; turn++ {
			base := start.Add(time.Duration(turn) * time.Minute)
			secret := "/private/work/project/service.go"
			writeCodexRecord(t, file, base, "event_msg", map[string]any{"type": "user_message", "message": "please inspect " + secret})
			writeCodexRecord(t, file, base.Add(time.Second), "response_item", map[string]any{"type": "function_call", "call_id": fmt.Sprintf("call-%d", turn), "arguments": `{"cmd":"sed -n '1,20p' ` + secret + `"}`})
			writeCodexRecord(t, file, base.Add(2*time.Second), "response_item", map[string]any{"type": "function_call_output", "call_id": fmt.Sprintf("call-%d", turn), "output": "ok"})
			writeCodexRecord(t, file, base.Add(3*time.Second), "event_msg", map[string]any{"type": "task_complete", "last_agent_message": "inspected " + secret})
		}
		_ = file.Close()
	}
	result, err := RunCodex(context.Background(), CodexConfig{
		SessionDirs: []string{dir}, DataEnd: time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC),
		RuleFrozenAt: time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Design.Cases) == 0 || len(result.Confirmation.Cases) == 0 {
		t.Fatalf("expected populated split, got design=%d confirmation=%d", len(result.Design.Cases), len(result.Confirmation.Cases))
	}
	if result.DesignReport.Trajectories != 1 || result.ConfirmReport.Trajectories != 1 {
		t.Fatalf("bootstrap clusters must be source sessions, got design=%d confirmation=%d", result.DesignReport.Trajectories, result.ConfirmReport.Trajectories)
	}
	payload, _ := json.Marshal(result)
	encoded := string(payload)
	for _, forbidden := range []string{"/private/work", "service.go", "2026-05-01", "2026-05-02"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("Codex artifact leaked %q", forbidden)
		}
	}
}

func TestCodexIncompleteToolCallDoesNotCreateOutcomeAnchor(t *testing.T) {
	turn := &codexTurnBuilder{codexTurn: codexTurn{downstreamUsage: make(map[string]struct{})}, pendingCalls: make(map[string]map[string]struct{})}
	turn.pendingCalls["unfinished"] = codexOutcomeAnchors(`{"cmd":"cat /private/work/unfinished.go"}`)
	if len(turn.downstreamUsage) != 0 {
		t.Fatal("an input without a matching output became downstream usage")
	}
}

func TestReadCodexSessionAppliesColdStartBoundaryWithoutLosingSessionIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 8, 27, 23, 0, 0, 0, time.UTC)
	writeCodexRecord(t, file, start, "session_meta", map[string]any{"type": "session_meta", "id": "long-running"})
	for index, base := range []time.Time{start.Add(30 * time.Minute), start.Add(90 * time.Minute)} {
		writeCodexRecord(t, file, base, "event_msg", map[string]any{"type": "user_message", "message": fmt.Sprintf("turn %d", index)})
		writeCodexRecord(t, file, base.Add(time.Second), "event_msg", map[string]any{"type": "task_complete", "last_agent_message": "done"})
	}
	_ = file.Close()
	boundary := time.Date(2026, 8, 28, 0, 0, 0, 0, time.UTC)
	current, _, err := readCodexSession(path, boundary, boundary.Add(2*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if current.id == "" || len(current.turns) != 1 || current.turns[0].user != "turn 1" {
		t.Fatalf("cold-start session = %+v", current)
	}
}

func TestRunCodexAllowsSingleClusterOnlyForExplicitHoldoutPilot(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 8, 28, 1, 0, 0, 0, time.UTC)
	writeCodexRecord(t, file, start, "session_meta", map[string]any{"type": "session_meta", "id": "pilot"})
	for turn := 0; turn < 12; turn++ {
		base := start.Add(time.Duration(turn) * time.Minute)
		writeCodexRecord(t, file, base, "event_msg", map[string]any{"type": "user_message", "message": "inspect /private/pilot.go"})
		callID := fmt.Sprintf("pilot-%d", turn)
		writeCodexRecord(t, file, base.Add(time.Second), "response_item", map[string]any{"type": "function_call", "call_id": callID, "arguments": `{"cmd":"inspect /private/pilot.go"}`})
		writeCodexRecord(t, file, base.Add(2*time.Second), "response_item", map[string]any{"type": "function_call_output", "call_id": callID, "output": "ok"})
		writeCodexRecord(t, file, base.Add(3*time.Second), "event_msg", map[string]any{"type": "task_complete", "last_agent_message": "done"})
	}
	_ = file.Close()
	result, err := RunCodex(context.Background(), CodexConfig{
		SessionDirs: []string{dir}, DataStart: start, DataEnd: start.Add(time.Hour), RuleFrozenAt: start.Add(2 * time.Hour), HoldoutOnly: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Design.Cases) != 0 || len(result.Confirmation.Cases) == 0 || result.ConfirmReport.Trajectories != 1 {
		t.Fatalf("pilot split design=%d holdout=%d trajectories=%d", len(result.Design.Cases), len(result.Confirmation.Cases), result.ConfirmReport.Trajectories)
	}
	if result.ConfirmReport.Comparisons["eventframe"].ClusterInferenceValid {
		t.Fatal("single-cluster pilot incorrectly qualified for cluster inference")
	}
}

func TestReadCodexSessionCurrentEnvelopeBindsOnlyCompletedCalls(t *testing.T) {
	path := filepath.Join(t.TempDir(), "current.jsonl")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	writeCodexRecord(t, file, start.Add(-time.Minute), "session_meta", map[string]any{"type": "session_meta", "id": "current-format"})
	writeCodexRecord(t, file, start, "event_msg", map[string]any{"type": "task_started"})
	writeCodexRecord(t, file, start.Add(time.Second), "response_item", map[string]any{"type": "message", "role": "user", "content": []any{map[string]any{"type": "input_text", "text": "inspect /private/current.go"}}})
	writeCodexRecord(t, file, start.Add(2*time.Second), "response_item", map[string]any{"type": "custom_tool_call", "call_id": "matched", "input": `{"cmd":"inspect /private/current.go"}`})
	writeCodexRecord(t, file, start.Add(3*time.Second), "response_item", map[string]any{"type": "custom_tool_call_output", "call_id": "matched", "output": "ok"})
	writeCodexRecord(t, file, start.Add(4*time.Second), "response_item", map[string]any{"type": "custom_tool_call", "call_id": "unfinished", "input": `{"cmd":"inspect /private/unseen.go"}`})
	writeCodexRecord(t, file, start.Add(5*time.Second), "response_item", map[string]any{"type": "message", "role": "assistant", "phase": "final_answer", "content": []any{map[string]any{"type": "output_text", "text": "inspected current.go"}}})
	writeCodexRecord(t, file, start.Add(6*time.Second), "event_msg", map[string]any{"type": "task_complete", "last_agent_message": ""})
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	current, _, err := readCodexSession(path, start, start.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(current.turns) != 1 || current.turns[0].user != "inspect /private/current.go" || current.turns[0].assistant != "inspected current.go" {
		t.Fatalf("current envelope did not yield one completed turn: %d", len(current.turns))
	}
	if _, ok := current.turns[0].downstreamUsage["path:/private/current.go"]; !ok {
		t.Fatal("matching tool output did not bind completed tool input")
	}
	if _, ok := current.turns[0].downstreamUsage["path:/private/unseen.go"]; ok {
		t.Fatal("unfinished tool input became outcome evidence")
	}
}

func TestReadCodexSessionRejectsAmbiguousOrFailedTurn(t *testing.T) {
	for _, mode := range []string{"multiple-user", "failed-task"} {
		t.Run(mode, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "session.jsonl")
			file, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
			writeCodexRecord(t, file, start, "session_meta", map[string]any{"type": "session_meta", "id": mode})
			writeCodexRecord(t, file, start.Add(time.Second), "event_msg", map[string]any{"type": "user_message", "message": "first request"})
			if mode == "multiple-user" {
				writeCodexRecord(t, file, start.Add(2*time.Second), "event_msg", map[string]any{"type": "user_message", "message": "later request"})
			}
			completion := map[string]any{"type": "task_complete", "last_agent_message": "done"}
			if mode == "failed-task" {
				completion["error"] = "failed"
			}
			writeCodexRecord(t, file, start.Add(3*time.Second), "event_msg", completion)
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			current, _, err := readCodexSession(path, start, start.Add(time.Hour))
			if err != nil || len(current.turns) != 0 {
				t.Fatalf("accepted %s turn: count=%d err=%v", mode, len(current.turns), err)
			}
		})
	}
}

func TestRunCodexCurrentEnvelopeStillExportsNoTranscriptText(t *testing.T) {
	dir := t.TempDir()
	for sessionIndex := 0; sessionIndex < 2; sessionIndex++ {
		start := time.Date(2026, 9, 1+sessionIndex, 0, 0, 0, 0, time.UTC)
		path := filepath.Join(dir, fmt.Sprintf("current-%d.jsonl", sessionIndex))
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		writeCodexRecord(t, file, start, "session_meta", map[string]any{"type": "session_meta", "id": fmt.Sprintf("current-%d", sessionIndex)})
		for turn := 0; turn < 12; turn++ {
			base := start.Add(time.Duration(turn) * time.Minute)
			secret := "/private/work/current/service.go"
			writeCodexRecord(t, file, base, "event_msg", map[string]any{"type": "task_started"})
			writeCodexRecord(t, file, base.Add(time.Second), "response_item", map[string]any{"type": "message", "role": "user", "content": []any{map[string]any{"type": "input_text", "text": "inspect " + secret}}})
			callID := fmt.Sprintf("current-%d", turn)
			writeCodexRecord(t, file, base.Add(2*time.Second), "response_item", map[string]any{"type": "custom_tool_call", "call_id": callID, "input": `{"cmd":"inspect ` + secret + `"}`})
			writeCodexRecord(t, file, base.Add(3*time.Second), "response_item", map[string]any{"type": "custom_tool_call_output", "call_id": callID, "output": "ok"})
			writeCodexRecord(t, file, base.Add(4*time.Second), "event_msg", map[string]any{"type": "task_complete", "last_agent_message": "inspected " + secret})
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	result, err := RunCodex(context.Background(), CodexConfig{
		SessionDirs: []string{dir}, DataStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		DataEnd: time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC), RuleFrozenAt: time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC),
	})
	if err != nil || len(result.Design.Cases) == 0 || len(result.Confirmation.Cases) == 0 {
		t.Fatalf("current-format replay design=%d confirmation=%d err=%v", len(result.Design.Cases), len(result.Confirmation.Cases), err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"/private/work", "service.go", "2026-09-01T00:00:01Z", "2026-09-02T00:00:01Z"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("current-format artifact leaked %q", forbidden)
		}
	}
}

func writeCodexRecord(t *testing.T, file *os.File, timestamp time.Time, recordType string, payload map[string]any) {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{"type": recordType, "timestamp": timestamp.Format(time.RFC3339Nano), "payload": payload})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(append(encoded, '\n')); err != nil {
		t.Fatal(err)
	}
}
