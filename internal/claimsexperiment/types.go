package claimsexperiment

const SchemaVersion = "eventframe.claims-experiment.v7"

type SuiteReport struct {
	SchemaVersion           string                      `json:"schema_version"`
	Protocol                ExperimentProtocol          `json:"protocol"`
	Run                     RunManifest                 `json:"run"`
	Residual                ResidualReport              `json:"residual_utility"`
	ResidualHeterogeneous   HeterogeneousResidualReport `json:"residual_heterogeneous"`
	ResidualSafeReplacement SafeResidualReport          `json:"residual_safe_replacement"`
	OmittedAudit            OmittedAuditReport          `json:"omitted_influence_audit"`
	AntiPigeon              AntiPigeonReport            `json:"anti_pigeon_granularity"`
	GroupComparison         GroupComparisonReport       `json:"bayesian_group_comparison"`
	Changepoint             ChangepointReport           `json:"changepoint_adaptation"`
	IntegrationControls     IntegrationControlsReport   `json:"integration_controls"`
}

type OmittedAuditReport struct {
	Trials            int                `json:"trials"`
	PopulationSize    int                `json:"population_size"`
	AuditProbability  float64            `json:"audit_probability"`
	ConfidenceLevel   float64            `json:"confidence_level"`
	CoveredTrials     int                `json:"covered_trials"`
	Errors            int                `json:"errors"`
	CoverageRate      float64            `json:"coverage_rate"`
	CoverageInterval  ProportionInterval `json:"coverage_wilson_interval"`
	MeanTrueInfluence float64            `json:"mean_true_influence"`
	MeanUpperBound    float64            `json:"mean_upper_bound"`
	Acceptance        AcceptanceResult   `json:"acceptance"`
}

type SafeResidualReport struct {
	Trajectories            int                `json:"trajectories"`
	EvaluationCases         int                `json:"evaluation_cases"`
	BaselineBrier           float64            `json:"baseline_brier"`
	SafeResidualBrier       float64            `json:"safe_residual_brier"`
	MeanGain                float64            `json:"mean_gain"`
	GainInterval            ProportionInterval `json:"trajectory_bootstrap_gain_interval"`
	AppliedCases            int                `json:"applied_cases"`
	AbstainedCases          int                `json:"abstained_cases"`
	HarmfulAppliedCases     int                `json:"harmful_applied_cases"`
	HarmfulAppliedRate      float64            `json:"harmful_applied_rate"`
	WorstTrajectoryExcess   float64            `json:"worst_trajectory_excess_loss"`
	HarmBudgetPerTrajectory float64            `json:"harm_budget_per_trajectory"`
	Acceptance              AcceptanceResult   `json:"acceptance"`
}

type HeterogeneousResidualReport struct {
	Trajectories         int                `json:"trajectories"`
	StableTrajectories   int                `json:"stable_trajectories"`
	ReversedTrajectories int                `json:"reversed_trajectories"`
	EvaluationCases      int                `json:"evaluation_cases"`
	BaselineBrier        float64            `json:"baseline_brier"`
	ResidualBrier        float64            `json:"residual_brier"`
	MeanGain             float64            `json:"mean_gain"`
	GainInterval         ProportionInterval `json:"trajectory_bootstrap_gain_interval"`
	AppliedCases         int                `json:"applied_cases"`
	HarmfulReuseCases    int                `json:"harmful_reuse_cases"`
	HarmfulReuseRate     float64            `json:"harmful_reuse_rate"`
	HarmfulReuseInterval ProportionInterval `json:"harmful_reuse_wilson_interval"`
	MaintenanceUpdates   int                `json:"maintenance_updates"`
	Acceptance           AcceptanceResult   `json:"acceptance"`
}

type RunManifest struct {
	Role                     string `json:"role"`
	SeedBase                 int64  `json:"seed_base"`
	DesignSeedBase           *int64 `json:"design_seed_base,omitempty"`
	ProtocolFrozenBeforeRun  bool   `json:"protocol_frozen_before_run"`
	ConfirmationSeedDistinct *bool  `json:"confirmation_seed_distinct,omitempty"`
}

type ExperimentProtocol struct {
	ID                          string                                    `json:"id"`
	Scope                       string                                    `json:"scope"`
	ConfidenceLevel             float64                                   `json:"confidence_level"`
	IntervalMethod              string                                    `json:"interval_method"`
	MonitoringInterpretation    string                                    `json:"monitoring_interpretation"`
	DetectionMatching           string                                    `json:"detection_matching"`
	DetectionWindowEnd          string                                    `json:"detection_window_end"`
	MeanDelayBasis              string                                    `json:"mean_delay_basis"`
	GroupCriteria               map[string]GroupAcceptanceCriterion       `json:"group_acceptance_criteria"`
	ChangepointCriteria         map[string]ChangepointAcceptanceCriterion `json:"changepoint_acceptance_criteria"`
	ResidualReplacementCriteria ResidualReplacementCriterion              `json:"residual_replacement_criteria"`
	OmittedAuditCriteria        OmittedAuditCriterion                     `json:"omitted_audit_criteria"`
}

type ResidualReplacementCriterion struct {
	MinimumGainLower             float64 `json:"minimum_gain_lower"`
	MaximumWorstTrajectoryExcess float64 `json:"maximum_worst_trajectory_excess"`
}

type OmittedAuditCriterion struct {
	MinimumCoverageLower float64 `json:"minimum_coverage_lower"`
}

type GroupAcceptanceCriterion struct {
	MinimumExpectedDecisionRate float64 `json:"minimum_expected_decision_rate"`
	MaximumWrongRateUpper       float64 `json:"maximum_wrong_rate_wilson_upper"`
}

type ChangepointAcceptanceCriterion struct {
	MinimumDetectionRate                *float64 `json:"minimum_detection_rate,omitempty"`
	MaximumUnmatchedAlarmsPerTrajectory float64  `json:"maximum_unmatched_alarms_per_trajectory"`
	MaximumMeanDelay                    *float64 `json:"maximum_mean_delay_among_detections,omitempty"`
}

type ProportionInterval struct {
	Method          string  `json:"method"`
	ConfidenceLevel float64 `json:"confidence_level"`
	Lower           float64 `json:"lower"`
	Upper           float64 `json:"upper"`
}

type AcceptanceResult struct {
	Evaluated  bool     `json:"evaluated"`
	Passed     bool     `json:"passed"`
	Violations []string `json:"violations,omitempty"`
}

type ResidualReport struct {
	TrainingCases         int     `json:"training_cases"`
	EvaluationCases       int     `json:"evaluation_cases"`
	BaselineBrier         float64 `json:"baseline_brier"`
	ResidualBrier         float64 `json:"residual_brier"`
	AbsoluteGain          float64 `json:"absolute_gain"`
	RelativeReduction     float64 `json:"relative_reduction"`
	ResidualAppliedCases  int     `json:"residual_applied_cases"`
	ResidualAppliedRate   float64 `json:"residual_applied_rate"`
	SystematicOutcomeRate float64 `json:"systematic_outcome_rate"`
}

type AntiPigeonReport struct {
	TrainingTurns   int                          `json:"training_turns"`
	EvaluationTurns int                          `json:"evaluation_turns"`
	Variants        map[string]AntiPigeonVariant `json:"variants"`
}

type AntiPigeonVariant struct {
	Brier          float64 `json:"brier"`
	FalseMergeRate float64 `json:"false_merge_rate"`
	PosteriorKeys  int     `json:"posterior_keys"`
}

type ChangepointReport struct {
	Trajectories    int                        `json:"trajectories"`
	SeedBase        int64                      `json:"seed_base"`
	DetectionWindow int                        `json:"detection_window"`
	MatchingRule    string                     `json:"matching_rule"`
	MeanDelayBasis  string                     `json:"mean_delay_basis"`
	Policy          ChangepointPolicy          `json:"policy"`
	Scenarios       map[string]ChangepointCase `json:"scenarios"`
}

type ChangepointPolicy struct {
	Hazard           float64 `json:"hazard"`
	Threshold        float64 `json:"threshold"`
	MaxRun           int     `json:"max_run"`
	RecentWindow     int     `json:"recent_window"`
	RecentThreshold  float64 `json:"recent_threshold"`
	FastRate         float64 `json:"fast_rate"`
	SlowRate         float64 `json:"slow_rate"`
	DriftThreshold   float64 `json:"drift_threshold"`
	DriftPersistence int     `json:"drift_persistence"`
	MinSamples       int     `json:"min_samples"`
	CUSUMSlack       float64 `json:"cusum_slack"`
	CUSUMThreshold   float64 `json:"cusum_threshold"`
	CooldownSamples  int     `json:"cooldown_samples"`
}

type ChangepointCase struct {
	DetectionWindow              int                 `json:"detection_window"`
	ExpectedChanges              int                 `json:"expected_changes"`
	DetectedChanges              int                 `json:"detected_changes"`
	DetectionRate                float64             `json:"detection_rate"`
	DetectionInterval            *ProportionInterval `json:"detection_interval,omitempty"`
	FalseAlarms                  int                 `json:"false_alarms"`
	TotalTriggers                int                 `json:"total_triggers"`
	UnmatchedAlarmsPerTrajectory float64             `json:"unmatched_alarms_per_trajectory"`
	UnmatchedTriggerRate         float64             `json:"unmatched_trigger_rate"`
	DelaySampleCount             int                 `json:"delay_sample_count"`
	MeanDelay                    float64             `json:"mean_delay"`
	MeanDelayBasis               string              `json:"mean_delay_basis"`
	MaxDelay                     int                 `json:"max_delay"`
	MissRate                     float64             `json:"miss_rate"`
	Acceptance                   AcceptanceResult    `json:"acceptance"`
}

type GroupComparisonReport struct {
	Trajectories     int                            `json:"trajectories"`
	SamplesPerMember int                            `json:"samples_per_member"`
	SeedBase         int64                          `json:"seed_base"`
	Scenarios        map[string]GroupComparisonCase `json:"scenarios"`
}

type GroupComparisonCase struct {
	Expected                 string             `json:"expected"`
	ExpectedDecisionCount    int                `json:"expected_decision_count"`
	ExpectedDecisionRate     float64            `json:"expected_decision_rate"`
	ExpectedDecisionInterval ProportionInterval `json:"expected_decision_interval"`
	ShareRate                float64            `json:"share_rate"`
	SplitRate                float64            `json:"split_rate"`
	UncertainRate            float64            `json:"uncertain_rate"`
	WrongCount               int                `json:"wrong_count"`
	WrongRate                float64            `json:"wrong_rate"`
	WrongRateInterval        ProportionInterval `json:"wrong_rate_interval"`
	Acceptance               AcceptanceResult   `json:"acceptance"`
}

type IntegrationControlsReport struct {
	GroupAuthority GroupAuthorityControl `json:"group_authority"`
}

type GroupAuthorityControl struct {
	Passed                          bool   `json:"passed"`
	DivergentRecommendation         string `json:"divergent_recommendation"`
	CompatibleRecommendation        string `json:"compatible_recommendation"`
	RequiresAntiPigeonCertification bool   `json:"requires_anti_pigeon_certification"`
	SnapshotUnchanged               bool   `json:"snapshot_unchanged"`
	PosteriorKeyAuthorityUnchanged  bool   `json:"posterior_key_authority_unchanged"`
}
