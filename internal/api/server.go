package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
	"github.com/JuanHuaXu/eventframed/internal/retrieval"
	"github.com/JuanHuaXu/eventframed/internal/service"
	"github.com/JuanHuaXu/eventframed/internal/store"
)

const maxRequestBytes = 4 << 20

type Server struct {
	service *service.Service
	logger  *slog.Logger
	mux     *http.ServeMux
	metrics *runtimeMetrics
}

func NewServer(runtime *service.Service, logger *slog.Logger) *Server {
	server := &Server{service: runtime, logger: logger, mux: http.NewServeMux(), metrics: newRuntimeMetrics()}
	server.mux.HandleFunc("GET /v1/health", server.health)
	server.mux.HandleFunc("GET /v1/ready", server.ready)
	server.mux.HandleFunc("POST /v1/turns:capture", server.captureTurn)
	server.mux.HandleFunc("POST /v1/events:observe", server.observe)
	server.mux.HandleFunc("POST /v1/context:recall", server.recall)
	server.mux.HandleFunc("POST /v1/openclaw/context:recall", server.recallOpenClaw)
	server.mux.HandleFunc("POST /v1/events:delete", server.delete)
	server.mux.HandleFunc("POST /v1/maintenance:retain", server.retain)
	server.mux.HandleFunc("POST /v1/maintenance:backup", server.backup)
	server.mux.HandleFunc("POST /v1/maintenance:compact", server.compact)
	server.mux.HandleFunc("POST /v1/bayesian/certificates:publish-selection", server.publishSelectionCertificate)
	server.mux.HandleFunc("POST /v1/bayesian/certificates:publish-anti-pigeon", server.publishAntiPigeonCertificate)
	server.mux.HandleFunc("POST /v1/bayesian/certificates:publish-omitted-influence", server.publishOmittedInfluenceCertificate)
	server.mux.HandleFunc("POST /v1/bayesian/certificates:estimate-omitted-influence", server.estimateOmittedInfluence)
	server.mux.HandleFunc("POST /v1/bayesian/outcomes:observe", server.observeBayesianOutcome)
	server.mux.HandleFunc("POST /v1/bayesian/groups:compare", server.compareBayesianGroup)
	server.mux.HandleFunc("GET /v1/abstraction/graph", server.getPredictiveGraph)
	server.mux.HandleFunc("POST /v1/abstraction/snaps:publish", server.publishPredictiveSnap)
	server.mux.HandleFunc("POST /v1/abstraction/snaps:rollback", server.rollbackPredictiveSnap)
	server.mux.HandleFunc("POST /v1/agency/proposals:issue", server.issueAgencyProposal)
	server.mux.HandleFunc("POST /v1/agency/proposals:claim", server.claimAgencyProposals)
	server.mux.HandleFunc("POST /v1/agency/proposals:resolve", server.resolveAgencyProposal)
	server.mux.HandleFunc("GET /metrics", server.metrics.handle)
	return server
}

func (s *Server) issueAgencyProposal(writer http.ResponseWriter, request *http.Request) {
	var input model.IssueAgencyProposalRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.IssueAgencyProposal(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrAgencyConflict) {
			writeError(writer, http.StatusConflict, "idempotency_conflict", err)
			return
		}
		if errors.Is(err, store.ErrAgencyChainBudget) {
			writeError(writer, http.StatusTooManyRequests, "causal_chain_budget", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "proposal_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) claimAgencyProposals(writer http.ResponseWriter, request *http.Request) {
	var input model.ClaimAgencyProposalsRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.ClaimAgencyProposals(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "claim_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) resolveAgencyProposal(writer http.ResponseWriter, request *http.Request) {
	var input model.ResolveAgencyProposalRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.ResolveAgencyProposal(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrAgencyConflict) || errors.Is(err, store.ErrAgencyLease) || errors.Is(err, store.ErrAgencyExpired) {
			writeError(writer, http.StatusConflict, "agency_conflict", err)
			return
		}
		if errors.Is(err, store.ErrAgencyNotFound) {
			writeError(writer, http.StatusNotFound, "proposal_not_found", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "resolution_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) getPredictiveGraph(writer http.ResponseWriter, request *http.Request) {
	response, err := s.service.GetPredictiveGraph(request.Context(), request.URL.Query().Get("tenant_id"))
	if err != nil {
		if errors.Is(err, store.ErrStaleSnapshot) {
			writeError(writer, http.StatusConflict, "snapshot_changed", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) publishPredictiveSnap(writer http.ResponseWriter, request *http.Request) {
	var input model.PredictiveSnapRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.PublishPredictiveSnap(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrStaleSnapshot) || errors.Is(err, store.ErrSnapConflict) {
			writeError(writer, http.StatusConflict, "snapshot_changed", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "snap_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) rollbackPredictiveSnap(writer http.ResponseWriter, request *http.Request) {
	var input model.RollbackSnapRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.RollbackPredictiveSnap(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrSnapConflict) {
			writeError(writer, http.StatusConflict, "snap_conflict", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "rollback_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) publishOmittedInfluenceCertificate(writer http.ResponseWriter, request *http.Request) {
	var input model.PublishOmittedInfluenceCertificateRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.PublishOmittedInfluenceCertificate(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "certificate_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) estimateOmittedInfluence(writer http.ResponseWriter, request *http.Request) {
	var input model.EstimateOmittedInfluenceRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.EstimateOmittedInfluence(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "audit_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) observeBayesianOutcome(writer http.ResponseWriter, request *http.Request) {
	var input model.BayesianOutcomeRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.ObserveBayesianOutcome(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrOutcomeConflict) {
			writeError(writer, http.StatusConflict, "idempotency_conflict", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "outcome_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) compareBayesianGroup(writer http.ResponseWriter, request *http.Request) {
	var input model.BayesianGroupComparisonRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.CompareBayesianGroup(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrStaleSnapshot) {
			writeError(writer, http.StatusConflict, "snapshot_changed", err)
			return
		}
		writeError(writer, http.StatusBadRequest, "comparison_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) publishSelectionCertificate(writer http.ResponseWriter, request *http.Request) {
	var input model.PublishSelectionCertificateRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.PublishSelectionCertificate(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "certificate_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) publishAntiPigeonCertificate(writer http.ResponseWriter, request *http.Request) {
	var input model.PublishAntiPigeonCertificateRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.PublishAntiPigeonCertificate(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "certificate_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) delete(writer http.ResponseWriter, request *http.Request) {
	var input model.DeleteRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Delete(request.Context(), input)
	if err != nil {
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "delete_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) retain(writer http.ResponseWriter, request *http.Request) {
	var input model.RetentionRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Retain(request.Context(), input)
	if err != nil {
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "retention_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) backup(writer http.ResponseWriter, request *http.Request) {
	var input model.BackupRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Backup(request.Context(), input)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "backup_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) compact(writer http.ResponseWriter, request *http.Request) {
	var input struct {
		ProtocolVersion string `json:"protocol_version"`
	}
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	if input.ProtocolVersion != model.ProtocolVersion {
		writeError(writer, http.StatusBadRequest, "compact_rejected", errors.New("unsupported protocol_version"))
		return
	}
	response, err := s.service.Compact(request.Context())
	if err != nil {
		writeError(writer, http.StatusBadRequest, "compact_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) Handler() http.Handler {
	return requestLog(s.logger, s.metrics, s.mux)
}

func (s *Server) health(writer http.ResponseWriter, request *http.Request) {
	response, err := s.service.Health(request.Context())
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, "store_unavailable", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) ready(writer http.ResponseWriter, request *http.Request) {
	response, err := s.service.Ready(request.Context())
	if err != nil {
		writeJSON(writer, http.StatusServiceUnavailable, response)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) observe(writer http.ResponseWriter, request *http.Request) {
	var input model.ObserveRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Observe(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrIdempotencyConflict) {
			writeError(writer, http.StatusConflict, "idempotency_conflict", err)
			return
		}
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "observe_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) captureTurn(writer http.ResponseWriter, request *http.Request) {
	var input model.CaptureTurnRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.CaptureTurn(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrIdempotencyConflict) {
			writeError(writer, http.StatusConflict, "idempotency_conflict", err)
			return
		}
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "capture_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) recall(writer http.ResponseWriter, request *http.Request) {
	var input model.RecallRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Recall(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrStaleSnapshot) {
			writeError(writer, http.StatusConflict, "snapshot_changed", err)
			return
		}
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "recall_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
}

func (s *Server) recallOpenClaw(writer http.ResponseWriter, request *http.Request) {
	var input model.RecallRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Recall(request.Context(), input)
	if err != nil {
		if errors.Is(err, store.ErrStaleSnapshot) {
			writeError(writer, http.StatusConflict, "snapshot_changed", err)
			return
		}
		if writeDependencyUnavailable(writer, err) {
			return
		}
		writeError(writer, http.StatusBadRequest, "recall_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, projectOpenClawPacket(response))
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, target any) error {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code string, err error) {
	writeJSON(writer, status, model.ErrorResponse{ProtocolVersion: model.ProtocolVersion, Code: code, Message: err.Error()})
}

func writeDependencyUnavailable(writer http.ResponseWriter, err error) bool {
	if !errors.Is(err, retrieval.ErrContractCircuitOpen) {
		return false
	}
	writer.Header().Set("Retry-After", "5")
	writeError(writer, http.StatusServiceUnavailable, "dependency_unavailable", err)
	return true
}

func requestLog(logger *slog.Logger, metrics *runtimeMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		observed := &statusWriter{ResponseWriter: writer, status: http.StatusOK}
		logger.Debug("request", "method", request.Method, "path", request.URL.Path)
		next.ServeHTTP(observed, request)
		metrics.observe(time.Since(started), observed.status)
	})
}
