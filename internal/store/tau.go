package store

import (
	"context"
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"time"
)

// Reporting uncertainty grants no execution permission and must survive lease
// and proposal expiry. Keep every known handle so reconciliation is explicit.
func AgencyUncertainRecord(record model.AgencyProposalRecord, request model.ResolveAgencyProposalRequest, now time.Time) (model.AgencyProposalRecord, error) {
	refs := append([]string(nil), record.UncertainExecutionRefs...)
	for _, ref := range []string{record.ExecutionRef, request.ExecutionRef} {
		if ref == "" {
			continue
		}
		found := false
		for _, old := range refs {
			if old == ref {
				found = true
				break
			}
		}
		if !found {
			if len(refs) >= 32 {
				return record, errors.New("uncertain execution handle bound reached")
			}
			refs = append(refs, ref)
		}
	}
	record.Status, record.ResolutionReason, record.ExecutionRef, record.ResolvedAt = model.AgencyUncertain, request.Reason, request.ExecutionRef, now
	record.ClaimedBy, record.LeaseUntil = "", time.Time{}
	record.UncertainExecutionRefs = refs
	return record, nil
}

type AgencyProposalReader interface {
	GetAgencyProposal(context.Context, string, string) (model.AgencyProposalRecord, error)
}

type TauScope struct{ TenantID, SessionID, TargetKind string }

// A cancellation acknowledgment must cover the current complete handle set.
// Comparing under the store lock fences stale replies and newly reported work.
func TauCancellationMatches(record model.AgencyProposalRecord, request model.ResolveAgencyProposalRequest) bool {
	if len(record.UncertainExecutionRefs) == 0 || len(request.CancelledExecutionRefs) != len(record.UncertainExecutionRefs) || len(request.CancelledExecutionRefs) > 32 {
		return false
	}
	seen := map[string]bool{}
	for _, ref := range request.CancelledExecutionRefs {
		if seen[ref] {
			return false
		}
		seen[ref] = true
	}
	for _, ref := range record.UncertainExecutionRefs {
		if !seen[ref] {
			return false
		}
	}
	return true
}

func TauScopeOf(p model.AgencyProposal) TauScope {
	if p.Tau == nil {
		return TauScope{}
	}
	return TauScope{p.TenantID, p.SessionID, p.Tau.TargetKind}
}

// Agency bookkeeping advances both counters. Every other mutation advances
// RuntimeVersion without AgencyVersion and invalidates a frozen tau forecast.
func TauSnapshotCompatible(tau *model.TauTurnBinding, current model.Snapshot) bool {
	if tau == nil {
		return true
	}
	b := tau.Base
	return current.RuntimeVersion >= b.RuntimeVersion && current.AgencyVersion >= b.AgencyVersion && current.RuntimeVersion-b.RuntimeVersion == current.AgencyVersion-b.AgencyVersion && b.PolicyVersion == current.PolicyVersion && b.ContractVersion == current.ContractVersion && b.GraphVersion == current.GraphVersion && b.PosteriorVersion == current.PosteriorVersion && b.ResidualVersion == current.ResidualVersion && b.AbstractionVersion == current.AbstractionVersion && b.EvidenceEpoch == current.EvidenceEpoch
}

// Only one live opportunity per session/target may reserve a turn. An approved
// opportunity also occupies a cooldown, so new plan IDs cannot bypass it.
func TauOpportunityBlocked(incoming, existing model.AgencyProposalRecord, current model.Snapshot, now time.Time) bool {
	a, b := incoming.Proposal, existing.Proposal
	if a.Tau == nil || b.Tau == nil || a.TenantID != b.TenantID || a.SessionID != b.SessionID || a.Tau.TargetKind != b.Tau.TargetKind {
		return false
	}
	if (existing.Status == model.AgencyPending || existing.Status == model.AgencyClaimed) && b.ExpiresAt.After(now) && TauSnapshotCompatible(b.Tau, current) {
		return true
	}
	if existing.Status == model.AgencyUncertain {
		return true
	}
	if existing.Status != model.AgencyApproved && existing.Status != model.AgencyUncertain {
		return false
	}
	cooldown := max(a.Tau.CooldownMS, b.Tau.CooldownMS)
	return now.Before(existing.ResolvedAt.Add(time.Duration(cooldown) * time.Millisecond))
}
