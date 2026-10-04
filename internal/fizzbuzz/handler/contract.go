package handler

import "fizzbuzz-web-server/internal/fizzbuzz/domain"

const (
	codeBadRequest    = "bad_request"
	codeInvalidParams = "invalid_params"
	codeInternal      = "internal"
)

type fizzbuzzRequestParams struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

type statisticsResponse struct {
	Params *fizzbuzzRequestParams `json:"params"`
	Hits   int                    `json:"hits"`
}

func toRequestParams(p *domain.FizzbuzzParams) *fizzbuzzRequestParams {
	if p == nil {
		return nil
	}
	return &fizzbuzzRequestParams{
		Int1:  p.Int1,
		Int2:  p.Int2,
		Limit: p.Limit,
		Str1:  p.Str1,
		Str2:  p.Str2,
	}
}

// API Response is the standard response envelope for all HTTP endpoints
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
