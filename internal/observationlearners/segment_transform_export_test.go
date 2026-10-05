package observationlearners

import "context"

// Exported only into the Go test binary for the external service integration.
// Fixture labels are workload, not labels inferred from retrieval diagnostics.
func ResearchTransformFixture(ctx context.Context) (float64, error) {
	m, e := fitSegmentPosteriorContext(ctx, -16, 256, transformHistory(272), 64, .01, .95)
	if e != nil {
		return 0, e
	}
	return m.predictions[17], nil
}
