package researchmemory

// admissionPhases is test-enabled only, installed before exclusive owner use.
// It records successful-path elapsed time, not a certificate that preflight is
// movable: source resolution and validation still rely on owner serialization.
type admissionPhases struct {
	SourceNS, PreflightNS, StageNS, AppendNS int64
}
