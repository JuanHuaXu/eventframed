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

func TestRunChatGPTDeduplicatesPagesSplitsAndDoesNotExportPrivateData(t *testing.T) {
	root := t.TempDir()
	raw := filepath.Join(root, "raw")
	if err := os.Mkdir(raw, 0o700); err != nil {
		t.Fatal(err)
	}
	manifest := map[string]any{"schema": "test", "split_rule": "frozen", "chats": []map[string]string{
		{"id_hash": "designhash", "split": "design"}, {"id_hash": "confirmhash", "split": "confirmation"},
	}}
	writeTestJSON(t, filepath.Join(root, "manifest.json"), manifest)
	for _, hash := range []string{"designhash", "confirmhash"} {
		turns := make([]map[string]any, 0, 12)
		for index := 0; index < 12; index++ {
			user := fmt.Sprintf("ordinary question %d", index)
			if index == 10 {
				user = "please inspect /private/secret/project.go"
			}
			if index == 11 {
				user = "proceed"
			}
			turns = append(turns, testChatGPTTurn(hash, index, user))
		}
		writeTestJSON(t, filepath.Join(raw, hash+"-000.json"), map[string]any{"page": map[string]any{"order": "newest_first"}, "turns": reverseTurns(turns[5:])})
		writeTestJSON(t, filepath.Join(raw, hash+"-001.json"), map[string]any{"page": map[string]any{"order": "newest_first"}, "turns": reverseTurns(turns[:6])})
	}
	result, err := RunChatGPT(context.Background(), ChatGPTConfig{
		RawDir: raw, ManifestPath: filepath.Join(root, "manifest.json"), RuleFrozenAt: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Design.Cases) != 1 || len(result.Confirmation.Cases) != 1 {
		t.Fatalf("cases design=%d confirmation=%d, want 1 each", len(result.Design.Cases), len(result.Confirmation.Cases))
	}
	if result.Design.Source.MessagesAccepted != 24 {
		t.Fatalf("deduplicated messages=%d, want 24", result.Design.Source.MessagesAccepted)
	}
	payload, _ := json.Marshal(result)
	for _, forbidden := range []string{"/private/secret", "project.go", "design-turn", "confirm-turn", "2026-07-01"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("export leaked %q", forbidden)
		}
	}
}

func TestChatGPTRejectsConflictingPaginationOverlap(t *testing.T) {
	root := t.TempDir()
	first := testChatGPTTurn("chat", 0, "question")
	second := testChatGPTTurn("chat", 0, "different question")
	writeTestJSON(t, filepath.Join(root, "chat-000.json"), map[string]any{"page": map[string]any{"order": "newest_first"}, "turns": []any{first}})
	writeTestJSON(t, filepath.Join(root, "chat-001.json"), map[string]any{"page": map[string]any{"order": "newest_first"}, "turns": []any{second}})
	if _, _, _, err := readChatGPTConversation([]string{filepath.Join(root, "chat-000.json"), filepath.Join(root, "chat-001.json")}, "design"); err == nil {
		t.Fatal("conflicting overlap was accepted")
	}
}

func TestChatGPTExcludesTurnWithImpossibleCompletionTime(t *testing.T) {
	root := t.TempDir()
	valid := testChatGPTTurn("chat", 0, "question")
	invalid := testChatGPTTurn("chat", 1, "later question")
	invalid["completedAt"] = invalid["startedAt"].(float64) - 1
	path := filepath.Join(root, "chat-000.json")
	writeTestJSON(t, path, map[string]any{"page": map[string]any{"order": "newest_first"}, "turns": []any{invalid, valid}})
	current, _, _, err := readChatGPTConversation([]string{path}, "design")
	if err != nil {
		t.Fatal(err)
	}
	if len(current.messages) != 1 {
		t.Fatalf("retained messages=%d, want 1", len(current.messages))
	}
}

func testChatGPTTurn(prefix string, index int, user string) map[string]any {
	started := float64(time.Date(2026, 7, 1, 0, index, 0, 0, time.UTC).Unix())
	return map[string]any{
		"id": fmt.Sprintf("%s-turn-%d", prefix, index), "status": "completed", "startedAt": started, "completedAt": started + 1,
		"items": []map[string]any{
			{"id": "u", "type": "userMessage", "content": []map[string]string{{"type": "text", "text": user}}},
			{"id": "a", "type": "agentMessage", "text": "response about " + user},
		},
	}
}

func reverseTurns(input []map[string]any) []map[string]any {
	output := append([]map[string]any(nil), input...)
	for left, right := 0, len(output)-1; left < right; left, right = left+1, right-1 {
		output[left], output[right] = output[right], output[left]
	}
	return output
}

func writeTestJSON(t *testing.T, path string, value any) {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
}
