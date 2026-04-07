package httphandler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/kweaver-ai/agent-observability/src/domain/service/tracesvc"
	"github.com/kweaver-ai/agent-observability/src/driveradapter/api/rdto"
)

type TraceHandler struct {
	traceQueryService *tracesvc.TraceQueryService
}

func NewTraceHandler(traceQueryService *tracesvc.TraceQueryService) *TraceHandler {
	return &TraceHandler{traceQueryService: traceQueryService}
}

// SearchTraces godoc
// @Summary Search traces with raw OpenSearch DSL
// @Description Proxy raw OpenSearch DSL to the configured trace index and return the original OpenSearch response body.
// @Tags traces
// @Accept json
// @Produce json
// @Param request body string true "OpenSearch DSL JSON body"
// @Success 200 {string} string "Raw OpenSearch search response"
// @Failure 400 {object} rdto.ErrorResponse
// @Failure 405 {object} rdto.ErrorResponse
// @Failure 504 {object} rdto.ErrorResponse
// @Router /api/agent-observability/v1/traces/_search [post]
func (h *TraceHandler) SearchTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, rdto.ErrorResponse{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "only POST is supported",
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, rdto.ErrorResponse{
			Code:    "INVALID_ARGUMENT",
			Message: "failed to read request body",
		})
		return
	}

	if len(body) == 0 {
		writeJSON(w, http.StatusBadRequest, rdto.ErrorResponse{
			Code:    "INVALID_ARGUMENT",
			Message: "opensearch dsl body is required",
		})
		return
	}

	var raw json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		writeJSON(w, http.StatusBadRequest, rdto.ErrorResponse{
			Code:    "INVALID_ARGUMENT",
			Message: "request body must be valid json",
		})
		return
	}

	traceData, err := h.traceQueryService.SearchTraces(r.Context(), raw)
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, rdto.ErrorResponse{
			Code:    "QUERY_FAILED",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(traceData)
}

// SearchTracesByConversationID godoc
// @Summary Search traces by conversation ID
// @Description Build a term filter automatically using attributes.gen_ai.conversation.id.keyword and return the original OpenSearch response body.
// @Tags traces
// @Accept json
// @Produce json
// @Param conversation_id query string true "Conversation ID"
// @Success 200 {string} string "Raw OpenSearch search response"
// @Failure 400 {object} rdto.ErrorResponse
// @Failure 405 {object} rdto.ErrorResponse
// @Failure 500 {object} rdto.ErrorResponse
// @Failure 504 {object} rdto.ErrorResponse
// @Router /api/agent-observability/v1/traces/by-conversation [get]
func (h *TraceHandler) SearchTracesByConversationID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, rdto.ErrorResponse{
			Code:    "METHOD_NOT_ALLOWED",
			Message: "only GET is supported",
		})
		return
	}

	conversationID := r.URL.Query().Get("conversation_id")
	if conversationID == "" {
		writeJSON(w, http.StatusBadRequest, rdto.ErrorResponse{
			Code:    "INVALID_ARGUMENT",
			Message: "conversation_id is required",
		})
		return
	}

	query, err := json.Marshal(map[string]any{
		"size": 1000,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{
						"term": map[string]string{
							"attributes.gen_ai.conversation.id.keyword": conversationID,
						},
					},
				},
			},
		},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, rdto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "failed to build query",
		})
		return
	}

	traceData, err := h.traceQueryService.SearchTraces(r.Context(), query)
	if err != nil {
		writeJSON(w, http.StatusGatewayTimeout, rdto.ErrorResponse{
			Code:    "QUERY_FAILED",
			Message: err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(traceData)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
