package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

type FizzbuzzHandler struct {
	service domain.FizzbuzzService
	logger  *slog.Logger
}

func NewFizzbuzzHandler(serivce domain.FizzbuzzService, logger *slog.Logger) *FizzbuzzHandler {
	return &FizzbuzzHandler{
		service: serivce,
		logger:  logger,
	}
}

func (h *FizzbuzzHandler) HandleFizzbuzz(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req fizzbuzzRequestParams
	if err := decoder.Decode(&req); err != nil {
		h.logger.Debug("error while decoding RequestParams", "err", err)
		writeError(w, http.StatusBadRequest, codeBadRequest, "invalid JSON body")
		return
	}

	h.logger.Debug("fizzbuzz request", "int1", req.Int1, "int2", req.Int2, "limit", req.Limit, "str1", req.Str1, "str2", req.Str2)
	fizzbuzzed, err := h.service.Generate(r.Context(), &domain.FizzbuzzParams{
		Int1:  req.Int1,
		Int2:  req.Int2,
		Limit: req.Limit,
		Str1:  req.Str1,
		Str2:  req.Str2,
	})

	if errors.Is(err, domain.ErrInvalidParams) {
		h.logger.Debug("error while validating params", "err", err)
		writeError(w, http.StatusBadRequest, codeInvalidParams, err.Error())
	} else if err != nil {
		h.logger.Error("internal server error", "err", err)
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
	} else {
		// TODO: Returns OK or created. To verify
		writeJSON(w, http.StatusOK, fizzbuzzed)
	}
}

func (h *FizzbuzzHandler) HandleStatistics(w http.ResponseWriter, r *http.Request) {
	// TODO: add any params give error?
	paramStats, err := h.service.GetMostFrequent(r.Context())
	if err != nil {
		h.logger.Error("internal server error", "err", err)
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, statisticsResponse{
		Params: toRequestParams(paramStats.Params),
		Hits:   paramStats.Hits,
	})
}
