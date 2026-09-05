package config

import (
	"path/filepath"
	"testing"
)

func TestWorkingBeliefRequiresExplicitTrustAndComposition(t *testing.T) {
	for _, args := range [][]string{{"--working-belief"}, {"--working-belief", "--evidence-trust-file", "keys.json", "--hierarchical-posterior"}} {
		if _, err := Parse(args); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	c, err := Parse([]string{"--working-belief", "--evidence-trust-file", "keys.json"})
	if err != nil || !c.WorkingBelief || c.EvidenceTrustFile != "keys.json" {
		t.Fatalf("valid config: %+v %v", c, err)
	}
}

func TestAgencySecretPathsMustBePairwiseDistinct(t *testing.T) {
	root := t.TempDir()
	privateKey := filepath.Join(root, "agency.key")
	publicKey := filepath.Join(root, "agency.pub")
	if _, err := Parse([]string{
		"-agency-enabled",
		"-agency-private-key", privateKey,
		"-agency-public-key", publicKey,
		"-agency-issuer-token", publicKey,
	}); err == nil {
		t.Fatal("public agency key was accepted as the issuer token")
	}
	if _, err := Parse([]string{
		"-agency-enabled",
		"-agency-private-key", privateKey,
		"-agency-public-key", publicKey,
		"-agency-issuer-token", filepath.Join(root, "issuer.token"),
	}); err != nil {
		t.Fatalf("distinct agency paths were rejected: %v", err)
	}
	if _, err := Parse([]string{
		"-agency-enabled",
		"-agency-private-key", privateKey,
		"-agency-public-key", publicKey,
		"-agency-issuer-token", filepath.Join(root, "issuer.token"),
		"-agency-authority-token", filepath.Join(root, "nested", "..", "agency.pub"),
	}); err == nil {
		t.Fatal("clean-path alias between public key and authority token was accepted")
	}
}

func TestAgencyModeRejectsPlaintextTCPListener(t *testing.T) {
	if _, err := Parse([]string{"-agency-enabled", "-listen", "tcp://127.0.0.1:8080"}); err == nil {
		t.Fatal("agency mode accepted a plaintext TCP listener")
	}
	if _, err := Parse([]string{"-listen", "tcp://127.0.0.1:8080"}); err != nil {
		t.Fatalf("non-agency TCP test listener was rejected: %v", err)
	}
}

func TestResidualModeRejectsUnknownBehavior(t *testing.T) {
	if _, err := Parse([]string{"-residual-mode", "sometimes"}); err == nil {
		t.Fatal("unknown residual mode was accepted")
	}
}

func TestLibraVDBContractEndpointAndDeprecatedAlias(t *testing.T) {
	config, err := Parse([]string{"-libravdb-contract-endpoint", "unix:/tmp/contracts.sock"})
	if err != nil || config.LibraVDBContractEndpoint != "unix:/tmp/contracts.sock" {
		t.Fatalf("contract endpoint = %q, %v", config.LibraVDBContractEndpoint, err)
	}
	legacy, err := Parse([]string{"-libravdb-ranker-endpoint", "tcp:127.0.0.1:50051"})
	if err != nil || legacy.LibraVDBContractEndpoint != "tcp:127.0.0.1:50051" {
		t.Fatalf("legacy endpoint = %q, %v", legacy.LibraVDBContractEndpoint, err)
	}
	if _, err := Parse([]string{
		"-libravdb-contract-endpoint", "unix:/tmp/contracts.sock",
		"-libravdb-ranker-endpoint", "unix:/tmp/other.sock",
	}); err == nil {
		t.Fatal("conflicting contract endpoints were accepted")
	}
}

func TestLibraVDBContractTLSConfiguration(t *testing.T) {
	config, err := Parse([]string{
		"-libravdb-contract-endpoint", "tcp:memory.example:443",
		"-libravdb-contract-tls-mode", "tls",
		"-libravdb-contract-tls-ca", "/tmp/ca.pem",
		"-libravdb-contract-tls-client-cert", "/tmp/client.pem",
		"-libravdb-contract-tls-client-key", "/tmp/client.key",
	})
	if err != nil || config.LibraVDBContractTLSMode != "tls" || config.LibraVDBContractTLSCA != "/tmp/ca.pem" {
		t.Fatalf("TLS configuration = %#v, %v", config, err)
	}
	if _, err := Parse([]string{"-libravdb-contract-tls-mode", "opportunistic"}); err == nil {
		t.Fatal("unknown TLS mode was accepted")
	}
	if _, err := Parse([]string{"-libravdb-contract-tls-client-cert", "/tmp/client.pem"}); err == nil {
		t.Fatal("client certificate without a key was accepted")
	}
	if _, err := Parse([]string{"-libravdb-contract-tls-mode", "insecure", "-libravdb-contract-tls-ca", "/tmp/ca.pem"}); err == nil {
		t.Fatal("TLS CA was accepted in insecure mode")
	}
}

func TestLibraVDBResilienceControlsAreBounded(t *testing.T) {
	config, err := Parse([]string{
		"-libravdb-contract-concurrency", "8", "-libravdb-contract-timeout-ms", "750",
		"-libravdb-contract-attempts", "2", "-libravdb-circuit-failures", "3", "-libravdb-circuit-cooldown-ms", "1500",
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.LibraVDBContractConcurrency != 8 || config.LibraVDBContractTimeoutMS != 750 || config.LibraVDBCircuitFailures != 3 {
		t.Fatalf("unexpected resilience config: %#v", config)
	}
	for _, args := range [][]string{
		{"-libravdb-contract-concurrency", "0"}, {"-libravdb-contract-timeout-ms", "0"},
		{"-libravdb-contract-attempts", "5"}, {"-libravdb-circuit-failures", "0"}, {"-libravdb-circuit-cooldown-ms", "0"},
	} {
		if _, parseErr := Parse(args); parseErr == nil {
			t.Fatalf("invalid resilience controls accepted: %v", args)
		}
	}
}

func TestRankDeltaStoreRequiresDistinctAbsolutePathAndPositiveCache(t *testing.T) {
	root := t.TempDir()
	database := filepath.Join(root, "events.libravdb")
	if _, err := Parse([]string{"-database", database, "-rank-delta-sqlite", "relative.sqlite"}); err == nil {
		t.Fatal("relative rank-delta SQLite path was accepted")
	}
	if _, err := Parse([]string{"-database", database, "-rank-delta-sqlite", database}); err == nil {
		t.Fatal("rank-delta SQLite path aliased the LibraVDB database")
	}
	if _, err := Parse([]string{"-rank-delta-cache-entries", "0"}); err == nil {
		t.Fatal("zero rank-delta cache capacity was accepted")
	}
}

func TestElasticRankDeltaConfiguration(t *testing.T) {
	config, err := Parse([]string{"-elastic-rank-delta-min-scale", "0.25", "-elastic-rank-delta-max-scale", "3"})
	if err != nil {
		t.Fatal(err)
	}
	if !config.ElasticRankDelta || config.ElasticRankDeltaMinScale != .25 || config.ElasticRankDeltaMaxScale != 3 {
		t.Fatalf("elastic rank-delta config = %#v", config)
	}
	if _, err := Parse([]string{"-elastic-rank-delta-min-scale", "2", "-elastic-rank-delta-max-scale", "1"}); err == nil {
		t.Fatal("inverted elastic rank-delta scales were accepted")
	}
	if _, err := Parse([]string{"-elastic-rank-delta-max-scale", "11"}); err == nil {
		t.Fatal("unbounded elastic rank-delta scale was accepted")
	}
}

func TestSharedEvidenceWeightIsBounded(t *testing.T) {
	configured, err := Parse([]string{"-shared-evidence-weight", "0.25"})
	if err != nil || configured.SharedEvidenceWeight != .25 {
		t.Fatalf("shared evidence weight = %v, %v", configured.SharedEvidenceWeight, err)
	}
	for _, value := range []string{"0", "1.1"} {
		if _, err := Parse([]string{"-shared-evidence-weight", value}); err == nil {
			t.Fatalf("invalid shared evidence weight %s was accepted", value)
		}
	}
}

func TestBackgroundFuzzConfigurationIsBoundedAndCanBeDisabled(t *testing.T) {
	configured, err := Parse([]string{"-background-fuzz-certainty", "0.3", "-background-fuzz-queue", "16", "-background-fuzz-interval-ms", "500"})
	if err != nil || !configured.BackgroundFuzz || configured.BackgroundFuzzCertainty != .3 || configured.BackgroundFuzzQueue != 16 || configured.BackgroundFuzzIntervalMS != 500 {
		t.Fatalf("background fuzz config = %#v, %v", configured, err)
	}
	disabled, err := Parse([]string{"-background-fuzz=false", "-background-fuzz-queue", "0"})
	if err != nil || disabled.BackgroundFuzz {
		t.Fatalf("disabled background fuzz config = %#v, %v", disabled, err)
	}
	for _, args := range [][]string{
		{"-background-fuzz-certainty", "0"}, {"-background-fuzz-queue", "0"},
		{"-background-fuzz-interval-ms", "0"}, {"-background-fuzz-max-events", "1"},
		{"-background-fuzz-max-trials", "513"},
	} {
		if _, parseErr := Parse(args); parseErr == nil {
			t.Fatalf("invalid background fuzz controls accepted: %v", args)
		}
	}
}
