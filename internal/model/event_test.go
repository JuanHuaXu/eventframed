package model_test

import (
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/testutil"
)

func TestEventValidateRejectsTimeTravelAtIngest(t *testing.T) {
	event := testutil.Event("event-1", "content", time.Now().UTC())
	event.AvailableAt = event.ObservedAt.Add(-time.Nanosecond)
	if err := event.Validate(8); err == nil {
		t.Fatal("expected availability ordering error")
	}
}

func TestEventValidateRejectsWrongEmbeddingDimension(t *testing.T) {
	event := testutil.Event("event-1", "content", time.Now().UTC())
	event.Embedding = []float32{1, 2}
	if err := event.Validate(8); err == nil {
		t.Fatal("expected embedding dimension error")
	}
}

func TestEventValidateRequiresProvenanceForNonEmptyFields(t *testing.T) {
	event := testutil.Event("event-1", "content", time.Now().UTC())
	event.What.Source = ""
	if err := event.Validate(8); err == nil {
		t.Fatal("expected field source error")
	}
	event = testutil.Event("event-1", "content", time.Now().UTC())
	event.Provenance.Producer = ""
	if err := event.Validate(8); err == nil {
		t.Fatal("expected producer error")
	}
}
