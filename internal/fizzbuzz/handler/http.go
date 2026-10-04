package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

type FizzbuzzHandler struct {
	serivce domain.FizzbuzzService
}

func NewFizzbuzzHandler(serivce domain.FizzbuzzService) *FizzbuzzHandler {
	return &FizzbuzzHandler{
		serivce: serivce,
	}
}

func (h *FizzbuzzHandler) HandleFizzbuzz(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var req RequestParams
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, codeBadRequest, "invalid JSON body")
		return
	}

	fizzbuzzed, err := h.serivce.Generate(r.Context(), &domain.FizzbuzzParams{
		Int1:  req.Int1,
		Int2:  req.Int2,
		Limit: req.Limit,
		Str1:  req.Str1,
		Str2:  req.Str2,
	})

	if errors.Is(err, domain.ErrInvalidParams) {
		writeError(w, http.StatusBadRequest, codeInvalidParams, err.Error())
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
	} else {
		// TODO: Returns OK or created. To verify
		writeJSON(w, http.StatusOK, fizzbuzzed)
	}
}

func (h *FizzbuzzHandler) HandleStatistics(w http.ResponseWriter, r *http.Request) {
	paramStats, err := h.serivce.GetMostFrequent(r.Context())

	// TODO: This feels weird, service sending ERROR but http handler returns OK
	// Maybe better if the systems return new paramStats
	if errors.Is(err, domain.ErrEmptyRepo) {
		writeJSON(w, http.StatusOK, domain.ParamStat{})
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, codeInternal, "internal error")
	} else {
		writeJSON(w, http.StatusOK, paramStats)
	}
}
