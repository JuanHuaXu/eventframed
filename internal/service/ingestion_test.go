package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
)

type ingestionCards struct {
	response   retrieval.GetUserCardResponse
	tenant, id string
}

type ingestionContractIndex struct {
	*recordingIndex
	*ingestionCards
}

type unavailableIngestionCards struct{ calls int }

func (c *unavailableIngestionCards) GetUserCard(ctx context.Context, _ string, request retrieval.GetUserCardRequest) (retrieval.GetUserCardResponse, error) {
	c.calls++
	if request.UserID == "account:a" {
		return retrieval.GetUserCardResponse{CardJSON: `{"name":"Alex"}`, Version: 1, UpdatedAt: time.Now().Add(-time.Minute).UnixMilli()}, nil
	}
	<-ctx.Done()
	return retrieval.GetUserCardResponse{}, ctx.Err()
}

func TestCaptureIdentityLookupBudgetAndIncompleteRoster(t *testing.T) {
	events := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	cards := &unavailableIngestionCards{}
	runtime, err := service.New(events, embedder, service.Config{UserCards: cards, DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2000})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	turn := identityTurn("partial", "Alex will deploy.", time.Now().UTC())
	turn.ParticipantUserIDs = []string{"account:b"}
	captureIdentity(t, runtime, turn)
	if cards.calls != 2 {
		t.Fatalf("lookup calls = %d", cards.calls)
	}
	stored, err := events.GetEvents(context.Background(), turn.TenantID, []string{turn.ID}, turn.AvailableAt)
	if err != nil {
		t.Fatal(err)
	}
	var resolution frame.WhoResolution
	_ = json.Unmarshal([]byte(stored[0].Attributes["who_resolution"]), &resolution)
	if resolution.Status != "alias-roster-incomplete" || resolution.UserID != "" {
		t.Fatalf("timeout claimed unique alias: %+v", resolution)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	turn.ID = "cancelled"
	if _, err := runtime.CaptureTurn(ctx, model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: turn.ID, Turn: turn}); !errors.Is(err, context.Canceled) {
		t.Fatalf("parent cancellation ignored: %v", err)
	}
}

func (c *ingestionCards) GetUserCard(_ context.Context, tenant string, request retrieval.GetUserCardRequest) (retrieval.GetUserCardResponse, error) {
	c.tenant, c.id = tenant, request.UserID
	return c.response, nil
}

func identityTurn(id, text string, at time.Time) model.TurnCapture {
	return model.TurnCapture{ID: id, TenantID: "tenant-a", SessionID: "session-a", UserID: "account:a", AgentID: "main", Sequence: uint64(at.UnixMilli()), UserText: text, AssistantText: "Ready.", OccurredAt: at, ObservedAt: at, AvailableAt: at}
}
func captureIdentity(t *testing.T, runtime *service.Service, turn model.TurnCapture) model.ObserveResponse {
	t.Helper()
	response, err := runtime.CaptureTurn(context.Background(), model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: turn.ID, Turn: turn})
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func TestCaptureIdentityDurableAndRetryIndexesCommittedFrame(t *testing.T) {
	ctx := context.Background()
	embedder, err := embed.NewHashEmbedder(8)
	if err != nil {
		t.Fatal(err)
	}
	settings := libravdbstore.Config{Path: filepath.Join(t.TempDir(), "events.libravdb"), Dimension: 8, Quantization: "none", EmbeddingModel: embedder.ModelKey()}
	events, err := libravdbstore.Open(settings)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Truncate(time.Millisecond)
	cards := &ingestionCards{response: retrieval.GetUserCardResponse{CardJSON: `{"name":"Example Operator","aliases":["Alex"]}`, Version: 1, UpdatedAt: now.Add(-time.Minute).UnixMilli()}}
	index := &recordingIndex{}
	// Match the daemon's wiring: the existing contract index also provides the
	// user-card reader, without a separate explicit UserCards configuration.
	runtime, err := service.New(events, embedder, service.Config{CandidateIndex: &ingestionContractIndex{index, cards}, CandidateRetriever: &fixedRetriever{}, CandidateCollectionPrefix: "eventframe-", DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2000})
	if err != nil {
		t.Fatal(err)
	}
	runtimeClosed := false
	t.Cleanup(func() {
		if !runtimeClosed {
			_ = runtime.Close()
		}
	})
	turn := identityTurn("identity-turn", "I will deploy using my console because they need it.", now)
	captureIdentity(t, runtime, turn)
	firstText := index.ensured.Text
	cards.response.CardJSON = `{"name":"Changed Operator"}`
	cards.response.Version = 2
	if !captureIdentity(t, runtime, turn).Duplicate {
		t.Fatal("retry not duplicate")
	}
	if index.ensured.Text != firstText || strings.Contains(index.ensured.Text, "Changed Operator") {
		t.Fatalf("retry replaced indexed frame: %s", index.ensured.Text)
	}
	if cards.tenant != turn.TenantID || cards.id != turn.UserID {
		t.Fatal("wrong card namespace/key")
	}
	if err := runtime.Close(); err != nil {
		t.Fatal(err)
	}
	runtimeClosed = true
	events, err = libravdbstore.Open(settings)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = events.Close() })
	stored, err := events.GetEvents(ctx, turn.TenantID, []string{turn.ID}, now)
	if err != nil || len(stored) != 1 {
		t.Fatalf("read = %v, %v", stored, err)
	}
	var resolution frame.WhoResolution
	if err := json.Unmarshal([]byte(stored[0].Attributes["who_resolution"]), &resolution); err != nil {
		t.Fatal(err)
	}
	if stored[0].Who.Value != "Example Operator [user:account:a]" || resolution.CardVersion != 1 || !strings.Contains(stored[0].Content, "my console") {
		t.Fatalf("stored identity = %+v / %+v", stored[0], resolution)
	}
	var unresolved frame.UnresolvedReferences
	if err := json.Unmarshal([]byte(stored[0].Attributes["unresolved_references"]), &unresolved); err != nil || unresolved.Status != "needs-resolution" || unresolved.Count == 0 {
		t.Fatalf("unresolved-reference metadata lost on durable reopen: %+v / %v", unresolved, err)
	}
}

func TestCaptureIdentityRetryKeepsCommittedUnresolvedReferences(t *testing.T) {
	ctx := context.Background()
	events := memorystore.New()
	embedder, _ := embed.NewHashEmbedder(8)
	index := &recordingIndex{}
	runtime, err := service.New(events, embedder, service.Config{CandidateIndex: index, CandidateRetriever: &fixedRetriever{}, CandidateCollectionPrefix: "eventframe-", DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2000})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = runtime.Close() })
	now := time.Now().UTC().Truncate(time.Millisecond)
	turn := identityTurn("unresolved-first", "He will deploy using his console.", now)
	turn.UserID, turn.PreviousTurnID = "", "later-imported-predecessor"
	captureIdentity(t, runtime, turn)
	initial, err := events.GetEvents(ctx, turn.TenantID, []string{turn.ID}, now)
	if err != nil || len(initial) != 1 {
		t.Fatalf("initial read = %v / %v", initial, err)
	}
	initialReport := initial[0].Attributes["unresolved_references"]
	initialIndexText := index.ensured.Text
	var report frame.UnresolvedReferences
	if err := json.Unmarshal([]byte(initialReport), &report); err != nil || report.Count == 0 {
		t.Fatalf("missing initial markers: %+v / %v", report, err)
	}
	previous := identityTurn(turn.PreviousTurnID, "Alex will deploy.", now.Add(-time.Minute))
	previous.UserID = ""
	captureIdentity(t, runtime, previous)
	if !captureIdentity(t, runtime, turn).Duplicate {
		t.Fatal("retry unexpectedly created a new event")
	}
	stored, err := events.GetEvents(ctx, turn.TenantID, []string{turn.ID}, now)
	if err != nil || len(stored) != 1 || stored[0].Attributes["unresolved_references"] != initialReport || stored[0].Who.Value != initial[0].Who.Value || index.ensured.Text != initialIndexText {
		t.Fatalf("retry rewrote committed references: %v / %v / %s", stored, err, index.ensured.Text)
	}
	// A new capture can use the now-available predecessor; retry immutability
	// must not disable fresh contextual resolution.
	turn.ID = "resolved-fresh"
	turn.Sequence++
	captureIdentity(t, runtime, turn)
	stored, err = events.GetEvents(ctx, turn.TenantID, []string{turn.ID}, now)
	if err != nil || len(stored) != 1 || stored[0].Who.Value != "Alex" {
		t.Fatalf("fresh resolution failed: %v / %v", stored, err)
	}
	if err := json.Unmarshal([]byte(stored[0].Attributes["unresolved_references"]), &report); err != nil || report.Count != 0 {
		t.Fatalf("fresh resolved references remained marked: %+v / %v", report, err)
	}
}

func TestCaptureIdentityContextAndFutureCardBoundaries(t *testing.T) {
	for _, boundary := range []string{"valid", "different-tenant", "different-session", "future-context", "future-card", "missing-context"} {
		t.Run(boundary, func(t *testing.T) {
			events := memorystore.New()
			embedder, _ := embed.NewHashEmbedder(8)
			now := time.Now().UTC().Truncate(time.Millisecond)
			cards := &ingestionCards{response: retrieval.GetUserCardResponse{CardJSON: `{"name":"Alex"}`, Version: 1, UpdatedAt: now.Add(-time.Minute).UnixMilli()}}
			runtime, err := service.New(events, embedder, service.Config{UserCards: cards, DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2000})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = runtime.Close() })
			previous := identityTurn("previous", "Alex will deploy.", now.Add(-time.Minute))
			if boundary == "different-session" {
				previous.SessionID = "session-other"
			}
			if boundary == "different-tenant" {
				previous.TenantID = "tenant-other"
			}
			if boundary == "future-context" {
				previous = identityTurn("previous", "Alex will deploy.", now.Add(time.Minute))
			}
			captureIdentity(t, runtime, previous)
			turn := identityTurn("current", "He will use his console.", now)
			turn.PreviousTurnID = previous.ID
			if boundary == "missing-context" {
				turn.PreviousTurnID = "missing"
			}
			if boundary == "future-card" {
				cards.response.UpdatedAt = now.Add(time.Minute).UnixMilli()
			}
			captureIdentity(t, runtime, turn)
			stored, err := events.GetEvents(context.Background(), turn.TenantID, []string{turn.ID}, now)
			if err != nil {
				t.Fatal(err)
			}
			var resolution frame.WhoResolution
			_ = json.Unmarshal([]byte(stored[0].Attributes["who_resolution"]), &resolution)
			if boundary == "valid" && (stored[0].Who.Value != "Alex [user:account:a]" || resolution.CardVersion != 1 || resolution.ContextEventID != previous.ID) {
				t.Fatalf("valid context = %+v", resolution)
			}
			if boundary == "future-card" && (resolution.CardVersion != 0 || !strings.Contains(stored[0].Attributes["user_card_lookup_status"], "future-version")) {
				t.Fatalf("future identity leaked: %+v", resolution)
			}
			if boundary != "valid" && boundary != "future-card" && stored[0].Who.Confidence != 0 {
				t.Fatalf("ineligible context used: %+v", stored[0].Who)
			}
		})
	}
}
