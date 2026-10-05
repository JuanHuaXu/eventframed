package researchbatch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/frame"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"github.com/JuanHuaXu/eventframed/internal/store/libravdbstore"
)

type parityEmbedder struct{ hash *embed.HashEmbedder }

func (e parityEmbedder) Embed(text string) ([]float32, error) { return e.hash.Embed(text) }
func (e parityEmbedder) Dimension() int                       { return 4 }
func (e parityEmbedder) Name() string                         { return "research" }
func (e parityEmbedder) ModelKey() string                     { return "research:d4" }

func rawTurnDigestForParity(turn model.TurnCapture) (string, error) {
	encoded, err := json.Marshal(struct {
		Contract string            `json:"contract"`
		Turn     model.TurnCapture `json:"turn"`
	}{Contract: "eventframe-raw-turn-v1", Turn: turn})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func TestBatchIntentCaptureTurnConversionParity(t *testing.T) {
	ctx := context.Background()
	control, err := libravdbstore.Open(libravdbstore.Config{
		Path: filepath.Join(t.TempDir(), "control.libravdb"), Dimension: 4,
		EmbeddingModel: "research:d4", Quantization: "none", MemoryMapping: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	hash, err := embed.NewHashEmbedder(4)
	if err != nil {
		t.Fatal(err)
	}
	embedder := parityEmbedder{hash: hash}
	runtime, err := service.New(control, embedder, service.Config{
		DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	controlBase := control.Snapshot(ctx)
	protoRoot := t.TempDir()
	proto := openPrototypeForTest(t, protoRoot, true)
	protoBase := proto.snapshot
	at := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	texts := []string{
		"Example Operator will review a sensor reading in Toronto because the old reading is stale.",
		"The support team will compare two reports in Boston using the incident timeline.",
		"An agent will verify the launch date at Cape Canaveral before answering.",
		"The analyst will inspect the revised budget in Seattle after the update.",
	}
	turns := make([]model.TurnCapture, len(texts))
	writes := make([]libravdbstore.ResearchEventWrite, len(texts))
	ids := make([]string, len(texts))
	for i, userText := range texts {
		id := "turn-" + string(rune('a'+i))
		turn := model.TurnCapture{
			ID: id, TenantID: "tenant-a", SessionID: "session-a", Sequence: uint64(i + 1),
			UserText: userText, AssistantText: "I will check the cited source and report the result.",
			OccurredAt: at.Add(-time.Duration(i+2) * time.Minute),
			ObservedAt: at.Add(-time.Duration(i+2)*time.Minute + time.Second), AvailableAt: at,
		}
		turns[i], ids[i] = turn, id
		request := model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Turn: turn}
		response, err := runtime.CaptureTurn(ctx, request)
		if err != nil || response.Duplicate {
			t.Fatal("control capture failed", id, response, err)
		}
		event := frame.FromTurn(turn)
		if err := event.Validate(embedder.Dimension()); err != nil {
			t.Fatal("independent frame invalid", err)
		}
		vector, err := embed.Document(embedder, event.FrameText())
		if err != nil {
			t.Fatal(err)
		}
		wrongVector, err := embed.Document(embedder, event.Content)
		if err != nil || len(wrongVector) != len(vector) {
			t.Fatal("negative-control embedding failed", err)
		}
		if reflect.DeepEqual(wrongVector, vector) {
			t.Fatal("full-text embedding negative control is non-discriminating")
		}
		event.EmbeddingModel = embedder.ModelKey()
		digest, err := rawTurnDigestForParity(turn)
		if err != nil {
			t.Fatal(err)
		}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		eventSum := sha256.Sum256(eventJSON)
		if digest == hex.EncodeToString(eventSum[:]) {
			t.Fatal("raw-turn and EventFrame digests unexpectedly coincide")
		}
		writes[i] = libravdbstore.ResearchEventWrite{Event: event, Vector: vector, Digest: digest}
	}
	if control.Snapshot(ctx).RuntimeVersion != controlBase.RuntimeVersion+uint64(len(turns)) {
		t.Fatal("control did not record one version per capture")
	}
	results, err := proto.apply(ctx, writes, "")
	if err != nil || len(results) != len(writes) {
		t.Fatal("batch conversion write failed", err)
	}
	if proto.snapshot.RuntimeVersion != protoBase.RuntimeVersion+uint64(len(writes)) {
		t.Fatal("prototype lost an accepted version")
	}
	compare := func(p *batchPrototype) {
		t.Helper()
		want, err := control.GetEventsWithVectors(ctx, "tenant-a", ids, at)
		if err != nil {
			t.Fatal(err)
		}
		got, err := p.backend.GetEventsWithVectors(ctx, "tenant-a", ids, at)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			left, leftErr := json.Marshal(want[i])
			right, rightErr := json.Marshal(got[i])
			if leftErr != nil || rightErr != nil {
				t.Fatal("event comparison could not serialize", ids[i], leftErr, rightErr)
			}
			if !bytes.Equal(left, right) || got[i].Content != "User: "+turns[i].UserText+"\n\nAssistant: "+turns[i].AssistantText {
				t.Fatal("control and batch payload/vector differ", ids[i])
			}
		}
	}
	compare(proto)
	if _, err := control.PutResearchEventBatch(ctx, writes); err != nil {
		t.Fatal("prepared raw-turn digest disagrees with CaptureTurn", err)
	}
	if err := proto.close(); err != nil {
		t.Fatal(err)
	}
	proto = openPrototypeForTest(t, protoRoot, false)
	defer proto.close()
	compare(proto)
	retry, err := proto.apply(ctx, writes, "")
	if err != nil || len(retry) != len(writes) {
		t.Fatal("reopened batch retry failed", err)
	}
	for _, result := range retry {
		if !result.Duplicate {
			t.Fatal("exact raw-turn retry was not duplicate")
		}
	}
	changed := turns[0]
	changed.UserText += " The source was corrected."
	badDigest, err := rawTurnDigestForParity(changed)
	if err != nil {
		t.Fatal(err)
	}
	bad := writes[0]
	bad.Digest = badDigest
	bad.Event = frame.FromTurn(changed)
	bad.Event.EmbeddingModel = embedder.ModelKey()
	bad.Vector, err = embed.Document(embedder, bad.Event.FrameText())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := proto.apply(ctx, []libravdbstore.ResearchEventWrite{bad}, ""); !errors.Is(err, store.ErrIdempotencyConflict) {
		t.Fatal("changed raw turn did not conflict", err)
	}
}
