package researchmemory

import "context"

// OpenSourceOwnerResolvedAdmissions additionally reuses validated, call-local
// source resolutions during admission. Only the private owner obtains and uses
// them under its lock. There is no caller-provided resolution or lifetime cache.
// Exact atomic append conflicts, uncertainty stops, actual worker forecasts and
// the external service admission-guard requirement are unchanged. Other source
// constructors retain their original preflight reads as experimental controls.
func OpenSourceOwnerResolvedAdmissions(ctx context.Context, path, tenant, stream string, epoch uint64, seed int64) (*SourceOwner, error) {
	o, err := OpenSourceOwnerBatchReads(ctx, path, tenant, stream, epoch, seed)
	if err != nil {
		return nil, err
	}
	o.reuseAdmissions = true
	return o, nil
}
