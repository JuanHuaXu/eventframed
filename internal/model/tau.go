package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// TauForecast is an as-of finite marked-time law. NoEvent includes no marked
// outcome inside HorizonMS; a caller must not renormalize it away.
type TauPoint struct {
	Mark        string  `json:"mark"`
	DelayMS     int64   `json:"delay_ms"`
	Probability float64 `json:"probability"`
}

func (p *TauPoint) UnmarshalJSON(data []byte) error {
	type plain TauPoint
	var out plain
	if err := decodeTauRequired(data, &out, []string{"delay_ms", "probability"}); err != nil {
		return err
	}
	*p = TauPoint(out)
	return nil
}

type TauForecast struct {
	Base       Snapshot   `json:"base"`
	ID         string     `json:"id"`
	Origin     time.Time  `json:"origin"`
	IssuedAt   time.Time  `json:"issued_at"`
	ValidUntil time.Time  `json:"valid_until"`
	HorizonMS  int64      `json:"horizon_ms"`
	Points     []TauPoint `json:"points"`
	NoEvent    float64    `json:"no_event"`
	TVRadius   float64    `json:"tv_radius"`
}

type TauTurnPolicy struct {
	CandidateDelaysMS *[]int64 `json:"candidate_delays_ms,omitempty"`
	QuietStartUTC     *int     `json:"quiet_start_utc,omitempty"`
	QuietEndUTC       *int     `json:"quiet_end_utc,omitempty"`
	TargetKind        string   `json:"target_kind"`
	Benefit           float64  `json:"benefit"`
	FalseCost         float64  `json:"false_cost"`
	TurnCost          float64  `json:"turn_cost"`
	MinProbability    float64  `json:"min_probability"`
	Margin            float64  `json:"margin"`
	LeadMS            int64    `json:"lead_ms"`
	WindowMS          int64    `json:"window_ms"`
	SlackMS           int64    `json:"slack_ms"`
	MinDelayMS        int64    `json:"min_delay_ms"`
	CooldownMS        int64    `json:"cooldown_ms"`
}

type TauDecision struct {
	Schedule         bool      `json:"schedule"`
	Reason           string    `json:"reason"`
	DeliverAt        time.Time `json:"deliver_at,omitempty"`
	LatestDispatch   time.Time `json:"latest_dispatch,omitempty"`
	Probability      float64   `json:"probability"`
	LowerProbability float64   `json:"lower_probability"`
	LowerUtility     float64   `json:"lower_utility"`
}

// The signed binding survives restart with its agency proposal. Any non-agency
// mutation invalidates this conservative first implementation's forecast.
type TauTurnBinding struct {
	RequestDigest    string      `json:"request_digest"`
	Forecast         TauForecast `json:"forecast"`
	Decision         TauDecision `json:"decision"`
	PlanID           string      `json:"plan_id"`
	TargetKind       string      `json:"target_kind"`
	ForecastID       string      `json:"forecast_id"`
	ForecastDigest   string      `json:"forecast_digest"`
	PolicyDigest     string      `json:"policy_digest"`
	Base             Snapshot    `json:"base"`
	LatestDispatch   time.Time   `json:"latest_dispatch"`
	LowerProbability float64     `json:"lower_probability"`
	LowerUtility     float64     `json:"lower_utility"`
	CooldownMS       int64       `json:"cooldown_ms"`
}

// Explicit zero is meaningful; omitted or null numerical declarations are not
// permission to assume perfect forecasts or free interruptions.
func (f *TauForecast) UnmarshalJSON(data []byte) error {
	type plain TauForecast
	var out plain
	if err := decodeTauRequired(data, &out, []string{"horizon_ms", "no_event", "tv_radius"}); err != nil {
		return err
	}
	*f = TauForecast(out)
	return nil
}

func (p *TauTurnPolicy) UnmarshalJSON(data []byte) error {
	type plain TauTurnPolicy
	var out plain
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, field := range []string{"quiet_start_utc", "quiet_end_utc"} {
		if value, exists := raw[field]; exists && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("tau %s cannot be null", field)
		}
	}
	if value, exists := raw["candidate_delays_ms"]; exists {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("tau candidate list cannot be null")
		}
		var offsets []json.RawMessage
		if err := json.Unmarshal(value, &offsets); err != nil {
			return err
		}
		for _, offset := range offsets {
			if bytes.Equal(bytes.TrimSpace(offset), []byte("null")) {
				return fmt.Errorf("tau candidate offset cannot be null")
			}
		}
	}
	if err := decodeTauRequired(data, &out, []string{"benefit", "false_cost", "turn_cost", "min_probability", "margin", "lead_ms", "window_ms", "slack_ms", "min_delay_ms", "cooldown_ms"}); err != nil {
		return err
	}
	*p = TauTurnPolicy(out)
	return nil
}

func decodeTauRequired(data []byte, out any, fields []string) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	for _, field := range fields {
		v, ok := raw[field]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return fmt.Errorf("explicit tau %s required", field)
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	return d.Decode(out)
}

// Episodes are explicit closed observation windows, never mined from a chat's
// prose. A no-event episode requires its own horizon-complete evidence frame.
type TauEpisode struct {
	AnchorID  string `json:"anchor_id"`
	OutcomeID string `json:"outcome_id"`
	NoEvent   bool   `json:"no_event"`
}

type PlanTauTurnRequest struct {
	ProtocolVersion string        `json:"protocol_version"`
	IssuerToken     string        `json:"issuer_token"`
	TenantID        string        `json:"tenant_id"`
	SessionID       string        `json:"session_id"`
	PlanID          string        `json:"plan_id"`
	OriginEventID   string        `json:"origin_event_id"`
	EvidenceIDs     []string      `json:"evidence_ids"`
	Base            Snapshot      `json:"base"`
	Forecast        TauForecast   `json:"forecast"`
	Policy          TauTurnPolicy `json:"policy"`
	Priority        float64       `json:"priority"`
	Reason          string        `json:"reason"`
	Episodes        []TauEpisode  `json:"episodes,omitempty"`
	BinDelaysMS     []int64       `json:"bin_delays_ms,omitempty"`
}

type PlanTauTurnResponse struct {
	ProtocolVersion string                       `json:"protocol_version"`
	Forecast        TauForecast                  `json:"forecast"`
	Decision        TauDecision                  `json:"decision"`
	Proposal        *IssueAgencyProposalResponse `json:"proposal,omitempty"`
	GuaranteeScope  string                       `json:"guarantee_scope"`
}
