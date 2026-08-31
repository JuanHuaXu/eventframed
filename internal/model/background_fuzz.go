package model

import "time"

type BackgroundFuzzResultSummary struct {
	JobID                 string    `json:"job_id"`
	Status                string    `json:"status"`
	TriggerReason         string    `json:"trigger_reason"`
	AnswerCertainty       float64   `json:"answer_certainty"`
	EventCount            int       `json:"event_count"`
	PerturbationCount     int       `json:"perturbation_count"`
	TrialCount            int       `json:"trial_count"`
	PropertyCount         int       `json:"property_count"`
	ConditionalInvariants int       `json:"conditional_invariants"`
	Error                 string    `json:"error,omitempty"`
	CompletedAt           time.Time `json:"completed_at"`
}

type BackgroundFuzzQueueStatus struct {
	ProtocolVersion   string                       `json:"protocol_version"`
	Enabled           bool                         `json:"enabled"`
	Capacity          int                          `json:"capacity"`
	QueueDepth        int                          `json:"queue_depth"`
	Running           int                          `json:"running"`
	EnqueuedTotal     uint64                       `json:"enqueued_total"`
	CompletedTotal    uint64                       `json:"completed_total"`
	FailedTotal       uint64                       `json:"failed_total"`
	StaleTotal        uint64                       `json:"stale_total"`
	DroppedTotal      uint64                       `json:"dropped_total"`
	DeduplicatedTotal uint64                       `json:"deduplicated_total"`
	LastResult        *BackgroundFuzzResultSummary `json:"last_result,omitempty"`
}
