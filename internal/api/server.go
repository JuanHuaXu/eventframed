package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/JuanHuaXu/eventframed/internal/model"
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
	server.mux.HandleFunc("POST /v1/events:observe", server.observe)
	server.mux.HandleFunc("POST /v1/context:recall", server.recall)
	server.mux.HandleFunc("POST /v1/events:delete", server.delete)
	server.mux.HandleFunc("POST /v1/maintenance:retain", server.retain)
	server.mux.HandleFunc("POST /v1/maintenance:backup", server.backup)
	server.mux.HandleFunc("POST /v1/maintenance:compact", server.compact)
	server.mux.HandleFunc("GET /metrics", server.metrics.handle)
	return server
}

func (s *Server) delete(writer http.ResponseWriter, request *http.Request) {
	var input model.DeleteRequest
	if err := decodeJSON(writer, request, &input); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_request", err)
		return
	}
	response, err := s.service.Delete(request.Context(), input)
	if err != nil {
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
		writeError(writer, http.StatusBadRequest, "observe_rejected", err)
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
		writeError(writer, http.StatusBadRequest, "recall_rejected", err)
		return
	}
	writeJSON(writer, http.StatusOK, response)
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

func requestLog(logger *slog.Logger, metrics *runtimeMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		observed := &statusWriter{ResponseWriter: writer, status: http.StatusOK}
		logger.Debug("request", "method", request.Method, "path", request.URL.Path)
		next.ServeHTTP(observed, request)
		metrics.observe(time.Since(started), observed.status)
	})
}
