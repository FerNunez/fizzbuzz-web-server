package handler

const (
	codeBadRequest    = "bad_request"
	codeInvalidParams = "invalid_params"
	codeInternal      = "internal"
)

type RequestParams struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
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
