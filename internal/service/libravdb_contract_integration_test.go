package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/embed"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store/memorystore"
	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestProductionLibraVDBContractsEndToEnd(t *testing.T) {
	endpoint := os.Getenv("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT")
	if endpoint == "" {
		t.Skip("EVENTFRAMED_TEST_LIBRAVDB_ENDPOINT is not set")
	}
	contracts, err := retrieval.OpenLibraVDBContracts(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	defer contracts.Close()
	embedder, _ := embed.NewHashEmbedder(8)
	runtime, err := service.New(memorystore.New(), embedder, service.Config{
		DefaultRecallK: 5, DefaultPackK: 2, DefaultTokenBudget: 200, OverfetchMultiplier: 2,
		CandidateRetriever: contracts, CandidateRetrieverRequired: true,
		CandidateRanker: contracts, CandidateRankerRequired: true,
		CandidateIndex: contracts, CandidateCollectionPrefix: "eventframe-e2e-",
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	event := testutil.Event("contract-e2e-"+now.Format("150405.000000000"), "magnetic teapot contract end to end", now.Add(-time.Minute))
	observe(t, runtime, event)
	request := model.RecallRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, SessionID: event.SessionID,
		Query: event.Content, AsOf: now, RecallK: 5, PackK: 2, TokenBudget: 200,
	}
	packet, err := runtime.Recall(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 1 || packet.Candidates[0].Event.ID != event.ID {
		t.Fatalf("contract packet = %#v", packet.Candidates)
	}
	if packet.NominationContract != contracts.RetrievalContractName() || packet.RetrievalContract != contracts.ContractName() {
		t.Fatalf("contract identities = %q / %q", packet.NominationContract, packet.RetrievalContract)
	}
	if _, err := runtime.Delete(context.Background(), model.DeleteRequest{
		ProtocolVersion: model.ProtocolVersion, TenantID: event.TenantID, EventID: event.ID,
	}); err != nil {
		t.Fatal(err)
	}
	packet, err = runtime.Recall(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(packet.Candidates) != 0 {
		t.Fatalf("deleted contract event was returned: %#v", packet.Candidates)
	}
}
