package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/api"
	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
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
	var deleted model.DeleteResponse
	postJSON(t, server.URL+"/v1/events:delete", model.DeleteRequest{ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, EventID: event.ID}, &deleted)
	if !deleted.Deleted || deleted.Snapshot.ResidualVersion != 2 {
		t.Fatalf("delete response = %+v", deleted)
	}
	postJSON(t, server.URL+"/v1/context:recall", model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, SessionID: event.SessionID,
		Query: "deployment key", AsOf: now, RecallK: 50, PackK: 10, TokenBudget: 200,
	}, &packet)
	if len(packet.Candidates) != 0 {
		t.Fatalf("deleted event recalled: %+v", packet.Candidates)
	}
	metrics, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer metrics.Body.Close()
	body, _ := io.ReadAll(metrics.Body)
	if !bytes.Contains(body, []byte("eventframed_http_requests_total")) {
		t.Fatalf("metrics = %s", body)
	}
}

type failingReadinessProbe struct{}

func (failingReadinessProbe) Ready(context.Context) error { return errors.New("sidecar unavailable") }

type openCircuitRetriever struct{}

func (openCircuitRetriever) SearchTextCollections(context.Context, retrieval.SearchRequest) ([]retrieval.Candidate, error) {
	return nil, retrieval.ErrContractCircuitOpen
}
func (openCircuitRetriever) RetrievalContractName() string { return "test-open-circuit" }

func TestReadinessIncludesRequiredExternalContracts(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, ExternalReadiness: failingReadinessProbe{},
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)

	health, err := http.Get(server.URL + "/v1/health")
	if err != nil {
		t.Fatal(err)
	}
	health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Fatalf("liveness status = %d", health.StatusCode)
	}
	ready, err := http.Get(server.URL + "/v1/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer ready.Body.Close()
	if ready.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("readiness status = %d", ready.StatusCode)
	}
	var response model.ReadinessResponse
	if err := json.NewDecoder(ready.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "not_ready" || response.ExternalContractsReady || response.ExternalError == "" {
		t.Fatalf("readiness response = %+v", response)
	}
}

func TestRecallReturnsRetryableDependencyStatusForOpenCircuit(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000,
		CandidateRetriever: openCircuitRetriever{}, CandidateRetrieverRequired: true,
		CandidateCollectionPrefix: "eventframe-",
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	payload, _ := json.Marshal(model.RecallRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant", SessionID: "session", Query: "query", AsOf: time.Now().UTC(), RecallK: 10, PackK: 10, TokenBudget: 1000})
	response, err := http.Post(server.URL+"/v1/openclaw/context:recall", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable || response.Header.Get("Retry-After") != "5" {
		t.Fatalf("status=%d retry-after=%q", response.StatusCode, response.Header.Get("Retry-After"))
	}
}

func TestCaptureTurnEnrichesAfterHTTPContract(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memorystore.New(), embedder, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	now := time.Now().UTC().Add(-time.Minute)
	turn := model.TurnCapture{
		ID: "http-turn", TenantID: "tenant-a", SessionID: "session-a", Sequence: uint64(now.UnixMilli()),
		UserText:      "Example Operator will deploy EventFrame in Toronto tomorrow because retrieval is stale using the OpenClaw plugin.",
		AssistantText: "I prepared the release.", OccurredAt: now, ObservedAt: now.Add(time.Second), AvailableAt: now.Add(time.Second),
	}
	var observed model.ObserveResponse
	postJSON(t, server.URL+"/v1/turns:capture", model.CaptureTurnRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: turn.ID, Turn: turn}, &observed)
	if observed.EventID != turn.ID {
		t.Fatalf("capture response = %+v", observed)
	}
	var packet model.ContextPacket
	postJSON(t, server.URL+"/v1/context:recall", model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: turn.TenantID, SessionID: turn.SessionID,
		Query: "retrieval stale", AsOf: now.Add(2 * time.Second), RecallK: 10, PackK: 10, TokenBudget: 1000,
	}, &packet)
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.Where.Value != "Toronto" {
		t.Fatalf("captured frame = %+v", packet.Candidates)
	}
	raw, err := json.Marshal(model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: turn.TenantID, SessionID: turn.SessionID,
		Query: "retrieval stale", AsOf: now.Add(2 * time.Second), RecallK: 10, PackK: 10, TokenBudget: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(server.URL+"/v1/openclaw/context:recall", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var adapterPacket struct {
		Candidates []struct {
			Event map[string]any `json:"event"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(response.Body).Decode(&adapterPacket); err != nil {
		t.Fatal(err)
	}
	if len(adapterPacket.Candidates) != 1 {
		t.Fatalf("adapter candidates = %+v", adapterPacket.Candidates)
	}
	for _, field := range []string{"who", "what", "where", "when", "why", "how", "embedding", "embedding_model", "attributes"} {
		if _, exposed := adapterPacket.Candidates[0].Event[field]; exposed {
			t.Fatalf("OpenClaw projection exposed internal field %q", field)
		}
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

func TestRawTurnContractRejectsAdapterSuppliedSemanticFields(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memorystore.New(), embedder, service.Config{DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4})
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	body := `{"protocol_version":"eventframe.v1alpha1","idempotency_key":"turn","turn":{"id":"turn","tenant_id":"tenant","session_id":"session","sequence":1,"user_text":"question","assistant_text":"answer","occurred_at":"2026-01-01T00:00:00Z","observed_at":"2026-01-01T00:00:01Z","available_at":"2026-01-01T00:00:01Z","who":{"value":"adapter-authored"}}}`
	response, err := http.Post(server.URL+"/v1/turns:capture", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

func TestGetPredictiveGraphOverHTTP(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000, OverfetchMultiplier: 4,
	})
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)

	response, err := http.Get(server.URL + "/v1/abstraction/graph?tenant_id=tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status %d: %s", response.StatusCode, body)
	}
	var graph model.PredictiveGraphResponse
	if err := json.NewDecoder(response.Body).Decode(&graph); err != nil {
		t.Fatal(err)
	}
	if graph.ProtocolVersion != model.ProtocolVersion || graph.Graph.TenantID != "tenant-a" || graph.Graph.Version != graph.Snapshot.GraphVersion {
		t.Fatalf("graph response = %+v", graph)
	}
}

func TestBayesianGroupComparisonOverHTTPIsProposalOnly(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000,
	})
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)

	var comparison model.BayesianGroupComparison
	postJSON(t, server.URL+"/v1/bayesian/groups:compare", model.BayesianGroupComparisonRequest{
		ProtocolVersion: model.ProtocolVersion,
		TenantID:        "tenant-a",
		MemberEventIDs:  []string{"event-b", "event-a"},
	}, &comparison)
	if comparison.Recommendation != "uncertain" || !comparison.RequiresAntiPigeonCertification || comparison.Members[0].EventID != "event-a" {
		t.Fatalf("comparison = %+v", comparison)
	}
}

func TestInvariantCompositionLifecycleOverHTTP(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, _ := service.New(memorystore.New(), embedder, service.Config{DefaultRecallK: 10, DefaultPackK: 10, DefaultTokenBudget: 1000})
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	now := time.Now().UTC()
	var observed model.ObserveResponse
	for _, id := range []string{"http-a", "http-b"} {
		event := testutil.Event(id, "public mission stage", now.Add(-time.Minute))
		postJSON(t, server.URL+"/v1/events:observe", model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: id, Event: event}, &observed)
	}
	certificate := model.AntiPigeonCertificate{
		ID: "http-ap", TenantID: "tenant-a", MemberEventIDs: []string{"http-a", "http-b"}, HorizonKey: model.RetrievalUsefulnessHorizon,
		GraphVersion: observed.Snapshot.GraphVersion, EvidenceEpoch: observed.Snapshot.EvidenceEpoch,
		TargetDiameterUCB: .01, DiameterLimit: .05, EffectiveSupport: 40, MinEffectiveSupport: 30,
		SimultaneousCoverage: .95, Procedure: "external HTTP fixture audit", Issuer: "test-auditor", ExternalAudit: true, ValidUntil: now.Add(time.Hour),
	}
	var certified model.CertificateResponse
	postJSON(t, server.URL+"/v1/bayesian/certificates:publish-anti-pigeon", model.PublishAntiPigeonCertificateRequest{ProtocolVersion: model.ProtocolVersion, Certificate: certificate}, &certified)
	var composed model.ComposeInvariantResponse
	postJSON(t, server.URL+"/v1/invariants:compose", model.ComposeInvariantRequest{
		ProtocolVersion: model.ProtocolVersion, ID: "http-macro", TenantID: "tenant-a", SessionID: "session-a",
		MemberEventIDs: certificate.MemberEventIDs, RepresentativeEventID: "http-a", Label: "public mission",
		RuleID: "mission-stages", Resolution: "mission", Confidence: .9, AntiPigeonCertificateID: certificate.ID,
		PublishedAt: now, BaseSnapshot: certified.Snapshot,
	}, &composed)
	if composed.Event.Composition == nil || composed.Event.Composition.AntiPigeonCertificateID != certificate.ID {
		t.Fatalf("composed = %+v", composed)
	}
	var decomposed model.DecomposeInvariantResponse
	postJSON(t, server.URL+"/v1/invariants:decompose", model.DecomposeInvariantRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", EventID: composed.Event.ID, Reason: "HTTP lifecycle test"}, &decomposed)
	if !decomposed.Deleted || len(decomposed.RestoredMemberEventIDs) != 2 {
		t.Fatalf("decomposed = %+v", decomposed)
	}
}

func TestAgencyLifecycleOverHTTP(t *testing.T) {
	embedder, _ := embed.NewHashEmbedder(8)
	signer, err := agency.NewSignerForTest()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 50, DefaultPackK: 10, DefaultTokenBudget: 2_000,
		AgencyPolicy: agency.DefaultPolicy(true), AgencySigner: signer, AgencyIssuerToken: "test-issuer-token-that-is-at-least-32-bytes", AgencyAuthorityToken: "test-authority-token-that-is-at-least-32-bytes",
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(api.NewServer(runtime, slog.New(slog.NewTextHandler(io.Discard, nil))).Handler())
	t.Cleanup(server.Close)
	now := time.Now().UTC()
	evidence := testutil.Event("event-a", "agency evidence", now.Add(-time.Second))
	var observed model.ObserveResponse
	postJSON(t, server.URL+"/v1/events:observe", model.ObserveRequest{ProtocolVersion: model.ProtocolVersion, IdempotencyKey: evidence.ID, Event: evidence}, &observed)
	var issued model.IssueAgencyProposalResponse
	postJSON(t, server.URL+"/v1/agency/proposals:issue", model.IssueAgencyProposalRequest{ProtocolVersion: model.ProtocolVersion, IssuerToken: "test-issuer-token-that-is-at-least-32-bytes", Proposal: model.AgencyProposalDraft{
		ID: "proposal-http", TenantID: "tenant-a", SessionID: "openclaw:session-a", Action: model.AgencyNotify,
		Reason: "A follow-up became timely.", EvidenceIDs: []string{"event-a"}, ExpectedUtility: .8, Priority: .7,
		NotBefore: now, ExpiresAt: now.Add(time.Hour), IdempotencyKey: "proposal-http", CausalChainID: "chain-http",
	}}, &issued)
	if issued.Record.Status != model.AgencyPending || issued.Record.Signed.Signature == "" {
		t.Fatalf("issued = %+v", issued)
	}
	if status := postJSONStatus(t, server.URL+"/v1/agency/proposals:claim", model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, TenantID: "tenant-a", ConsumerID: "authority-http", Limit: 10}); status != http.StatusBadRequest {
		t.Fatalf("unauthenticated claim status = %d", status)
	}
	var claimed model.ClaimAgencyProposalsResponse
	postJSON(t, server.URL+"/v1/agency/proposals:claim", model.ClaimAgencyProposalsRequest{ProtocolVersion: model.ProtocolVersion, AuthorityToken: "test-authority-token-that-is-at-least-32-bytes", TenantID: "tenant-a", ConsumerID: "authority-http", Limit: 10}, &claimed)
	if len(claimed.Records) != 1 || claimed.Records[0].Status != model.AgencyClaimed {
		t.Fatalf("claimed = %+v", claimed)
	}
	var resolved model.ResolveAgencyProposalResponse
	postJSON(t, server.URL+"/v1/agency/proposals:resolve", model.ResolveAgencyProposalRequest{
		ProtocolVersion: model.ProtocolVersion, AuthorityToken: "test-authority-token-that-is-at-least-32-bytes", TenantID: "tenant-a", ProposalID: "proposal-http", ConsumerID: "authority-http",
		Decision: model.AgencyApproved, Reason: "authorized", ExecutionRef: "job-http",
	}, &resolved)
	if resolved.Record.Status != model.AgencyApproved || resolved.Record.ExecutionRef != "job-http" {
		t.Fatalf("resolved = %+v", resolved)
	}
}

func postJSONStatus(t *testing.T, url string, input any) int {
	t.Helper()
	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	return response.StatusCode
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
