package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		limit    int64
		body     string
		wantCode int
		wantBody string
		wantErr  bool
	}{
		{
			name:     "should read a body under the limit",
			limit:    10,
			body:     "hello",
			wantCode: http.StatusOK,
			wantBody: "hello",
		},
		{
			name:     "should read a body exactly at the limit",
			limit:    5,
			body:     "hello",
			wantCode: http.StatusOK,
			wantBody: "hello",
		},
		{
			name:     "should fail reading a body over the limit",
			limit:    4,
			body:     "hello",
			wantCode: http.StatusRequestEntityTooLarge,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				gotBody string
				gotErr  error
			)
			handler := RequestSize(tt.limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				gotBody, gotErr = string(b), err
				if err != nil {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantCode)
			}

			if tt.wantErr {
				var maxErr *http.MaxBytesError
				if !errors.As(gotErr, &maxErr) {
					t.Fatalf("error = %v, want *http.MaxBytesError", gotErr)
				}
				if maxErr.Limit != tt.limit {
					t.Errorf("MaxBytesError.Limit = %d, want %d", maxErr.Limit, tt.limit)
				}
				return
			}

			if gotErr != nil {
				t.Fatalf("unexpected error: %v", gotErr)
			}
			if gotBody != tt.wantBody {
				t.Errorf("body = %q, want %q", gotBody, tt.wantBody)
			}
		})
	}
}
