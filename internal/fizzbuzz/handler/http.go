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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong request"))
		return
	}
	r.Body.Close()

	fizzbuzzed, err := h.serivce.Generate(r.Context(), &domain.FizzbuzzParams{
		Int1:  req.Int1,
		Int2:  req.Int2,
		Limit: req.Limit,
		Str1:  req.Str1,
		Str2:  req.Str2,
	})

	if errors.Is(err, domain.ErrInvalidParams) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not process request"))
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{Data: fizzbuzzed})
	}
}

func (h *FizzbuzzHandler) HandleStatistics(w http.ResponseWriter, r *http.Request) {
	paramStats, err := h.serivce.GetMostFrequent(r.Context())

	if errors.Is(err, domain.ErrEmptyRepo) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{Data: domain.ParamStat{}})
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("could not process request"))
	} else {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(APIResponse{Data: paramStats})
	}
}
