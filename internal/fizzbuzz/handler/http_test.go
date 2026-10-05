package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"fizzbuzz-web-server/internal/fizzbuzz/domain"
)

// fakeService inplements FizzbuzzService interface.
// It returns canned values and errors for the service
type fakeService struct {
	out     []string
	err     error
	stat    *domain.ParamStat
	statErr error
}

func (f *fakeService) Generate(_ context.Context, _ *domain.FizzbuzzParams) ([]string, error) {
	return f.out, f.err
}

func (f *fakeService) GetMostFrequent(_ context.Context) (*domain.ParamStat, error) {
	return f.stat, f.statErr
}

// testResponse mirrors APIResponse, keeping Data raw so each test decodes it into its own type.
type testResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *APIError       `json:"error"`
}

func newTestHandler(svc domain.FizzbuzzService) *FizzbuzzHandler {
	return NewFizzbuzzHandler(svc, slog.New(slog.DiscardHandler))
}

// decode checks the Content-Type, decodes the response data into the given pointer and returns the API error, if any.
func decode(t *testing.T, rec *httptest.ResponseRecorder, data any) *APIError {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expects Content-Type application/json, got %q", ct)
	}
	var resp testResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("expects valid JSON body, got error: %v", err)
	}
	if len(resp.Data) > 0 {
		if err := json.Unmarshal(resp.Data, data); err != nil {
			t.Fatalf("expects valid data, got error: %v", err)
		}
	}
	return resp.Error
}

func TestHandleFizzbuzz(t *testing.T) {
	validBody := `{"int1":3,"int2":5,"limit":5,"str1":"fizz","str2":"buzz"}`

	tests := []struct {
		name       string
		body       string
		svc        *fakeService
		wantStatus int
		wantCode   string
		wantData   []string
	}{
		{
			name:       "valid body",
			body:       validBody,
			svc:        &fakeService{out: []string{"1", "2", "fizz", "4", "buzz"}},
			wantStatus: http.StatusOK,
			wantData:   []string{"1", "2", "fizz", "4", "buzz"},
		},
		{
			name:       "malformed JSON",
			body:       `{`,
			svc:        &fakeService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   codeBadRequest,
		},
		{
			name:       "unknown field",
			body:       `{"foo":1}`,
			svc:        &fakeService{},
			wantStatus: http.StatusBadRequest,
			wantCode:   codeBadRequest,
		},
		{
			name:       "svc return invalid params",
			body:       validBody,
			svc:        &fakeService{err: domain.ErrInvalidParams},
			wantStatus: http.StatusBadRequest,
			wantCode:   codeInvalidParams,
		},
		{
			name:       "service return error",
			body:       validBody,
			svc:        &fakeService{err: errors.New("an_error")},
			wantStatus: http.StatusInternalServerError,
			wantCode:   codeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptest requester and response recorder
			req := httptest.NewRequest(http.MethodPost, "/fizzbuzz", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			newTestHandler(tt.svc).HandleFizzbuzz(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expects status %d, got %d", tt.wantStatus, rec.Code)
			}
			var data []string
			apiErr := decode(t, rec, &data)
			if tt.wantCode != "" {
				if apiErr == nil || apiErr.Code != tt.wantCode {
					t.Fatalf("expects error code %q, got %+v", tt.wantCode, apiErr)
				}
				return
			}
			if !slices.Equal(data, tt.wantData) {
				t.Fatalf("expects data %v, got %v", tt.wantData, data)
			}
		})
	}
}

func TestHandleStatistics(t *testing.T) {
	params := domain.FizzbuzzParams{Int1: 3, Int2: 5, Limit: 16, Str1: "fizz", Str2: "buzz"}

	tests := []struct {
		name       string
		svc        *fakeService
		wantStatus int
		wantCode   string
		want       statisticsResponse
	}{
		{
			name:       "empty repo",
			svc:        &fakeService{stat: &domain.ParamStat{}},
			wantStatus: http.StatusOK,
			want:       statisticsResponse{},
		},
		{
			name:       "most frequent",
			svc:        &fakeService{stat: &domain.ParamStat{Params: &params, Hits: 3}},
			wantStatus: http.StatusOK,
			want:       statisticsResponse{Params: toRequestParams(&params), Hits: 3},
		},
		{
			name:       "service error",
			svc:        &fakeService{statErr: errors.New("boom")},
			wantStatus: http.StatusInternalServerError,
			wantCode:   codeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/fizzbuzz/statistics", nil)
			rec := httptest.NewRecorder()

			newTestHandler(tt.svc).HandleStatistics(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expects status %d, got %d", tt.wantStatus, rec.Code)
			}
			var data statisticsResponse
			apiErr := decode(t, rec, &data)
			if tt.wantCode != "" {
				if apiErr == nil || apiErr.Code != tt.wantCode {
					t.Fatalf("expects error code %q, got %+v", tt.wantCode, apiErr)
				}
				return
			}
			if data.Hits != tt.want.Hits || !equalParams(data.Params, tt.want.Params) {
				t.Fatalf("expects %+v, got %+v", tt.want, data)
			}
		})
	}
}

func equalParams(a, b *fizzbuzzRequestParams) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
