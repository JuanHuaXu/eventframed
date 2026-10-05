// Package researchvalidity contains an unwired, conditional posterior reuse
// guard. It does not certify the provenance or score bounds supplied to it.
package researchvalidity

import (
	"math"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
)

type Status uint8

const (
	Unknown Status = iota
	Blocked
	Compatible
)

type Result struct {
	Status Status
	Reason string
}

type Record struct {
	TenantID       string
	PosteriorKey   string
	QueryDigest    string
	ModelID        string
	HorizonKey     string
	SourceID       string
	SourceAt       time.Time
	FrontierDigest string
	FrontierK      int
	CutoffLower    float64
	Base           model.Snapshot
}

type Request struct {
	TenantID       string
	PosteriorKey   string
	QueryDigest    string
	FrontierDigest string
	ModelID        string
	HorizonKey     string
	SourceID       string
	AsOf           time.Time
	Current        model.Snapshot
}

type MutationKind uint8

const (
	Insert MutationKind = iota + 1
	Outcome
	Certificate
	Other
)

type Mutation struct {
	RuntimeVersion     uint64
	EvidenceEpochAfter uint64
	Kind               MutationKind
	AvailableAt        time.Time
	ScoreUpper         float64
	ScoreBounded       bool
	GraphChecked       bool
	GraphTouches       bool
	AffectedKnown      bool
	AffectedKeys       []string
}

// Check proves only a fixed-query, fixed-frontier posterior reuse condition.
// The caller must independently verify the witness, source, and certificates.
func Check(record Record, request Request, mutations []Mutation) Result {
	unknown := func(reason string) Result { return Result{Unknown, reason} }
	blocked := func(reason string) Result { return Result{Blocked, reason} }
	if record.TenantID == "" || record.PosteriorKey == "" || record.QueryDigest == "" ||
		record.ModelID == "" || record.HorizonKey == "" || record.SourceID == "" ||
		record.FrontierDigest == "" || record.FrontierK < 1 || record.SourceAt.IsZero() ||
		math.IsNaN(record.CutoffLower) || math.IsInf(record.CutoffLower, 0) ||
		record.CutoffLower < -1 || record.CutoffLower > 1 || request.AsOf.IsZero() {
		return unknown("missing or invalid binding")
	}
	if record.TenantID != request.TenantID || record.PosteriorKey != request.PosteriorKey ||
		record.QueryDigest != request.QueryDigest || record.ModelID != request.ModelID ||
		record.HorizonKey != request.HorizonKey || record.SourceID != request.SourceID ||
		record.FrontierDigest != request.FrontierDigest {
		return blocked("posterior identity or model changed")
	}
	if record.SourceAt.After(request.AsOf) {
		return blocked("posterior source is unavailable as of request")
	}
	base, current := record.Base, request.Current
	if base.PolicyVersion != current.PolicyVersion || base.ContractVersion != current.ContractVersion ||
		base.GraphVersion != current.GraphVersion || base.AbstractionVersion != current.AbstractionVersion ||
		base.AgencyVersion != current.AgencyVersion || current.RuntimeVersion < base.RuntimeVersion ||
		current.EvidenceEpoch < base.EvidenceEpoch {
		return blocked("semantic versions moved or runtime regressed")
	}
	if len(mutations) > 10000 {
		return unknown("mutation witness exceeds declared bound")
	}
	version := base.RuntimeVersion
	epoch := base.EvidenceEpoch
	for _, mutation := range mutations {
		if version == math.MaxUint64 || mutation.RuntimeVersion != version+1 {
			return unknown("mutation witness has a gap or is unordered")
		}
		version++
		nextEpoch := epoch
		if mutation.Kind == Insert {
			if epoch == math.MaxUint64 {
				return unknown("evidence epoch overflow")
			}
			nextEpoch++
		}
		if mutation.EvidenceEpochAfter != nextEpoch {
			return unknown("mutation witness does not account for evidence epoch motion")
		}
		epoch = nextEpoch
		switch mutation.Kind {
		case Insert:
			if mutation.AvailableAt.IsZero() {
				return unknown("insert availability is missing")
			}
			if mutation.AvailableAt.After(request.AsOf) {
				continue
			}
			if !mutation.ScoreBounded || !mutation.GraphChecked ||
				math.IsNaN(mutation.ScoreUpper) || math.IsInf(mutation.ScoreUpper, 0) ||
				mutation.ScoreUpper < -1 || mutation.ScoreUpper > 1 {
				return unknown("visible insert lacks a complete score/dependency bound")
			}
			if mutation.GraphTouches || mutation.ScoreUpper >= record.CutoffLower {
				return blocked("visible insert can change the frontier or dependency closure")
			}
		case Outcome:
			if mutation.AvailableAt.IsZero() {
				return unknown("outcome availability is missing")
			}
			if mutation.AvailableAt.After(request.AsOf) {
				continue
			}
			if !mutation.AffectedKnown {
				return unknown("outcome dependency set is incomplete")
			}
			for _, key := range mutation.AffectedKeys {
				if key == record.PosteriorKey {
					return blocked("posterior dependency received new evidence")
				}
			}
		case Certificate:
			// Certificate validity is checked separately at the served-law gate.
		default:
			return unknown("unclassified runtime mutation")
		}
	}
	if version != current.RuntimeVersion || epoch != current.EvidenceEpoch {
		return unknown("mutation witness does not reach the current snapshot")
	}
	return Result{Compatible, "complete non-impact witness"}
}
