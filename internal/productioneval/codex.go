package productioneval

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
)

const (
	CodexSchemaVersion  = "eventframe.codex-replay.v1"
	codexDesignFraction = 0.8
	maxToolInputBytes   = 64 * 1024
)

var codexAnchorStoplist = map[string]struct{}{
	"code:functions.exec": {}, "code:tools.exec_command": {}, "code:tools.apply_patch": {},
	"code:yield_time_ms": {}, "code:max_output_tokens": {}, "code:session_id": {},
	"code:call_id": {}, "code:response_length": {}, "code:workdir": {},
}

type CodexConfig struct {
	SessionDirs           []string
	DataStart             time.Time
	DataEnd               time.Time
	RuleFrozenAt          time.Time
	Embedder              embed.Embedder
	DesignOnly            bool
	HoldoutOnly           bool
	Ablations             bool
	BaselineCalibration   calibration.Logit
	PredictiveCalibration calibration.Logit
	CandidateRanker       retrieval.CandidateRanker
	CandidateRetriever    retrieval.CandidateRetriever
	TextIndexer           retrieval.TextIndexer
	MaxSessions           int
	MaxSegmentsPerSession int
}

type CodexProtocol struct {
	SchemaVersion         string            `json:"schema_version"`
	RuleFrozenAt          time.Time         `json:"rule_frozen_at"`
	DataEnd               time.Time         `json:"data_end"`
	DataStart             time.Time         `json:"data_start,omitempty"`
	SourceMode            string            `json:"source_mode"`
	Split                 string            `json:"split"`
	DesignFraction        float64           `json:"design_fraction"`
	CandidateUnit         string            `json:"candidate_unit"`
	OutcomeUnit           string            `json:"outcome_unit"`
	Label                 string            `json:"label"`
	AnchorClasses         []string          `json:"anchor_classes"`
	MaxTextBytes          int               `json:"max_text_bytes_per_turn_field"`
	MaxToolInputBytes     int               `json:"max_tool_input_bytes"`
	MaxSegmentTurns       int               `json:"max_segment_turns"`
	SessionCap            int               `json:"session_cap,omitempty"`
	SegmentCapPerSession  int               `json:"segment_cap_per_session,omitempty"`
	MinimumPriorTurns     int               `json:"minimum_prior_turns"`
	EmbeddingModel        string            `json:"embedding_model"`
	NominationContract    string            `json:"nomination_contract"`
	RetrievalContract     string            `json:"retrieval_contract"`
	AssemblyOwner         string            `json:"assembly_owner"`
	Variants              []string          `json:"variants"`
	RawDataExported       bool              `json:"raw_data_exported"`
	ExportTimeEncoding    string            `json:"export_time_encoding"`
	BootstrapSamples      int               `json:"bootstrap_samples"`
	BootstrapSeed         int64             `json:"bootstrap_seed"`
	EvaluationScope       string            `json:"evaluation_scope"`
	BaselineCalibration   calibration.Logit `json:"baseline_calibration"`
	PredictiveCalibration calibration.Logit `json:"predictive_calibration"`
}

type CodexResult struct {
	Protocol             CodexProtocol     `json:"protocol"`
	Design               Artifact          `json:"design"`
	Confirmation         Artifact          `json:"confirmation"`
	DesignReport         evaluation.Report `json:"design_report"`
	ConfirmReport        evaluation.Report `json:"confirmation_report"`
	DesignControlReport  evaluation.Report `json:"design_control_report"`
	ConfirmControlReport evaluation.Report `json:"confirmation_control_report"`
}

type codexSession struct {
	id        string
	startedAt time.Time
	turns     []codexTurn
}

type codexTurn struct {
	user            string
	assistant       string
	startedAt       time.Time
	completedAt     time.Time
	downstreamUsage map[string]struct{}
}

type codexTurnBuilder struct {
	codexTurn
	pendingCalls map[string]map[string]struct{}
	lastAgent    string
}

type codexEnvelope struct {
	Type      string          `json:"type"`
	Timestamp json.RawMessage `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

func FrozenCodexProtocol(config CodexConfig) CodexProtocol {
	embeddingModel := fmt.Sprintf("feature-hash-v1:d%d", embeddingDimension)
	if config.Embedder != nil {
		embeddingModel = config.Embedder.ModelKey()
	}
	retrievalContract := (retrieval.PassthroughRanker{}).ContractName()
	if config.CandidateRanker != nil {
		retrievalContract = config.CandidateRanker.ContractName()
	}
	nominationContract := "embedded-vector-search"
	if config.CandidateRetriever != nil {
		nominationContract = config.CandidateRetriever.RetrievalContractName()
	}
	scope := "design and frozen confirmation"
	splitDescription := "chronological session-level 80/20 design/confirmation split after eligibility filtering and before outcome scoring"
	if config.DesignOnly {
		scope = "design only; prior confirmation remains sealed"
	}
	if config.HoldoutOnly {
		scope = "cold-start temporal holdout only; no design cases scored"
		splitDescription = "all accepted sessions assigned to the holdout block"
	}
	variants := []string{"baseline", "update_all", "eventframe", "shuffled_feedback"}
	if config.Ablations {
		variants = []string{"baseline", "update_all", "eventframe", "contextual", "hierarchy", "residual_shadow", "full_upgrade", "shuffled_full", "shuffled_feedback"}
	}
	return CodexProtocol{
		SchemaVersion: CodexSchemaVersion, RuleFrozenAt: config.RuleFrozenAt.UTC(), DataStart: config.DataStart.UTC(), DataEnd: config.DataEnd.UTC(),
		SourceMode: "read-only local Codex JSONL session records", Split: splitDescription,
		DesignFraction: codexDesignFraction, CandidateUnit: "one completed prior Codex user/assistant turn",
		OutcomeUnit:   "anchors in a current-turn tool input whose call produced a matching output, plus anchors in a successful patch_apply_end record",
		Label:         "a prior completed turn is relevant iff its normalized explicit anchors intersect the current turn's completed downstream tool-use anchors",
		AnchorClasses: []string{"url", "absolute_path", "code_like_identifier"}, MaxTextBytes: maxTextBytes, MaxToolInputBytes: maxToolInputBytes,
		MaxSegmentTurns: maxSegmentMessages, SessionCap: config.MaxSessions, SegmentCapPerSession: config.MaxSegmentsPerSession,
		MinimumPriorTurns: minimumPriorEvents, EmbeddingModel: embeddingModel,
		NominationContract: nominationContract, RetrievalContract: retrievalContract,
		AssemblyOwner: "original LibraVDB OpenClaw context engine in production; fixed pack_k fallback in replay",
		Variants:      variants, RawDataExported: false,
		ExportTimeEncoding: "trajectory-relative ordinal seconds from 2000-01-01T00:00:00Z; production timestamps are not exported",
		BootstrapSamples:   bootstrapSamples, BootstrapSeed: bootstrapSeed, EvaluationScope: scope,
		BaselineCalibration: config.BaselineCalibration, PredictiveCalibration: config.PredictiveCalibration,
	}
}

func RunCodex(ctx context.Context, config CodexConfig) (CodexResult, error) {
	if len(config.SessionDirs) == 0 || config.DataEnd.IsZero() || config.RuleFrozenAt.IsZero() {
		return CodexResult{}, errors.New("session directories, data end, and rule freeze time are required")
	}
	if !config.DataStart.IsZero() && !config.DataStart.Before(config.DataEnd) {
		return CodexResult{}, errors.New("data start must precede data end")
	}
	if config.DesignOnly && config.HoldoutOnly {
		return CodexResult{}, errors.New("design-only and holdout-only are mutually exclusive")
	}
	if config.MaxSessions == 1 || config.MaxSessions < 0 {
		return CodexResult{}, errors.New("max sessions must be zero or at least two")
	}
	if config.MaxSegmentsPerSession < 0 {
		return CodexResult{}, errors.New("max segments per session cannot be negative")
	}
	if !config.BaselineCalibration.Valid() {
		config.BaselineCalibration = calibration.Identity()
	}
	if !config.PredictiveCalibration.Valid() {
		config.PredictiveCalibration = config.BaselineCalibration
	}
	sessions, manifest, err := loadCodexSessions(config)
	if err != nil {
		return CodexResult{}, err
	}
	eligible := sessions[:0]
	for _, current := range sessions {
		if len(current.turns) > minimumPriorEvents {
			eligible = append(eligible, current)
		}
	}
	sessions = eligible
	minimumSessions := 2
	if config.HoldoutOnly {
		minimumSessions = 1
	}
	if len(sessions) < minimumSessions {
		return CodexResult{}, fmt.Errorf("fewer than %d Codex sessions have enough completed turns", minimumSessions)
	}
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].startedAt.Equal(sessions[j].startedAt) {
			return sessions[i].id < sessions[j].id
		}
		return sessions[i].startedAt.Before(sessions[j].startedAt)
	})
	if config.MaxSessions > 0 && len(sessions) > config.MaxSessions {
		sessions = sessions[:config.MaxSessions]
	}
	split := 0
	if !config.HoldoutOnly {
		split = int(float64(len(sessions)) * codexDesignFraction)
		if split < 1 {
			split = 1
		}
		if split >= len(sessions) {
			split = len(sessions) - 1
		}
	}
	manifest.SessionsAccepted = len(sessions)
	manifest.SessionsDesign = split
	manifest.SessionsConfirmation = len(sessions) - split
	result := CodexResult{Protocol: FrozenCodexProtocol(config)}
	result.Design = Artifact{SchemaVersion: CodexSchemaVersion, Block: "design", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	result.Confirmation = Artifact{SchemaVersion: CodexSchemaVersion, Block: "confirmation", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	for index, current := range sessions {
		block := "design"
		if config.HoldoutOnly || index >= split {
			block = "confirmation"
		}
		if config.DesignOnly && block == "confirmation" {
			continue
		}
		replay := codexReplaySession(current, block, config.Embedder, config.BaselineCalibration, config.PredictiveCalibration, config.Ablations, config.CandidateRanker, config.CandidateRetriever, config.TextIndexer)
		replay.retrievalNamespace = config.RuleFrozenAt.UTC().Format(time.RFC3339Nano)
		replay.maxSegments = config.MaxSegmentsPerSession
		cases, replayErr := replaySession(ctx, replay)
		if replayErr != nil {
			return CodexResult{}, fmt.Errorf("replay Codex session %s: %w", shortHash(current.id), replayErr)
		}
		if block == "design" {
			result.Design.Cases = append(result.Design.Cases, cases...)
		} else {
			result.Confirmation.Cases = append(result.Confirmation.Cases, cases...)
		}
	}
	sortCases(result.Design.Cases)
	sortCases(result.Confirmation.Cases)
	result.DesignReport, err = evaluateArtifact(result.Design)
	if err != nil {
		return CodexResult{}, fmt.Errorf("evaluate Codex design block: %w", err)
	}
	result.ConfirmReport, err = evaluateArtifact(result.Confirmation)
	if err != nil {
		return CodexResult{}, fmt.Errorf("evaluate Codex confirmation block: %w", err)
	}
	result.DesignControlReport, err = evaluateArtifactWithBaseline(result.Design, "shuffled_feedback")
	if err != nil {
		return CodexResult{}, fmt.Errorf("evaluate Codex design negative control: %w", err)
	}
	result.ConfirmControlReport, err = evaluateArtifactWithBaseline(result.Confirmation, "shuffled_feedback")
	if err != nil {
		return CodexResult{}, fmt.Errorf("evaluate Codex confirmation negative control: %w", err)
	}
	return result, nil
}

func loadCodexSessions(config CodexConfig) ([]codexSession, SourceManifest, error) {
	var paths []string
	for _, root := range config.SessionDirs {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".jsonl" {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, SourceManifest{}, fmt.Errorf("walk Codex sessions: %w", err)
		}
	}
	sort.Strings(paths)
	manifest := SourceManifest{FilesScanned: len(paths)}
	byID := make(map[string]codexSession)
	var fileDigests []string
	for _, path := range paths {
		current, fileDigest, err := readCodexSession(path, config.DataStart, config.DataEnd)
		if err != nil {
			return nil, SourceManifest{}, err
		}
		fileDigests = append(fileDigests, fileDigest)
		if current.id == "" || len(current.turns) == 0 || !current.startedAt.Before(config.DataEnd) {
			continue
		}
		manifest.FilesAccepted++
		manifest.MessagesAccepted += len(current.turns)
		if existing, ok := byID[current.id]; !ok || len(current.turns) > len(existing.turns) {
			byID[current.id] = current
		}
	}
	sort.Strings(fileDigests)
	aggregate := sha256.Sum256([]byte(strings.Join(fileDigests, "\n")))
	manifest.AggregateSHA256 = hex.EncodeToString(aggregate[:])
	sessions := make([]codexSession, 0, len(byID))
	for _, current := range byID {
		sessions = append(sessions, current)
	}
	return sessions, manifest, nil
}

func readCodexSession(path string, dataStart, dataEnd time.Time) (codexSession, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return codexSession{}, "", fmt.Errorf("open Codex session: %w", err)
	}
	defer file.Close()
	digest := sha256.New()
	scanner := bufio.NewScanner(io.TeeReader(file, digest))
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	var current codexSession
	var turn *codexTurnBuilder
	line := 0
	for scanner.Scan() {
		line++
		var envelope codexEnvelope
		if err := json.Unmarshal(scanner.Bytes(), &envelope); err != nil {
			return codexSession{}, "", fmt.Errorf("parse Codex session line %d: %w", line, err)
		}
		timestamp, err := parseTimestamp(envelope.Timestamp)
		if err != nil || !timestamp.Before(dataEnd) {
			continue
		}
		var payload map[string]json.RawMessage
		if json.Unmarshal(envelope.Payload, &payload) != nil {
			continue
		}
		payloadType := rawString(payload["type"])
		if envelope.Type == "session_meta" {
			if current.id == "" {
				rawID := rawString(payload["id"])
				if rawID == "" {
					rawID = rawString(payload["session_id"])
				}
				current.id = digestID("codex-session", rawID)
				current.startedAt = timestamp
			}
			continue
		}
		if !dataStart.IsZero() && timestamp.Before(dataStart) {
			continue
		}
		switch {
		case envelope.Type == "event_msg" && payloadType == "user_message":
			query := boundedText(rawText(payload["message"]))
			if query != "" {
				turn = &codexTurnBuilder{codexTurn: codexTurn{user: query, startedAt: timestamp, downstreamUsage: make(map[string]struct{})}, pendingCalls: make(map[string]map[string]struct{})}
			}
		case envelope.Type == "event_msg" && payloadType == "agent_message" && turn != nil:
			if text := boundedText(rawText(payload["message"])); text != "" {
				turn.lastAgent = text
			}
		case envelope.Type == "response_item" && (payloadType == "function_call" || payloadType == "custom_tool_call") && turn != nil:
			callID := rawString(payload["call_id"])
			if callID != "" {
				input := payload["arguments"]
				if len(input) == 0 {
					input = payload["input"]
				}
				turn.pendingCalls[callID] = codexOutcomeAnchors(boundedRaw(input, maxToolInputBytes))
			}
		case envelope.Type == "response_item" && (payloadType == "function_call_output" || payloadType == "custom_tool_call_output") && turn != nil:
			if anchors, ok := turn.pendingCalls[rawString(payload["call_id"])]; ok {
				mergeAnchors(turn.downstreamUsage, anchors)
			}
		case envelope.Type == "event_msg" && payloadType == "mcp_tool_call_end" && turn != nil:
			mergeAnchors(turn.downstreamUsage, codexOutcomeAnchors(boundedRaw(payload["invocation"], maxToolInputBytes)))
		case envelope.Type == "event_msg" && payloadType == "patch_apply_end" && turn != nil && rawBool(payload["success"]):
			mergeAnchors(turn.downstreamUsage, codexOutcomeAnchors(boundedRaw(payload["changes"], maxToolInputBytes)))
		case envelope.Type == "event_msg" && payloadType == "task_complete" && turn != nil:
			assistant := boundedText(rawText(payload["last_agent_message"]))
			if assistant == "" {
				assistant = turn.lastAgent
			}
			if turn.user != "" && assistant != "" && timestamp.After(turn.startedAt) {
				turn.assistant = assistant
				turn.completedAt = timestamp
				current.turns = append(current.turns, turn.codexTurn)
			}
			turn = nil
		}
	}
	if err := scanner.Err(); err != nil {
		return codexSession{}, "", fmt.Errorf("scan Codex session: %w", err)
	}
	fileDigest := hex.EncodeToString(digest.Sum(nil))
	if current.id == "" {
		current.id = digestID("codex-session-file", fileDigest)
	}
	return current, fileDigest, nil
}

func codexReplaySession(current codexSession, block string, activeEmbedder embed.Embedder, baselineCalibration, predictiveCalibration calibration.Logit, ablations bool, candidateRanker retrieval.CandidateRanker, candidateRetriever retrieval.CandidateRetriever, textIndexer retrieval.TextIndexer) session {
	messages := make([]message, 0, len(current.turns))
	for index, turn := range current.turns {
		content := "User: " + turn.user + "\n\nAssistant: " + turn.assistant
		messages = append(messages, message{
			id: digestID("codex-turn", current.id, fmt.Sprint(index)), role: "user", text: content, query: turn.user,
			queryAt: turn.startedAt, available: turn.completedAt, anchors: ExtractAnchors(content), labelAnchors: turn.downstreamUsage,
		})
	}
	return session{digest: current.id, block: block, negativeControl: true, embedder: activeEmbedder, baselineCalibration: baselineCalibration, predictiveCalibration: predictiveCalibration, ablations: ablations, candidateRanker: candidateRanker, candidateRetriever: candidateRetriever, textIndexer: textIndexer, messages: messages}
}

func rawString(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	return ""
}

func rawBool(raw json.RawMessage) bool {
	var value bool
	return json.Unmarshal(raw, &value) == nil && value
}

func rawText(raw json.RawMessage) string {
	if direct := rawString(raw); direct != "" {
		return direct
	}
	var parts []map[string]json.RawMessage
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var texts []string
	for _, part := range parts {
		if text := rawString(part["text"]); text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "\n")
}

func boundedRaw(raw json.RawMessage, limit int) string {
	if len(raw) > limit {
		raw = raw[:limit]
	}
	return string(raw)
}

func codexOutcomeAnchors(text string) map[string]struct{} {
	anchors := ExtractAnchors(text)
	for value := range codexAnchorStoplist {
		delete(anchors, value)
	}
	return anchors
}

func mergeAnchors(destination, source map[string]struct{}) {
	for value := range source {
		destination[value] = struct{}{}
	}
}
