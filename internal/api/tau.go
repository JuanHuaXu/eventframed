package api

import (
	"errors"
	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/store"
	"net/http"
)

func (s *Server) planTauTurn(writer http.ResponseWriter, request *http.Request) {
	var input model.PlanTauTurnRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	result, err := s.service.PlanTauTurn(request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrStaleSnapshot) || errors.Is(err, store.ErrAgencyConflict) {
			status = http.StatusConflict
		}
		if errors.Is(err, store.ErrAgencyChainBudget) {
			status = http.StatusTooManyRequests
		}
		writeError(writer, status, "tau_plan_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) getAgencyProposal(writer http.ResponseWriter, request *http.Request) {
	var input model.GetAgencyProposalRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	r, err := s.service.GetAgencyProposal(request.Context(), input)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, store.ErrAgencyNotFound) {
			status = http.StatusNotFound
		}
		writeError(writer, status, "agency_read_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, r)
}
