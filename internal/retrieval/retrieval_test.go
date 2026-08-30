package retrieval

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPassthroughRankerHonorsSecondPassLimit(t *testing.T) {
	ranked, err := (PassthroughRanker{}).RankCandidates(context.Background(), RankRequest{
		Candidates: []Candidate{{ID: "a"}, {ID: "b"}, {ID: "c"}}, K1: 3, K2: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ranked) != 2 || ranked[0].ID != "a" || ranked[1].ID != "b" {
		t.Fatalf("unexpected pass-through rank: %+v", ranked)
	}
}

func TestContractGuardOpensAndRecoversAfterCooldown(t *testing.T) {
	guard := newContractGuard(ContractClientConfig{MaxConcurrent: 1, FailureThreshold: 2, OpenDuration: 20 * time.Millisecond})
	guard.failure()
	if guard.isOpen() {
		t.Fatal("circuit opened before the threshold")
	}
	guard.failure()
	if !guard.isOpen() {
		t.Fatal("circuit did not open at the threshold")
	}
	if err := guard.acquire(context.Background()); !errors.Is(err, ErrContractCircuitOpen) {
		t.Fatalf("open circuit acquire = %v", err)
	}
	time.Sleep(25 * time.Millisecond)
	if guard.isOpen() {
		t.Fatal("circuit did not enter half-open state after cooldown")
	}
}

func TestContractGuardBoundsConcurrentCalls(t *testing.T) {
	guard := newContractGuard(ContractClientConfig{MaxConcurrent: 1})
	if err := guard.acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := guard.acquire(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second acquire = %v", err)
	}
	guard.release()
}

func TestLibraVDBRankerRejectsInvalidEndpoint(t *testing.T) {
	if _, err := OpenLibraVDBRanker("unix:"); err == nil {
		t.Fatal("expected empty unix socket path to fail")
	}
}

func TestContractTLSAutoModePreservesLocalPlaintextOnly(t *testing.T) {
	for _, endpoint := range []string{"unix:/tmp/libravdb.sock", "tcp:127.0.0.1:50051", "tcp:[::1]:50051", "tcp:localhost:50051"} {
		if !localContractEndpoint(endpoint) {
			t.Fatalf("local endpoint %q classified as remote", endpoint)
		}
	}
	if localContractEndpoint("tcp:memory.example:443") {
		t.Fatal("remote endpoint classified as local")
	}
	if _, err := contractTransportCredentials(ContractClientConfig{TLSMode: "tls", CAFile: "/missing/ca.pem"}, "tcp:memory.example:443"); err == nil {
		t.Fatal("missing TLS CA was accepted")
	}
}
