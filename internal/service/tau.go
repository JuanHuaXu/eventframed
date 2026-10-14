package service

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/agency"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

// PlanTauTurn turns a trusted producer's law into a durable proposal. It never
// grants execution authority: the OpenClaw adapter still owns that decision.
func (s *Service) PlanTauTurn(ctx context.Context, request model.PlanTauTurnRequest) (model.PlanTauTurnResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.PlanTauTurnResponse{}, err
	}
	if !s.config.AgencyPolicy.Enabled || s.config.AgencySigner == nil || subtle.ConstantTimeCompare([]byte(request.IssuerToken), []byte(s.config.AgencyIssuerToken)) != 1 {
		return model.PlanTauTurnResponse{}, errors.New("tau planner requires enabled agency and issuer authentication")
	}
	if !boundedAgencyIdentifier(request.PlanID) || !boundedAgencyIdentifier(request.TenantID) || request.SessionID == "" || len(request.SessionID) > 1024 || !boundedAgencyIdentifier(request.OriginEventID) {
		return model.PlanTauTurnResponse{}, errors.New("bounded tau plan, tenant, session and origin required")
	}
	now := time.Now().UTC()
	base := request.Base
	identity, _ := json.Marshal([]string{request.TenantID, request.PlanID})
	idHash := sha256.Sum256(identity)
	id := "tau-" + hex.EncodeToString(idHash[:16])
	digestRequest := request
	digestRequest.IssuerToken = ""
	requestDigest := tauDigest(digestRequest)
	reader, ok := s.store.(store.AgencyProposalReader)
	if !ok {
		return model.PlanTauTurnResponse{}, errors.New("store lacks durable agency decision reads")
	}
	if existing, err := reader.GetAgencyProposal(ctx, request.TenantID, id); err == nil {
		if existing.Proposal.Tau == nil || existing.Proposal.Tau.RequestDigest != requestDigest {
			return model.PlanTauTurnResponse{}, store.ErrAgencyConflict
		}
		binding := existing.Proposal.Tau
		issued := model.IssueAgencyProposalResponse{ProtocolVersion: model.ProtocolVersion, Duplicate: true, Record: existing, Snapshot: s.store.Snapshot(ctx)}
		return model.PlanTauTurnResponse{ProtocolVersion: model.ProtocolVersion, Forecast: binding.Forecast, Decision: binding.Decision, Proposal: &issued, GuaranteeScope: "replayed original durable decision; current proposal status governs delivery"}, nil
	} else if !errors.Is(err, store.ErrAgencyNotFound) {
		return model.PlanTauTurnResponse{}, err
	}
	if request.Forecast.Base != base {
		return model.PlanTauTurnResponse{}, errors.New("tau forecast must bind its original snapshot")
	}
	if !store.TauSnapshotCompatible(&model.TauTurnBinding{Base: base}, s.store.Snapshot(ctx)) {
		return model.PlanTauTurnResponse{}, store.ErrStaleSnapshot
	}
	ids := append([]string(nil), request.EvidenceIDs...)
	if len(request.Episodes) > 0 {
		if len(ids) > 0 || len(request.Episodes) > 15 {
			return model.PlanTauTurnResponse{}, errors.New("fit mode derives its bounded evidence set from episodes")
		}
		ids = append(ids, request.OriginEventID)
		for _, ep := range request.Episodes {
			ids = append(ids, ep.AnchorID, ep.OutcomeID)
		}
	}
	if len(ids) < 1 || len(ids) > s.config.AgencyPolicy.MaxEvidence {
		return model.PlanTauTurnResponse{}, errors.New("bounded tau evidence set required")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !boundedAgencyIdentifier(id) || seen[id] {
			return model.PlanTauTurnResponse{}, errors.New("tau evidence identities must be distinct")
		}
		seen[id] = true
	}
	if !seen[request.OriginEventID] {
		return model.PlanTauTurnResponse{}, errors.New("tau evidence must include origin")
	}
	events, err := s.store.GetEvents(ctx, request.TenantID, ids, now)
	if err != nil {
		return model.PlanTauTurnResponse{}, err
	}
	byID := map[string]model.Event{}
	for _, e := range events {
		if e.TenantID == request.TenantID {
			byID[e.ID] = e
		}
	}
	if len(byID) != len(ids) {
		return model.PlanTauTurnResponse{}, store.ErrAgencyEvidence
	}
	origin := byID[request.OriginEventID]
	if origin.SessionID != request.SessionID {
		return model.PlanTauTurnResponse{}, errors.New("tau origin session differs from destination")
	}
	f := request.Forecast
	if len(request.Episodes) > 0 {
		if f.IssuedAt.IsZero() || f.IssuedAt.After(now) {
			return model.PlanTauTurnResponse{}, errors.New("fit mode requires a fixed as-of issue time")
		}
		f, err = agency.FitTau(f.ID, origin.OccurredAt, f.IssuedAt, f.HorizonMS, request.BinDelaysMS, request.Episodes, byID, request.Policy.TargetKind)
		if err != nil {
			return model.PlanTauTurnResponse{}, err
		}
		// This is a working law, not an empirical certificate of its accuracy.
		f.TVRadius = request.Forecast.TVRadius
		f.Base = base
	}
	if !f.Origin.Equal(origin.OccurredAt) || f.IssuedAt.Before(origin.AvailableAt) {
		return model.PlanTauTurnResponse{}, errors.New("tau law must bind the available origin frame")
	}
	for _, event := range events {
		if event.AvailableAt.After(f.IssuedAt) {
			return model.PlanTauTurnResponse{}, errors.New("tau evidence became available after forecast issue")
		}
	}
	decision, err := agency.PlanTau(f, request.Policy, now)
	if err != nil {
		return model.PlanTauTurnResponse{}, err
	}
	response := model.PlanTauTurnResponse{ProtocolVersion: model.ProtocolVersion, Forecast: f, Decision: decision, GuaranteeScope: "conditional on declared forecast-error and utility contracts"}
	if !decision.Schedule {
		return response, nil
	}
	draft := model.AgencyProposalDraft{ID: id, TenantID: request.TenantID, SessionID: request.SessionID, Action: model.AgencySchedule, Reason: request.Reason, EvidenceIDs: ids, ExpectedUtility: decision.LowerUtility, Priority: request.Priority, NotBefore: decision.DeliverAt, ScheduledFor: &decision.DeliverAt, ExpiresAt: decision.LatestDispatch, IdempotencyKey: id, CausalChainID: id}
	proposal, err := agency.BuildProposal(draft, now, s.config.AgencyPolicy)
	if err != nil {
		return model.PlanTauTurnResponse{}, err
	}
	proposal.Tau = &model.TauTurnBinding{RequestDigest: requestDigest, Forecast: f, Decision: decision, PlanID: request.PlanID, TargetKind: request.Policy.TargetKind, ForecastID: f.ID, ForecastDigest: tauDigest(f), PolicyDigest: tauDigest(request.Policy), Base: base, LatestDispatch: decision.LatestDispatch, LowerProbability: decision.LowerProbability, LowerUtility: decision.LowerUtility, CooldownMS: request.Policy.CooldownMS}
	issued, err := s.persistAgencyProposal(ctx, proposal, now)
	if err != nil {
		return model.PlanTauTurnResponse{}, err
	}
	response.Proposal = &issued
	return response, nil
}

func tauDigest(v any) string {
	data, _ := json.Marshal(v)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *Service) GetAgencyProposal(ctx context.Context, request model.GetAgencyProposalRequest) (model.IssueAgencyProposalResponse, error) {
	if err := checkProtocol(request.ProtocolVersion); err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	if !s.config.AgencyPolicy.Enabled || subtle.ConstantTimeCompare([]byte(request.AuthorityToken), []byte(s.config.AgencyAuthorityToken)) != 1 || !boundedAgencyIdentifier(request.TenantID) || !boundedAgencyIdentifier(request.ProposalID) {
		return model.IssueAgencyProposalResponse{}, errors.New("agency authority authentication and bounded identity required")
	}
	reader, ok := s.store.(store.AgencyProposalReader)
	if !ok {
		return model.IssueAgencyProposalResponse{}, errors.New("store lacks agency decision reads")
	}
	r, err := reader.GetAgencyProposal(ctx, request.TenantID, request.ProposalID)
	if err != nil {
		return model.IssueAgencyProposalResponse{}, err
	}
	return model.IssueAgencyProposalResponse{ProtocolVersion: model.ProtocolVersion, Record: r, Snapshot: s.store.Snapshot(ctx)}, nil
}
