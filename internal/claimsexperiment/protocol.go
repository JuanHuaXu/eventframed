package claimsexperiment

const (
	ProtocolID               = "eventframe.claims-experiment.protocol.v7-rescue-replacement"
	IntervalMethod           = "wilson_score_two_sided"
	MonitoringInterpretation = "fixed_terminal_sample_not_sequential_confidence_sequence"
	DetectionMatchingRule    = "chronological_one_to_one_declared_change_to_first_unmatched_trigger_in_window"
	DetectionWindowEndRule   = "inclusive"
	DetectedChangeDelayBasis = "matched_detected_changes_only"
	defaultConfidenceLevel   = .95
)

func FrozenProtocol() ExperimentProtocol {
	return ExperimentProtocol{
		ID:                          ProtocolID,
		Scope:                       "rescue/replacement synthetic mechanism criteria; old failed propositions remain failed and these are not production thresholds",
		ConfidenceLevel:             defaultConfidenceLevel,
		IntervalMethod:              IntervalMethod,
		MonitoringInterpretation:    MonitoringInterpretation,
		DetectionMatching:           DetectionMatchingRule,
		DetectionWindowEnd:          DetectionWindowEndRule,
		MeanDelayBasis:              DetectedChangeDelayBasis,
		ResidualReplacementCriteria: ResidualReplacementCriterion{MinimumGainLower: 0, MaximumWorstTrajectoryExcess: .02},
		OmittedAuditCriteria:        OmittedAuditCriterion{MinimumCoverageLower: .95},
		GroupCriteria: map[string]GroupAcceptanceCriterion{
			"shared_0.8_0.8":  {MinimumExpectedDecisionRate: .80, MaximumWrongRateUpper: .06},
			"split_0.65_0.35": {MinimumExpectedDecisionRate: .80, MaximumWrongRateUpper: .06},
			"split_0.9_0.1":   {MinimumExpectedDecisionRate: .95, MaximumWrongRateUpper: .06},
		},
		ChangepointCriteria: map[string]ChangepointAcceptanceCriterion{
			"stable":              {MaximumUnmatchedAlarmsPerTrajectory: .02},
			"abrupt_noiseless":    {MinimumDetectionRate: number(.99), MaximumUnmatchedAlarmsPerTrajectory: 0, MaximumMeanDelay: number(0)},
			"abrupt":              {MinimumDetectionRate: number(.80), MaximumUnmatchedAlarmsPerTrajectory: .30, MaximumMeanDelay: number(15)},
			"gradual":             {MinimumDetectionRate: number(.90), MaximumUnmatchedAlarmsPerTrajectory: .20, MaximumMeanDelay: number(50)},
			"recurring_noiseless": {MinimumDetectionRate: number(.99), MaximumUnmatchedAlarmsPerTrajectory: 0, MaximumMeanDelay: number(0)},
			"recurring":           {MinimumDetectionRate: number(.75), MaximumUnmatchedAlarmsPerTrajectory: .35, MaximumMeanDelay: number(15)},
		},
	}
}

func number(value float64) *float64 { return &value }
