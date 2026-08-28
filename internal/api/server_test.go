package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/api"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestObserveThenRecallOverHTTP(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(api.NewServer(runtime, logger).Handler())
	t.Cleanup(server.Close)

	now := time.Now().UTC()
	event := testutil.Event("http-event", "the deployment key is blue", now.Add(-time.Second))
	var observed model.ObserveResponse
	postJSON(t, server.URL+"/v1/events:observe", model.ObserveRequest{
		ProtocolVersion: model.ProtocolVersion, IdempotencyKey: event.ID, Event: event,
	}, &observed)
	if observed.EventID != event.ID {
		t.Fatalf("observe response = %+v", observed)
	}

	var packet model.ContextPacket
	postJSON(t, server.URL+"/v1/context:recall", model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, SessionID: event.SessionID,
		Query: "deployment key", AsOf: now, RecallK: 50, PackK: 10, TokenBudget: 200,
	}, &packet)
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != event.ID {
		t.Fatalf("recall response = %+v", packet)
	}
}

func TestRejectsUnknownJSONFields(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	response, err := http.Post(server.URL+"/v1/context:recall", "application/json", bytes.NewBufferString(`{"unknown":true}`))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func postJSON(t *testing.T, url string, input, output any) {
	t.Helper()
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("content-type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status %d: %s", response.StatusCode, body)
	}
	if err := json.NewDecoder(response.Body).Decode(output); err != nil {
		t.Fatal(err)
	}
}
