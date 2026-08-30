package productioneval

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/evaluation"
)

const ChatGPTSchemaVersion = "eventframe.chatgpt-replay.v1"

var narrowFollowupPattern = regexp.MustCompile(`(?i)^\s*(?:yes|no|ok(?:ay)?|sure|proceed|continue|go ahead|do that|do it|implement it|fix it|push it|run it|rerun)(?:[\s,.!?:;]|$)`)

type ChatGPTConfig struct {
	RawDir                string
	ManifestPath          string
	RuleFrozenAt          time.Time
	Embedder              embed.Embedder
	Ablations             bool
	BaselineCalibration   calibration.Logit
	PredictiveCalibration calibration.Logit
}

type ChatGPTProtocol struct {
	SchemaVersion         string            `json:"schema_version"`
	RuleFrozenAt          time.Time         `json:"rule_frozen_at"`
	SourceMode            string            `json:"source_mode"`
	Split                 string            `json:"split"`
	CandidateUnit         string            `json:"candidate_unit"`
	Label                 string            `json:"label"`
	MinimumPriorTurns     int               `json:"minimum_prior_turns"`
	RecallK               int               `json:"recall_k"`
	PackK                 int               `json:"pack_k"`
	RawDataExported       bool              `json:"raw_data_exported"`
	ExportTimeEncoding    string            `json:"export_time_encoding"`
	BootstrapSamples      int               `json:"bootstrap_samples"`
	BootstrapSeed         int64             `json:"bootstrap_seed"`
	BaselineCalibration   calibration.Logit `json:"baseline_calibration"`
	PredictiveCalibration calibration.Logit `json:"predictive_calibration"`
}

type ChatGPTResult struct {
	Protocol             ChatGPTProtocol   `json:"protocol"`
	Design               Artifact          `json:"design"`
	Confirmation         Artifact          `json:"confirmation"`
	DesignReport         evaluation.Report `json:"design_report"`
	ConfirmReport        evaluation.Report `json:"confirmation_report"`
	DesignControlReport  evaluation.Report `json:"design_control_report"`
	ConfirmControlReport evaluation.Report `json:"confirmation_control_report"`
}

type chatGPTManifest struct {
	Schema    string `json:"schema"`
	SplitRule string `json:"split_rule"`
	Chats     []struct {
		IDHash string `json:"id_hash"`
		Split  string `json:"split"`
	} `json:"chats"`
}

type chatGPTPage struct {
	Page struct {
		Order string `json:"order"`
	} `json:"page"`
	Turns []chatGPTTurn `json:"turns"`
}

type chatGPTTurn struct {
	ID          string          `json:"id"`
	Status      string          `json:"status"`
	StartedAt   float64         `json:"startedAt"`
	CompletedAt float64         `json:"completedAt"`
	Items       json.RawMessage `json:"items"`
}

type chatGPTItem struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	Content json.RawMessage `json:"content"`
}

func FrozenChatGPTProtocol(config ChatGPTConfig) ChatGPTProtocol {
	baseline := config.BaselineCalibration
	if !baseline.Valid() {
		baseline = calibration.Identity()
	}
	predictive := config.PredictiveCalibration
	if !predictive.Valid() {
		predictive = baseline
	}
	return ChatGPTProtocol{
		SchemaVersion: ChatGPTSchemaVersion, RuleFrozenAt: config.RuleFrozenAt.UTC(),
		SourceMode: "read-only private ChatGPT connector pages", Split: "frozen conversation-level split from the private extraction manifest",
		CandidateUnit:     "one completed prior user/assistant turn",
		Label:             "explicit normalized anchor overlap, plus a frozen narrow immediate-followup rule; no model-generated relevance labels",
		MinimumPriorTurns: minimumPriorEvents, RecallK: maxSegmentMessages, PackK: 10,
		RawDataExported: false, ExportTimeEncoding: "trajectory-relative ordinal seconds; source timestamps are not exported",
		BootstrapSamples: bootstrapSamples, BootstrapSeed: bootstrapSeed,
		BaselineCalibration: baseline, PredictiveCalibration: predictive,
	}
}

func RunChatGPT(ctx context.Context, config ChatGPTConfig) (ChatGPTResult, error) {
	if config.RawDir == "" || config.ManifestPath == "" || config.RuleFrozenAt.IsZero() {
		return ChatGPTResult{}, errors.New("raw directory, manifest, and rule freeze time are required")
	}
	sessions, manifest, err := loadChatGPTSessions(config)
	if err != nil {
		return ChatGPTResult{}, err
	}
	if manifest.SessionsDesign == 0 || manifest.SessionsConfirmation == 0 {
		return ChatGPTResult{}, errors.New("ChatGPT manifest must contain eligible design and confirmation conversations")
	}
	protocol := FrozenChatGPTProtocol(config)
	result := ChatGPTResult{Protocol: protocol}
	result.Design = Artifact{SchemaVersion: ChatGPTSchemaVersion, Block: "design", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	result.Confirmation = Artifact{SchemaVersion: ChatGPTSchemaVersion, Block: "confirmation", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	for _, current := range sessions {
		current.embedder = config.Embedder
		current.baselineCalibration = protocol.BaselineCalibration
		current.predictiveCalibration = protocol.PredictiveCalibration
		current.ablations = config.Ablations
		current.negativeControl = true
		cases, replayErr := replaySession(ctx, current)
		if replayErr != nil {
			return ChatGPTResult{}, fmt.Errorf("replay ChatGPT conversation %s: %w", shortHash(current.digest), replayErr)
		}
		if current.block == "design" {
			result.Design.Cases = append(result.Design.Cases, cases...)
		} else {
			result.Confirmation.Cases = append(result.Confirmation.Cases, cases...)
		}
	}
	sortCases(result.Design.Cases)
	sortCases(result.Confirmation.Cases)
	result.DesignReport, err = evaluateArtifact(result.Design)
	if err != nil {
		return ChatGPTResult{}, err
	}
	result.ConfirmReport, err = evaluateArtifact(result.Confirmation)
	if err != nil {
		return ChatGPTResult{}, err
	}
	result.DesignControlReport, err = evaluateArtifactWithBaseline(result.Design, "shuffled_feedback")
	if err != nil {
		return ChatGPTResult{}, err
	}
	result.ConfirmControlReport, err = evaluateArtifactWithBaseline(result.Confirmation, "shuffled_feedback")
	return result, err
}

func loadChatGPTSessions(config ChatGPTConfig) ([]session, SourceManifest, error) {
	payload, err := os.ReadFile(config.ManifestPath)
	if err != nil {
		return nil, SourceManifest{}, fmt.Errorf("read ChatGPT manifest: %w", err)
	}
	var manifest chatGPTManifest
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return nil, SourceManifest{}, fmt.Errorf("parse ChatGPT manifest: %w", err)
	}
	if len(manifest.Chats) == 0 {
		return nil, SourceManifest{}, errors.New("ChatGPT manifest has no conversations")
	}
	var sessions []session
	var source SourceManifest
	var digests []string
	seenHashes := make(map[string]struct{}, len(manifest.Chats))
	for _, entry := range manifest.Chats {
		if entry.IDHash == "" || (entry.Split != "design" && entry.Split != "confirmation") {
			return nil, SourceManifest{}, errors.New("ChatGPT manifest contains an invalid hash or split")
		}
		if _, duplicate := seenHashes[entry.IDHash]; duplicate {
			return nil, SourceManifest{}, fmt.Errorf("duplicate ChatGPT manifest hash %q", entry.IDHash)
		}
		seenHashes[entry.IDHash] = struct{}{}
		paths, globErr := filepath.Glob(filepath.Join(config.RawDir, entry.IDHash+"-*.json"))
		if globErr != nil || len(paths) == 0 {
			return nil, SourceManifest{}, fmt.Errorf("no raw pages for ChatGPT conversation %s", entry.IDHash)
		}
		sort.Strings(paths)
		current, digest, pageCount, readErr := readChatGPTConversation(paths, entry.Split)
		if readErr != nil {
			return nil, SourceManifest{}, fmt.Errorf("read ChatGPT conversation %s: %w", entry.IDHash, readErr)
		}
		source.FilesScanned += pageCount
		if len(current.messages) <= minimumPriorEvents {
			continue
		}
		source.FilesAccepted += pageCount
		source.MessagesAccepted += len(current.messages)
		source.SessionsAccepted++
		if entry.Split == "design" {
			source.SessionsDesign++
		} else {
			source.SessionsConfirmation++
		}
		digests = append(digests, digest)
		sessions = append(sessions, current)
	}
	sort.Strings(digests)
	aggregate := sha256.Sum256([]byte(strings.Join(digests, "\n")))
	source.AggregateSHA256 = hex.EncodeToString(aggregate[:])
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].digest < sessions[j].digest })
	return sessions, source, nil
}

func readChatGPTConversation(paths []string, block string) (session, string, int, error) {
	turns := make(map[string]chatGPTTurn)
	digest := sha256.New()
	for _, path := range paths {
		payload, err := os.ReadFile(path)
		if err != nil {
			return session{}, "", 0, err
		}
		digest.Write(payload)
		var page chatGPTPage
		if err := json.Unmarshal(payload, &page); err != nil {
			return session{}, "", 0, err
		}
		if page.Page.Order != "newest_first" {
			return session{}, "", 0, fmt.Errorf("page order %q is not newest_first", page.Page.Order)
		}
		for _, turn := range page.Turns {
			if turn.ID == "" || turn.StartedAt <= 0 {
				return session{}, "", 0, errors.New("turn lacks valid identity or start timestamp")
			}
			if turn.CompletedAt <= 0 || turn.CompletedAt < turn.StartedAt {
				continue
			}
			if existing, ok := turns[turn.ID]; ok {
				if existing.StartedAt != turn.StartedAt || existing.CompletedAt != turn.CompletedAt || string(existing.Items) != string(turn.Items) {
					return session{}, "", 0, fmt.Errorf("overlapping turn %q differs across pages", turn.ID)
				}
				continue
			}
			turns[turn.ID] = turn
		}
	}
	ordered := make([]chatGPTTurn, 0, len(turns))
	for _, turn := range turns {
		ordered = append(ordered, turn)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].StartedAt == ordered[j].StartedAt {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].StartedAt < ordered[j].StartedAt
	})
	conversationDigest := hex.EncodeToString(digest.Sum(nil))
	messages := make([]message, 0, len(ordered))
	for index, turn := range ordered {
		user, assistant, err := chatGPTTurnText(turn.Items)
		if err != nil {
			return session{}, "", 0, fmt.Errorf("turn %q: %w", turn.ID, err)
		}
		if strings.TrimSpace(user) == "" || strings.TrimSpace(assistant) == "" {
			continue
		}
		started := unixFloatTime(turn.StartedAt)
		completed := unixFloatTime(turn.CompletedAt)
		anchors := ExtractAnchors(user + "\n" + assistant)
		labels := ExtractAnchors(user)
		if narrowFollowupPattern.MatchString(user) && len(messages) > 0 {
			key := "followup:" + shortHash(messages[len(messages)-1].id)
			messages[len(messages)-1].anchors[key] = struct{}{}
			labels[key] = struct{}{}
		}
		messages = append(messages, message{
			id: digestID("chatgpt-turn", conversationDigest, turn.ID, fmt.Sprint(index)), role: "user",
			text: boundedText("User: " + user + "\nAssistant: " + assistant), query: boundedText(user),
			queryAt: started, available: completed, anchors: anchors, labelAnchors: labels,
		})
	}
	return session{digest: conversationDigest, block: block, messages: messages}, conversationDigest, len(paths), nil
}

func chatGPTTurnText(raw json.RawMessage) (string, string, error) {
	var items []chatGPTItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return "", "", err
	}
	var userParts, assistantParts []string
	for _, item := range items {
		switch item.Type {
		case "userMessage":
			text := extractText(item.Content)
			if strings.TrimSpace(text) != "" {
				userParts = append(userParts, text)
			}
		case "agentMessage":
			if strings.TrimSpace(item.Text) != "" {
				assistantParts = append(assistantParts, item.Text)
			}
		}
	}
	return boundedText(strings.Join(userParts, "\n")), boundedText(strings.Join(assistantParts, "\n")), nil
}

func unixFloatTime(value float64) time.Time {
	seconds := int64(value)
	nanos := int64((value - float64(seconds)) * float64(time.Second))
	return time.Unix(seconds, nanos).UTC()
}
