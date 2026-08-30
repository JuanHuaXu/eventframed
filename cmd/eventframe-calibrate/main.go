package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/calibration"
	"github.com/JuanHuaXu/eventframed/internal/evaluation"
	"github.com/JuanHuaXu/eventframed/internal/productioneval"
)

type artifact struct {
	SchemaVersion string            `json:"schema_version"`
	FitAt         time.Time         `json:"fit_at"`
	DatasetSHA256 string            `json:"dataset_sha256"`
	Cases         int               `json:"cases"`
	Pairs         int               `json:"pairs"`
	Variant       string            `json:"variant"`
	BeforeBrier   float64           `json:"before_brier"`
	AfterBrier    float64           `json:"after_brier"`
	Selection     string            `json:"selection"`
	Calibration   calibration.Logit `json:"calibration"`
	InputMap      calibration.Logit `json:"input_map"`
	DeploymentMap calibration.Logit `json:"deployment_map"`
	Confirmation  *validation       `json:"confirmation,omitempty"`
}

type validation struct {
	DatasetSHA256 string                `json:"dataset_sha256"`
	Cases         int                   `json:"cases"`
	Before        evaluation.Metrics    `json:"before"`
	After         evaluation.Metrics    `json:"after"`
	Comparison    evaluation.Comparison `json:"after_vs_baseline"`
}

func main() {
	input := flag.String("input", "", "design dataset JSON")
	output := flag.String("output", "", "calibration artifact JSON")
	variant := flag.String("variant", "baseline", "forecast variant to calibrate")
	confirmation := flag.String("confirmation", "", "optional untouched confirmation dataset JSON")
	inputScale := flag.Float64("input-scale", 1, "scale of the calibration map already applied to the dataset")
	inputBias := flag.Float64("input-bias", 0, "bias of the calibration map already applied to the dataset")
	inputFloor := flag.Float64("input-floor", 1e-6, "floor of the calibration map already applied to the dataset")
	flag.Parse()
	if *input == "" || *output == "" {
		fail(fmt.Errorf("input and output are required"))
	}
	inputMap := calibration.Logit{Scale: *inputScale, Bias: *inputBias, Floor: *inputFloor}
	if !inputMap.Valid() {
		fail(fmt.Errorf("input calibration map is invalid"))
	}
	payload, err := os.ReadFile(*input)
	if err != nil {
		fail(err)
	}
	var dataset productioneval.Artifact
	if err := json.Unmarshal(payload, &dataset); err != nil {
		fail(err)
	}
	emittedObservations, err := extract(dataset, *variant)
	if err != nil {
		fail(err)
	}
	fitObservations := invertObservations(emittedObservations, inputMap)
	fitted, err := calibration.Fit(fitObservations)
	if err != nil {
		fail(err)
	}
	beforeBrier := calibration.Brier(emittedObservations, calibration.Identity())
	afterBrier := calibration.Brier(fitObservations, fitted)
	selection := "fitted_logit"
	if afterBrier > beforeBrier {
		fitted = inputMap
		afterBrier = beforeBrier
		selection = "input_map_noninferiority_fallback"
	}
	digest := sha256.Sum256(payload)
	result := artifact{
		SchemaVersion: "eventframe.calibration.v1", FitAt: time.Now().UTC(), DatasetSHA256: hex.EncodeToString(digest[:]),
		Cases: len(dataset.Cases), Pairs: len(emittedObservations), Variant: *variant,
		BeforeBrier: beforeBrier, AfterBrier: afterBrier, Selection: selection, Calibration: fitted,
		InputMap: inputMap, DeploymentMap: fitted,
	}
	if *confirmation != "" {
		validated, err := validateConfirmation(*confirmation, *variant, inputMap, fitted)
		if err != nil {
			fail(err)
		}
		result.Confirmation = &validated
	}
	if err := productioneval.WriteJSON(*output, result); err != nil {
		fail(err)
	}
	encoded, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(encoded))
}

func validateConfirmation(path, variant string, inputMap, fitted calibration.Logit) (validation, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return validation{}, err
	}
	var dataset productioneval.Artifact
	if err := json.Unmarshal(payload, &dataset); err != nil {
		return validation{}, err
	}
	before, err := productioneval.EvaluateArtifactWithBaseline(dataset, "baseline")
	if err != nil {
		return validation{}, err
	}
	for caseIndex := range dataset.Cases {
		forecast, ok := dataset.Cases[caseIndex].Variants[variant]
		if !ok {
			return validation{}, fmt.Errorf("case %s lacks variant %q", dataset.Cases[caseIndex].ID, variant)
		}
		for candidateIndex := range forecast.Candidates {
			if forecast.Candidates[candidateIndex].Nominated {
				raw := inputMap.Inverse(forecast.Candidates[candidateIndex].Probability)
				forecast.Candidates[candidateIndex].Probability = fitted.Apply(raw)
			}
		}
		dataset.Cases[caseIndex].Variants[variant] = forecast
	}
	after, err := productioneval.EvaluateArtifactWithBaseline(dataset, "baseline")
	if err != nil {
		return validation{}, err
	}
	digest := sha256.Sum256(payload)
	return validation{
		DatasetSHA256: hex.EncodeToString(digest[:]), Cases: len(dataset.Cases),
		Before: before.Variants[variant], After: after.Variants[variant], Comparison: after.Comparisons[variant],
	}, nil
}

func invertObservations(observations []calibration.Observation, inputMap calibration.Logit) []calibration.Observation {
	inverted := make([]calibration.Observation, len(observations))
	for index, observation := range observations {
		inverted[index] = observation
		inverted[index].Probability = inputMap.Inverse(observation.Probability)
	}
	return inverted
}

func extract(dataset productioneval.Artifact, variant string) ([]calibration.Observation, error) {
	var observations []calibration.Observation
	for _, item := range dataset.Cases {
		forecast, ok := item.Variants[variant]
		if !ok {
			return nil, fmt.Errorf("case %s lacks variant %q", item.ID, variant)
		}
		relevant := make(map[string]struct{}, len(item.RelevantEventIDs))
		for _, id := range item.RelevantEventIDs {
			relevant[id] = struct{}{}
		}
		for _, candidate := range forecast.Candidates {
			if !candidate.Nominated {
				continue
			}
			outcome := 0.0
			if _, ok := relevant[candidate.EventID]; ok {
				outcome = 1
			}
			observations = append(observations, calibration.Observation{Probability: candidate.Probability, Outcome: outcome, Weight: 1 / float64(len(forecast.Candidates))})
		}
	}
	return observations, nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eventframe-calibrate:", err)
	os.Exit(1)
}
