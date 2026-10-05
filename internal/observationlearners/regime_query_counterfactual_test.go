package observationlearners

type regimeCounterfactualRecord struct {
	Original, Flipped regimeEnvelopeRecord
	Error             string `json:",omitempty"`
}

// Only labels absent from natural publication evidence may be flipped together.
// Separate branch fitting then admits precisely one counterfactual answer.
func regimeCounterfactualInput(input softV120Record) softV120Record {
	out := input
	out.Steps = append([]softV120Step(nil), input.Steps...)
	if len(out.Steps) != 256 {
		return out
	}
	for j := 152; j < 160; j++ {
		s := &out.Steps[j]
		if s.Missing || j+s.Delay > 161 {
			s.Y = !s.Y
		}
	}
	return out
}

func runRegimeCounterfactual(input softV120Record) regimeCounterfactualRecord {
	r := regimeCounterfactualRecord{Original: runRegimeEnvelope(input)}
	if r.Original.Error != "" {
		r.Error = r.Original.Error
		return r
	}
	r.Flipped = runRegimeEnvelope(regimeCounterfactualInput(input))
	if r.Flipped.Error != "" {
		r.Error = r.Flipped.Error
	}
	return r
}
