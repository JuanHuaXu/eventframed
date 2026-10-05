package observationlearners

import (
	"encoding/json"
	"io"
	"os"
	"testing"
)

// Unlike the poisoning fixture, this leaves every other label exactly as stored.
// It independently proves the batched flip cannot combine purchased answers.
func TestRegimeCounterfactualSingleLabelIsolation(t *testing.T) {
	path := os.Getenv("EVENTFRAME_ACQUISITION_TRAIN_INPUT")
	if path == "" {
		t.Skip("explicit source")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	d := json.NewDecoder(f)
	var header softV120Artifact
	if err := d.Decode(&header); err != nil {
		t.Fatal(err)
	}
	checked, branches := 0, 0
	for {
		var input softV120Record
		err := d.Decode(&input)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if input.Phase != 0 || input.Index != 0 || (input.Case != 0 && input.Case != 5 && input.Case != 19 && input.Case != 20) {
			continue
		}
		flipped := runRegimeEnvelope(regimeCounterfactualInput(input))
		if flipped.Error != "" {
			t.Fatal(flipped.Error)
		}
		for i, j := range flipped.Choices {
			one := input
			one.Steps = append([]softV120Step(nil), input.Steps...)
			if j >= 0 && (one.Steps[j].Missing || j+one.Steps[j].Delay > 161) {
				one.Steps[j].Y = !one.Steps[j].Y
			}
			got := publishRegimeOutcome(one, regimeOutcomeDecision{Clock: 160, Selected: [4]int{j, j, j, j}})
			want := flipped.Bundles[i/4]
			a := i % 4
			if got.Error != "" || got.Predictions[0] != want.Predictions[a] || got.AtPublication[0] != want.AtPublication[a] || got.LogEvidence[0] != want.LogEvidence[a] {
				t.Fatal("joint flipping changes isolated branch")
			}
			branches++
		}
		checked++
	}
	if checked != 8 || branches <= 8 {
		t.Fatal("fixture count", checked, branches)
	}
	t.Logf("eight fixtures, %d branches: exact single-label versus grouped-flip equality", branches)
}
