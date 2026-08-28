package config

import (
	"path/filepath"
	"testing"
)

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
