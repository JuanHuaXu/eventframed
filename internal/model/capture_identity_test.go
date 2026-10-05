package model_test

import (
	"strings"
	"testing"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

func TestCaptureIdentityEnvelopeValidation(t *testing.T) {
	base := model.TurnCapture{ID: "turn", TenantID: "tenant", SessionID: "session", Sequence: 1, UserText: "request", AssistantText: "response", OccurredAt: time.Now(), ObservedAt: time.Now().Add(time.Second), AvailableAt: time.Now().Add(2 * time.Second)}
	for _, test := range []struct {
		name   string
		change func(*model.TurnCapture)
	}{
		{"self-reference", func(c *model.TurnCapture) { c.PreviousTurnID = c.ID }},
		{"blank-account", func(c *model.TurnCapture) { c.UserID = " " }},
		{"control-account", func(c *model.TurnCapture) { c.UserID = "account:\na" }},
		{"oversize-account", func(c *model.TurnCapture) { c.UserID = strings.Repeat("a", 1025) }},
		{"duplicate-roster", func(c *model.TurnCapture) { c.ParticipantUserIDs = []string{"a", "a"} }},
		{"oversize-roster", func(c *model.TurnCapture) { c.ParticipantUserIDs = []string{"a", "b", "c", "d", "e", "f", "g", "h"} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := base
			test.change(&c)
			if err := c.Validate(); err == nil {
				t.Fatal("invalid identity accepted")
			}
		})
	}
	base.UserID, base.ParticipantUserIDs, base.PreviousTurnID = "account:a", []string{"account:b"}, "previous"
	if err := base.Validate(); err != nil {
		t.Fatal(err)
	}
}
