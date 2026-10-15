package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"pet-community/types"
)

func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	return ctx, recorder
}

func TestRespondSuccess(t *testing.T) {
	ctx, recorder := setupTestContext()

	RespondSuccess(ctx, map[string]string{"key": "value"})

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var resp types.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != types.CodeSuccess {
		t.Errorf("expected code %d, got %d", types.CodeSuccess, resp.Code)
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		wantStatus int
	}{
		{"unauthorized", types.CodeUnauthorized, http.StatusUnauthorized},
		{"forbidden", types.CodeForbidden, http.StatusForbidden},
		{"not found", types.CodeNotFound, http.StatusNotFound},
		{"system error", types.CodeSystemError, http.StatusInternalServerError},
		{"default bad request", 999, http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, recorder := setupTestContext()
			RespondError(ctx, tc.code, "error message")

			if recorder.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, recorder.Code)
			}

			var resp types.Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}
			if resp.Code != tc.code {
				t.Errorf("expected code %d, got %d", tc.code, resp.Code)
			}
		})
	}
}

func TestRespondValidationError(t *testing.T) {
	ctx, recorder := setupTestContext()

	errors := map[string]string{"field": "required"}
	RespondValidationError(ctx, errors)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	var resp types.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != types.CodeParamError {
		t.Errorf("expected code %d, got %d", types.CodeParamError, resp.Code)
	}
}
