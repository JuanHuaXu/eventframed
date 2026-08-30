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
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/bayes"
	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/ranking"
	"github.com/JuanHuaXu/eventframed/internal/residual"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

const (
	SchemaVersion            = "eventframe.production-replay.v1"
	defaultTenant            = "production-replay"
	maxTextBytes             = 16 * 1024
	maxSegmentMessages       = 100
	minimumPriorEvents       = 10
	embeddingDimension       = 384
	bootstrapSamples         = 2_000
	bootstrapSeed      int64 = 26_082_801
)

var (
	uuidFilePattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\.jsonl$`)
	urlPattern      = regexp.MustCompile(`(?i)https?://[^\s<>"']+`)
	pathPattern     = regexp.MustCompile(`(?:^|[\s("'])((?:/[A-Za-z0-9._~!$&()*+,;=:@%+-]+){2,})`)
	codePattern     = regexp.MustCompile(`\b[A-Za-z][A-Za-z0-9]*(?:[._:/-][A-Za-z0-9_:-]+)+\b`)
)

type Config struct {
	SessionDir        string
	ConfirmationStart time.Time
	DataEnd           time.Time
	RuleFrozenAt      time.Time
}

type Protocol struct {
	SchemaVersion        string    `json:"schema_version"`
	RuleFrozenAt         time.Time `json:"rule_frozen_at"`
	ConfirmationStart    time.Time `json:"confirmation_start"`
	DataEnd              time.Time `json:"data_end"`
	Agent                string    `json:"agent"`
	SourceMode           string    `json:"source_mode"`
	Label                string    `json:"label"`
	AnchorClasses        []string  `json:"anchor_classes"`
	MaxTextBytes         int       `json:"max_text_bytes"`
	MaxSegmentMessages   int       `json:"max_segment_messages"`
	MinimumPriorEvents   int       `json:"minimum_prior_events"`
	EmbeddingModel       string    `json:"embedding_model"`
	RecallK              int       `json:"recall_k"`
	PackK                int       `json:"pack_k"`
	Variants             []string  `json:"variants"`
	BootstrapSamples     int       `json:"bootstrap_samples"`
	BootstrapSeed        int64     `json:"bootstrap_seed"`
	RawTextExported      bool      `json:"raw_text_exported"`
	ExportTimeEncoding   string    `json:"export_time_encoding"`
	TrajectoryDefinition string    `json:"trajectory_definition"`
}

type SourceManifest struct {
	FilesScanned         int    `json:"files_scanned"`
	FilesAccepted        int    `json:"files_accepted"`
	FilesDesign          int    `json:"files_design"`
	FilesConfirmation    int    `json:"files_confirmation"`
	FilesCrossingCutoff  int    `json:"files_crossing_cutoff"`
	MessagesAccepted     int    `json:"messages_accepted"`
	SessionsAccepted     int    `json:"sessions_accepted,omitempty"`
	SessionsDesign       int    `json:"sessions_design,omitempty"`
	SessionsConfirmation int    `json:"sessions_confirmation,omitempty"`
	AggregateSHA256      string `json:"aggregate_sha256"`
}

type Artifact struct {
	SchemaVersion string            `json:"schema_version"`
	Block         string            `json:"block"`
	RuleFrozenAt  time.Time         `json:"rule_frozen_at"`
	Source        SourceManifest    `json:"source"`
	Cases         []evaluation.Case `json:"cases"`
}

type Result struct {
	Protocol      Protocol          `json:"protocol"`
	Design        Artifact          `json:"design"`
	Confirmation  Artifact          `json:"confirmation"`
	DesignReport  evaluation.Report `json:"design_report"`
	ConfirmReport evaluation.Report `json:"confirmation_report"`
}

type message struct {
	id           string
	role         string
	text         string
	query        string
	queryAt      time.Time
	available    time.Time
	anchors      map[string]struct{}
	labelAnchors map[string]struct{}
}

type session struct {
	digest                string
	block                 string
	negativeControl       bool
	embedder              embed.Embedder
	baselineCalibration   calibration.Logit
	predictiveCalibration calibration.Logit
	ablations             bool
	candidateRanker       retrieval.CandidateRanker
	candidateRetriever    retrieval.CandidateRetriever
	textIndexer           retrieval.TextIndexer
	retrievalNamespace    string
	maxSegments           int
	messages              []message
}

type rawRecord struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Timestamp json.RawMessage `json:"timestamp"`
	Message   struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type variantRunner struct {
	name     string
	runtime  *service.Service
	memory   *memorystore.Store
	embedder embed.Embedder
	feedback string
}

type cachedCandidateRetriever struct {
	upstream retrieval.CandidateRetriever
	results  map[string][]retrieval.Candidate
}

type cachedCandidateRanker struct {
	upstream retrieval.CandidateRanker
	results  map[string][]retrieval.Candidate
}

func newCachedCandidateRanker(upstream retrieval.CandidateRanker) *cachedCandidateRanker {
	return &cachedCandidateRanker{upstream: upstream, results: make(map[string][]retrieval.Candidate)}
}

func (r *cachedCandidateRanker) ContractName() string { return r.upstream.ContractName() }

func (r *cachedCandidateRanker) RankCandidates(ctx context.Context, request retrieval.RankRequest) ([]retrieval.Candidate, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	key := string(encoded)
	if cached, ok := r.results[key]; ok {
		return append([]retrieval.Candidate(nil), cached...), nil
	}
	results, err := r.upstream.RankCandidates(ctx, request)
	if err != nil {
		return nil, err
	}
	r.results[key] = append([]retrieval.Candidate(nil), results...)
	return append([]retrieval.Candidate(nil), results...), nil
}

func newCachedCandidateRetriever(upstream retrieval.CandidateRetriever) *cachedCandidateRetriever {
	return &cachedCandidateRetriever{upstream: upstream, results: make(map[string][]retrieval.Candidate)}
}

func (r *cachedCandidateRetriever) RetrievalContractName() string {
	return r.upstream.RetrievalContractName()
}

func (r *cachedCandidateRetriever) SearchTextCollections(ctx context.Context, request retrieval.SearchRequest) ([]retrieval.Candidate, error) {
	encoded, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	key := string(encoded)
	if cached, ok := r.results[key]; ok {
		return append([]retrieval.Candidate(nil), cached...), nil
	}
	results, err := r.upstream.SearchTextCollections(ctx, request)
	if err != nil {
		return nil, err
	}
	r.results[key] = append([]retrieval.Candidate(nil), results...)
	return append([]retrieval.Candidate(nil), results...), nil
}

func FrozenProtocol(config Config) Protocol {
	return Protocol{
		SchemaVersion: SchemaVersion, RuleFrozenAt: config.RuleFrozenAt.UTC(),
		ConfirmationStart: config.ConfirmationStart.UTC(), DataEnd: config.DataEnd.UTC(),
		Agent: "main", SourceMode: "read-only OpenClaw UUID session JSONL",
		Label:         "a prior message is relevant iff it shares at least one normalized explicit URL, absolute path, or code-like identifier with the current user message",
		AnchorClasses: []string{"url", "absolute_path", "code_like_identifier"},
		MaxTextBytes:  maxTextBytes, MaxSegmentMessages: maxSegmentMessages, MinimumPriorEvents: minimumPriorEvents,
		EmbeddingModel: embed.BindRepresentation(fmt.Sprintf("feature-hash-v1:d%d", embeddingDimension)), RecallK: maxSegmentMessages, PackK: maxSegmentMessages,
		Variants: []string{"baseline", "update_all", "eventframe"}, BootstrapSamples: bootstrapSamples, BootstrapSeed: bootstrapSeed,
		RawTextExported: false, ExportTimeEncoding: "trajectory-relative ordinal seconds from 2000-01-01T00:00:00Z; production timestamps are not exported",
		TrajectoryDefinition: "one isolated source-session segment of at most 100 chronological user/assistant text messages",
	}
}

func Run(ctx context.Context, config Config) (Result, error) {
	if config.SessionDir == "" || config.ConfirmationStart.IsZero() || config.DataEnd.IsZero() || config.RuleFrozenAt.IsZero() {
		return Result{}, errors.New("session directory and all protocol times are required")
	}
	if !config.ConfirmationStart.Before(config.DataEnd) {
		return Result{}, errors.New("confirmation start must precede data end")
	}
	sessions, manifest, err := loadSessions(config)
	if err != nil {
		return Result{}, err
	}
	result := Result{Protocol: FrozenProtocol(config)}
	result.Design = Artifact{SchemaVersion: SchemaVersion, Block: "design", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	result.Confirmation = Artifact{SchemaVersion: SchemaVersion, Block: "confirmation", RuleFrozenAt: config.RuleFrozenAt.UTC(), Source: manifest}
	for _, current := range sessions {
		cases, replayErr := replaySession(ctx, current)
		if replayErr != nil {
			return Result{}, fmt.Errorf("replay session %s: %w", shortHash(current.digest), replayErr)
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
		return Result{}, fmt.Errorf("evaluate design block: %w", err)
	}
	result.ConfirmReport, err = evaluateArtifact(result.Confirmation)
	if err != nil {
		return Result{}, fmt.Errorf("evaluate confirmation block: %w", err)
	}
	return result, nil
}

func loadSessions(config Config) ([]session, SourceManifest, error) {
	entries, err := os.ReadDir(config.SessionDir)
	if err != nil {
		return nil, SourceManifest{}, fmt.Errorf("read session directory: %w", err)
	}
	var sessions []session
	var manifest SourceManifest
	var acceptedDigests []string
	for _, entry := range entries {
		if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !uuidFilePattern.MatchString(entry.Name()) {
			continue
		}
		manifest.FilesScanned++
		current, err := readSession(filepath.Join(config.SessionDir, entry.Name()), config.DataEnd)
		if err != nil {
			return nil, SourceManifest{}, err
		}
		if len(current.messages) == 0 {
			continue
		}
		first := current.messages[0].available
		last := current.messages[len(current.messages)-1].available
		switch {
		case last.Before(config.ConfirmationStart):
			current.block = "design"
			manifest.FilesDesign++
		case !first.Before(config.ConfirmationStart):
			current.block = "confirmation"
			manifest.FilesConfirmation++
		default:
			manifest.FilesCrossingCutoff++
			continue
		}
		manifest.FilesAccepted++
		manifest.MessagesAccepted += len(current.messages)
		acceptedDigests = append(acceptedDigests, current.digest)
		sessions = append(sessions, current)
	}
	sort.Strings(acceptedDigests)
	aggregate := sha256.Sum256([]byte(strings.Join(acceptedDigests, "\n")))
	manifest.AggregateSHA256 = hex.EncodeToString(aggregate[:])
	sort.Slice(sessions, func(i, j int) bool {
		if sessions[i].messages[0].available.Equal(sessions[j].messages[0].available) {
			return sessions[i].digest < sessions[j].digest
		}
		return sessions[i].messages[0].available.Before(sessions[j].messages[0].available)
	})
	return sessions, manifest, nil
}

func readSession(path string, dataEnd time.Time) (session, error) {
	file, err := os.Open(path)
	if err != nil {
		return session{}, fmt.Errorf("open session: %w", err)
	}
	defer file.Close()
	digest := sha256.New()
	scanner := bufio.NewScanner(io.TeeReader(file, digest))
	scanner.Buffer(make([]byte, 64*1024), 32*1024*1024)
	var messages []message
	line := 0
	for scanner.Scan() {
		line++
		var record rawRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return session{}, fmt.Errorf("parse session line %d: %w", line, err)
		}
		if record.Type != "message" || (record.Message.Role != "user" && record.Message.Role != "assistant") {
			continue
		}
		available, err := parseTimestamp(record.Timestamp)
		if err != nil {
			return session{}, fmt.Errorf("parse session timestamp at line %d: %w", line, err)
		}
		if !available.Before(dataEnd) {
			continue
		}
		text := boundedText(extractText(record.Message.Content))
		if strings.TrimSpace(text) == "" {
			continue
		}
		anchors := ExtractAnchors(text)
		messages = append(messages, message{id: record.ID, role: record.Message.Role, text: text, query: text, queryAt: available, available: available, anchors: anchors, labelAnchors: anchors})
	}
	if err := scanner.Err(); err != nil {
		return session{}, fmt.Errorf("scan session: %w", err)
	}
	fileDigest := hex.EncodeToString(digest.Sum(nil))
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].available.Before(messages[j].available) })
	for index := range messages {
		if index > 0 && !messages[index].available.After(messages[index-1].available) {
			messages[index].available = messages[index-1].available.Add(time.Nanosecond)
		}
		messages[index].id = digestID("event", fileDigest, messages[index].id, fmt.Sprint(index))
	}
	return session{digest: fileDigest, messages: messages}, nil
}

func parseTimestamp(raw json.RawMessage) (time.Time, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, value)
}

func extractText(raw json.RawMessage) string {
	var direct string
	if json.Unmarshal(raw, &direct) == nil {
		return direct
	}
	var parts []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) != nil {
		return ""
	}
	var texts []string
	for _, part := range parts {
		if part.Type == "text" && strings.TrimSpace(part.Text) != "" {
			texts = append(texts, part.Text)
		}
	}
	return strings.Join(texts, "\n")
}

func boundedText(text string) string {
	if len(text) <= maxTextBytes {
		return text
	}
	return text[:maxTextBytes]
}

func ExtractAnchors(text string) map[string]struct{} {
	anchors := make(map[string]struct{})
	add := func(kind, value string) {
		value = strings.ToLower(strings.TrimRight(value, ".,;:!?)]}"))
		if len(value) >= 8 {
			anchors[kind+":"+value] = struct{}{}
		}
	}
	for _, value := range urlPattern.FindAllString(text, -1) {
		add("url", value)
	}
	for _, match := range pathPattern.FindAllStringSubmatch(text, -1) {
		add("path", match[1])
	}
	for _, value := range codePattern.FindAllString(text, -1) {
		add("code", value)
	}
	return anchors
}

func replaySession(ctx context.Context, current session) ([]evaluation.Case, error) {
	var cases []evaluation.Case
	clusterID := digestID("source-session-cluster", current.digest)
	for start := 0; start < len(current.messages); start += maxSegmentMessages {
		if current.maxSegments > 0 && start/maxSegmentMessages >= current.maxSegments {
			break
		}
		end := min(start+maxSegmentMessages, len(current.messages))
		segment := current.messages[start:end]
		segmentID := digestID("trajectory", current.digest, fmt.Sprint(start/maxSegmentMessages))
		segmentCases, err := replaySegment(ctx, current.block, segmentID, segment, current.negativeControl, current.ablations, current.embedder, current.baselineCalibration, current.predictiveCalibration, current.candidateRanker, current.candidateRetriever, current.textIndexer, current.retrievalNamespace)
		if err != nil {
			return nil, err
		}
		offset := time.Duration(start*2) * time.Second
		for index := range segmentCases {
			segmentCases[index].TrajectoryID = clusterID
			shiftCaseTimes(&segmentCases[index], offset)
		}
		cases = append(cases, segmentCases...)
	}
	return cases, nil
}

func shiftCaseTimes(item *evaluation.Case, offset time.Duration) {
	item.PredictedAt = item.PredictedAt.Add(offset)
	item.OutcomeAvailableAt = item.OutcomeAvailableAt.Add(offset)
	for name, forecast := range item.Variants {
		forecast.StateAsOf = forecast.StateAsOf.Add(offset)
		for index := range forecast.Candidates {
			forecast.Candidates[index].SourceAvailableAt = forecast.Candidates[index].SourceAvailableAt.Add(offset)
		}
		item.Variants[name] = forecast
	}
}

func replaySegment(ctx context.Context, block, trajectoryID string, messages []message, negativeControl, ablations bool, activeEmbedder embed.Embedder, baselineCalibration, predictiveCalibration calibration.Logit, candidateRanker retrieval.CandidateRanker, candidateRetriever retrieval.CandidateRetriever, textIndexer retrieval.TextIndexer, retrievalNamespace string) ([]evaluation.Case, error) {
	if activeEmbedder == nil {
		activeEmbedder, _ = embed.NewHashEmbedder(embeddingDimension)
	}
	if candidateRetriever != nil {
		candidateRetriever = newCachedCandidateRetriever(candidateRetriever)
	}
	if candidateRanker != nil {
		candidateRanker = newCachedCandidateRanker(candidateRanker)
	}
	runners, err := newRunners(negativeControl, ablations, activeEmbedder, baselineCalibration, predictiveCalibration, candidateRanker, candidateRetriever)
	if err != nil {
		return nil, err
	}
	defer func() {
		for _, current := range runners {
			_ = current.runtime.Close()
		}
	}()
	collection := "eventframe-eval-" + shortHash(digestID(retrievalNamespace, trajectoryID))
	for index, item := range messages {
		event := eventFromMessage(trajectoryID, index, item)
		if textIndexer != nil {
			metadata, marshalErr := json.Marshal(map[string]any{
				"collection": collection, "ts": event.AvailableAt.UnixMilli(), "session_id": trajectoryID,
				"user_id": defaultTenant, "authored": false, "access_count": 0,
				"authority": event.Priority, "salience": event.Priority,
			})
			if marshalErr != nil {
				return nil, fmt.Errorf("encode replay retrieval metadata: %w", marshalErr)
			}
			if insertErr := textIndexer.InsertText(ctx, collection, retrieval.Candidate{ID: event.ID, Text: event.FrameText(), Metadata: metadata}); insertErr != nil {
				return nil, fmt.Errorf("seed LibraVDB retrieval contract: %w", insertErr)
			}
		}
		vector, embedErr := embed.Document(activeEmbedder, event.FrameText())
		if embedErr != nil {
			return nil, fmt.Errorf("embed stored turn: %w", embedErr)
		}
		event.Embedding, event.EmbeddingModel = vector, activeEmbedder.ModelKey()
		for _, current := range runners {
			if _, err := current.runtime.Observe(ctx, model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event}); err != nil {
				return nil, fmt.Errorf("seed %s: %w", current.name, err)
			}
		}
	}
	var cases []evaluation.Case
	for index, query := range messages {
		if query.role != "user" || len(query.labelAnchors) == 0 {
			continue
		}
		universe := eligiblePrior(messages[:index], query.queryAt)
		if len(universe) < minimumPriorEvents {
			continue
		}
		var relevant []string
		for _, candidate := range universe {
			if intersects(query.labelAnchors, candidate.anchors) {
				relevant = append(relevant, candidate.id)
			}
		}
		if len(relevant) == 0 {
			continue
		}
		runtimeAsOf := query.queryAt.Add(-time.Nanosecond)
		exportedPredictedAt := exportedTime(index)
		queryText := frame.QueryText(query.query)
		queryVector, embedErr := embed.Query(activeEmbedder, queryText)
		if embedErr != nil {
			return nil, fmt.Errorf("embed replay query: %w", embedErr)
		}
		item := evaluation.Case{
			ID: digestID("case", trajectoryID, query.id), TrajectoryID: trajectoryID,
			PredictedAt: exportedPredictedAt, OutcomeAvailableAt: exportedPredictedAt.Add(time.Second), Priority: priorityFor(query.anchors),
			UniverseEventIDs: messageIDs(universe), RelevantEventIDs: relevant, Variants: make(map[string]evaluation.VariantForecast),
		}
		for _, current := range runners {
			if current.feedback != "none" {
				if err := publishFrontierCertificates(ctx, current, trajectoryID); err != nil {
					return nil, fmt.Errorf("refresh %s certificates: %w", current.name, err)
				}
			}
			recallRequest := model.RecallRequest{
				ProtocolVersion: model.ProtocolVersion, TenantID: defaultTenant, SessionID: trajectoryID,
				Query: query.query, Embedding: queryVector, EmbeddingModel: activeEmbedder.ModelKey(), AsOf: runtimeAsOf,
				RecallK: maxSegmentMessages, PackK: 10, TokenBudget: 2_000,
			}
			if candidateRetriever != nil {
				recallRequest.RetrievalCollections = []string{collection}
				recallRequest.RetrievalExcludeByCollection = map[string][]string{collection: messageIDs(messages[index:])}
			}
			packet, err := current.runtime.Recall(ctx, recallRequest)
			if err != nil {
				return nil, fmt.Errorf("%s recall: %w", current.name, err)
			}
			forecast, err := forecastFromPacket(packet, universe)
			if err != nil {
				return nil, fmt.Errorf("%s forecast: %w", current.name, err)
			}
			forecast.StateAsOf = exportedPredictedAt
			item.Variants[current.name] = forecast
			if current.feedback == "none" {
				continue
			}
			relevantSet := stringSet(relevant)
			if current.feedback == "shuffled" {
				relevantSet = shuffledRelevantSet(item.ID, universe, len(relevant))
			}
			if packet.BayesianShadow.Activated != packet.BayesianShadow.Nominated {
				return nil, fmt.Errorf("%s activated %d of contract-nominated frontier %d", current.name, packet.BayesianShadow.Activated, packet.BayesianShadow.Nominated)
			}
			for _, decision := range packet.BayesianShadow.Decisions {
				_, useful := relevantSet[decision.EventID]
				_, err := current.runtime.ObserveBayesianOutcome(ctx, model.BayesianOutcomeRequest{
					ProtocolVersion: model.ProtocolVersion, IdempotencyKey: digestID("outcome", current.name, item.ID, decision.EventID),
					TenantID: defaultTenant, JournalID: packet.BayesianShadow.JournalID, EventID: decision.EventID, Useful: useful,
					ObservedAt: query.available, AvailableAt: query.available, Source: model.OutcomeFullStream, InclusionProbability: 1,
				})
				if err != nil {
					return nil, fmt.Errorf("%s outcome: %w", current.name, err)
				}
			}
		}
		cases = append(cases, item)
	}
	return cases, nil
}

func newRunners(negativeControl, ablations bool, activeEmbedder embed.Embedder, baselineCalibration, predictiveCalibration calibration.Logit, candidateRanker retrieval.CandidateRanker, candidateRetriever retrieval.CandidateRetriever) ([]variantRunner, error) {
	if candidateRanker == nil {
		candidateRanker = retrieval.PassthroughRanker{}
	}
	definitions := []struct {
		name         string
		feedback     string
		residual     residual.Policy
		residualMode string
		ranking      ranking.Policy
	}{
		{name: "baseline", feedback: "none", residual: disabledResidual(), residualMode: service.ResidualModeDisabled},
		{name: "update_all", feedback: "all", residual: disabledResidual(), residualMode: service.ResidualModeDisabled},
		{name: "eventframe", feedback: "all", residual: residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .10, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001}, residualMode: service.ResidualModeApply},
	}
	if ablations {
		contextual := ranking.DefaultPolicy()
		contextual.ContextualEnabled = true
		hierarchy := ranking.DefaultPolicy()
		hierarchy.HierarchicalEnabled = true
		fullRanking := ranking.DefaultPolicy()
		fullRanking.ContextualEnabled, fullRanking.HierarchicalEnabled = true, true
		definitions = append(definitions,
			struct {
				name, feedback string
				residual       residual.Policy
				residualMode   string
				ranking        ranking.Policy
			}{name: "contextual", feedback: "all", residual: disabledResidual(), residualMode: service.ResidualModeDisabled, ranking: contextual},
			struct {
				name, feedback string
				residual       residual.Policy
				residualMode   string
				ranking        ranking.Policy
			}{name: "hierarchy", feedback: "all", residual: disabledResidual(), residualMode: service.ResidualModeDisabled, ranking: hierarchy},
			struct {
				name, feedback string
				residual       residual.Policy
				residualMode   string
				ranking        ranking.Policy
			}{name: "residual_shadow", feedback: "all", residual: residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .10, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001}, residualMode: service.ResidualModeShadow},
			struct {
				name, feedback string
				residual       residual.Policy
				residualMode   string
				ranking        ranking.Policy
			}{name: "full_upgrade", feedback: "all", residual: residual.Policy{Clip: .15, MinSupport: 3, MinConfidence: .55, ConfidenceDelta: .05, MotionLimit: .10, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001}, residualMode: service.ResidualModeShadow, ranking: fullRanking},
			struct {
				name, feedback string
				residual       residual.Policy
				residualMode   string
				ranking        ranking.Policy
			}{name: "shuffled_full", feedback: "shuffled", residual: disabledResidual(), residualMode: service.ResidualModeDisabled, ranking: fullRanking},
		)
	}
	if negativeControl {
		definitions = append(definitions, struct {
			name         string
			feedback     string
			residual     residual.Policy
			residualMode string
			ranking      ranking.Policy
		}{name: "shuffled_feedback", feedback: "shuffled", residual: disabledResidual(), residualMode: service.ResidualModeDisabled})
	}
	runners := make([]variantRunner, 0, len(definitions))
	for _, definition := range definitions {
		memory := memorystore.New()
		policy := bayes.Policy{VectorWeight: .6, NeighborWeight: .15, NoveltyWeight: .15, IndependenceWeight: .1, Threshold: 0, CriticalThreshold: 0, AuditProbability: 1, MaxActive: maxSegmentMessages, AuditSeed: "production-replay-frontier-all-v1"}
		if definition.feedback == "none" {
			policy.Threshold, policy.CriticalThreshold = 2, 2
		}
		runtime, err := service.New(memory, activeEmbedder, service.Config{
			DefaultRecallK: maxSegmentMessages, DefaultPackK: maxSegmentMessages, DefaultTokenBudget: 10_000_000, OverfetchMultiplier: 2,
			BayesianPolicy: policy, BayesianScoreWeight: .10,
			BaselineCalibration: baselineCalibration, PredictiveCalibration: predictiveCalibration,
			BayesianChangePolicy: bayes.ChangePolicy{Hazard: .05, Threshold: .30, MaxRun: 64, FastRate: .25, SlowRate: .025, DriftThreshold: .30, DriftPersistence: 12, MinSamples: 20, CUSUMSlack: .10, CUSUMThreshold: 8, CooldownSamples: 20},
			ResidualPolicy:       definition.residual,
			ResidualMode:         definition.residualMode,
			RankingPolicy:        definition.ranking,
			CandidateRanker:      candidateRanker, CandidateRankerRequired: true,
			CandidateRetriever: candidateRetriever, CandidateRetrieverRequired: candidateRetriever != nil,
		})
		if err != nil {
			return nil, err
		}
		runners = append(runners, variantRunner{name: definition.name, runtime: runtime, memory: memory, embedder: activeEmbedder, feedback: definition.feedback})
	}
	return runners, nil
}

func publishFrontierCertificates(ctx context.Context, current variantRunner, trajectoryID string) error {
	now := time.Now().UTC()
	snapshot := current.memory.Snapshot(ctx)
	suffix := shortHash(digestID(current.name, trajectoryID, fmt.Sprint(snapshot.EvidenceEpoch)))
	selection := model.SelectionSupportCertificate{
		ID: "production-replay-selection-" + suffix, TenantID: defaultTenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		MinSelectionProbability: 1, SimultaneousCoverage: 1, Procedure: "conditional contract-frontier certificate: every candidate returned by frozen SearchTextCollections is activated; no claim covers non-nominated corpus events",
		Issuer: "eventframe-production-eval", ExternalAudit: true, ValidFrom: now.Add(-time.Minute), ValidUntil: now.Add(time.Hour),
	}
	if _, err := current.runtime.PublishSelectionCertificate(ctx, model.PublishSelectionCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: selection}); err != nil {
		return err
	}
	omitted := model.OmittedInfluenceCertificate{
		ID: "production-replay-omitted-" + suffix, TenantID: defaultTenant, PolicyVersion: snapshot.PolicyVersion, EvidenceEpoch: snapshot.EvidenceEpoch,
		DivergenceUCB: 0, DivergenceLimit: 0, AuditProbability: 1, SimultaneousCoverage: 1,
		Procedure: "update-all over the contract-nominated bounded frontier; no returned frontier update is omitted and no claim covers non-nominated corpus events", Issuer: "eventframe-production-eval", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	_, err := current.runtime.PublishOmittedInfluenceCertificate(ctx, model.PublishOmittedInfluenceCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: omitted})
	return err
}

func forecastFromPacket(packet model.ContextPacket, universe []message) (evaluation.VariantForecast, error) {
	decisions := make(map[string]model.BayesianDecision, len(packet.BayesianShadow.Decisions))
	for _, decision := range packet.BayesianShadow.Decisions {
		decisions[decision.EventID] = decision
	}
	packed := make(map[string]struct{}, len(packet.Candidates))
	for _, candidate := range packet.Candidates {
		packed[candidate.Event.ID] = struct{}{}
	}
	forecast := evaluation.VariantForecast{Candidates: make([]evaluation.CandidateForecast, 0, len(universe)), PackedCount: packet.Packed, UsedTokens: packet.UsedTokens, AdaptiveExpanded: packet.AdaptiveExpanded}
	for index, source := range universe {
		decision, ok := decisions[source.id]
		if !ok {
			forecast.Candidates = append(forecast.Candidates, evaluation.CandidateForecast{
				EventID: source.id, SourceAvailableAt: exportedTime(index), Probability: 0, RankScore: 0,
				Nominated: false, Activated: false, Packed: false,
			})
			continue
		}
		if decision.Forecast.ModelKind == "" {
			return evaluation.VariantForecast{}, fmt.Errorf("event %s lacks forecast commitment", shortHash(source.id))
		}
		_, isPacked := packed[source.id]
		forecast.Candidates = append(forecast.Candidates, evaluation.CandidateForecast{
			EventID: source.id, SourceAvailableAt: exportedTime(index), Probability: decision.Forecast.CorrectedLaw.Useful, RankScore: decision.Forecast.RankScore,
			Nominated: true, Activated: decision.Activated, Packed: isPacked,
		})
	}
	sort.SliceStable(forecast.Candidates, func(i, j int) bool {
		if forecast.Candidates[i].RankScore == forecast.Candidates[j].RankScore {
			if forecast.Candidates[i].SourceAvailableAt.Equal(forecast.Candidates[j].SourceAvailableAt) {
				return forecast.Candidates[i].EventID < forecast.Candidates[j].EventID
			}
			return forecast.Candidates[i].SourceAvailableAt.After(forecast.Candidates[j].SourceAvailableAt)
		}
		return forecast.Candidates[i].RankScore > forecast.Candidates[j].RankScore
	})
	return forecast, nil
}

func eventFromMessage(sessionID string, sequence int, item message) model.Event {
	event := frame.FromText(item.text, item.role, sessionID, item.available, model.SourceObserved)
	event.ID, event.TenantID, event.SessionID, event.Sequence, event.Kind = item.id, defaultTenant, sessionID, uint64(sequence+1), "conversation_message"
	event.OccurredAt, event.ObservedAt, event.AvailableAt = item.available, item.available, item.available
	event.Priority = priorityFor(item.anchors)
	event.Provenance = model.Provenance{Producer: "openclaw-production-session-read-only-replay"}
	event.Attributes = map[string]string{"role": item.role, "content_truncated": fmt.Sprint(len(item.text) == maxTextBytes), "semantic_extractor": "fivew1h-deterministic-v1"}
	return event
}

func evaluateArtifact(artifact Artifact) (evaluation.Report, error) {
	return evaluateArtifactWithBaseline(artifact, "baseline")
}

func evaluateArtifactWithBaseline(artifact Artifact, baseline string) (evaluation.Report, error) {
	if len(artifact.Cases) == 0 {
		return evaluation.Report{
			SchemaVersion: evaluation.SchemaVersion, EvaluationBlock: artifact.Block,
			Variants: map[string]evaluation.Metrics{}, Comparisons: map[string]evaluation.Comparison{},
		}, nil
	}
	dataset := evaluation.Dataset{
		SchemaVersion: evaluation.SchemaVersion, EvaluationBlock: artifact.Block, BaselineVariant: baseline,
		PolicyFrozenAt: artifact.Cases[0].PredictedAt.Add(-time.Second), PriorityWeightScale: 4,
		BootstrapSamples: bootstrapSamples, BootstrapSeed: bootstrapSeed, Cases: artifact.Cases,
	}
	return evaluation.Evaluate(dataset)
}

func EvaluateArtifactWithBaseline(artifact Artifact, baseline string) (evaluation.Report, error) {
	return evaluateArtifactWithBaseline(artifact, baseline)
}

func WriteJSON(path string, value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(payload, '\n'), 0o600)
}

func messageIDs(messages []message) []string {
	ids := make([]string, len(messages))
	for index := range messages {
		ids[index] = messages[index].id
	}
	return ids
}

func eligiblePrior(messages []message, queryAt time.Time) []message {
	eligible := make([]message, 0, len(messages))
	for _, current := range messages {
		if current.available.Before(queryAt) {
			eligible = append(eligible, current)
		}
	}
	return eligible
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func shuffledRelevantSet(caseID string, universe []message, positives int) map[string]struct{} {
	ids := messageIDs(universe)
	sort.Slice(ids, func(i, j int) bool {
		return digestID("shuffled-feedback", caseID, ids[i]) < digestID("shuffled-feedback", caseID, ids[j])
	})
	if positives > len(ids) {
		positives = len(ids)
	}
	return stringSet(ids[:positives])
}

func intersects(left, right map[string]struct{}) bool {
	if len(left) > len(right) {
		left, right = right, left
	}
	for value := range left {
		if _, ok := right[value]; ok {
			return true
		}
	}
	return false
}

func priorityFor(anchors map[string]struct{}) float64 {
	return min(1, .5+.1*float64(min(len(anchors), 5)))
}

func disabledResidual() residual.Policy {
	return residual.Policy{Clip: .15, MinSupport: 1e9, MinConfidence: 1, ConfidenceDelta: .05, MotionLimit: 0, MaxAge: 30 * 24 * time.Hour, ImprovementDelta: .001}
}

func digestID(parts ...string) string {
	digest := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(digest[:])
}

func shortHash(value string) string {
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func sortCases(cases []evaluation.Case) {
	sort.SliceStable(cases, func(i, j int) bool {
		if cases[i].TrajectoryID == cases[j].TrajectoryID {
			return cases[i].PredictedAt.Before(cases[j].PredictedAt)
		}
		if cases[i].PredictedAt.Equal(cases[j].PredictedAt) {
			return cases[i].TrajectoryID < cases[j].TrajectoryID
		}
		return cases[i].PredictedAt.Before(cases[j].PredictedAt)
	})
}

func exportedTime(messageIndex int) time.Time {
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(messageIndex*2) * time.Second)
}
