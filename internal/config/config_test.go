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
}
