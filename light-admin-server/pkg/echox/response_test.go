package echox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	apperrors "github.com/top-system/light-admin/errors"
)

func TestResponseBusinessCodes(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		status   int
		code     string
	}{
		{"success", Response{Code: http.StatusOK}, http.StatusOK, "00000"},
		{"validation", Response{Code: http.StatusBadRequest}, http.StatusBadRequest, "A0400"},
		{"unauthorized", Response{Code: http.StatusUnauthorized}, http.StatusUnauthorized, "A0401"},
		{"forbidden", Response{Code: http.StatusForbidden}, http.StatusForbidden, "A0403"},
		{"not found", Response{Code: http.StatusNotFound}, http.StatusNotFound, "A0404"},
		{"server", Response{Code: http.StatusInternalServerError}, http.StatusInternalServerError, "C0500"},
		{"domain conflict", Response{Code: http.StatusBadRequest, Message: apperrors.UserAlreadyExists}, http.StatusConflict, "B1001"},
		{"wrapped domain conflict", Response{Code: http.StatusBadRequest, Message: fmt.Errorf("create user: %w", apperrors.UserAlreadyExists)}, http.StatusConflict, "B1001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			recorder := httptest.NewRecorder()
			ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/", nil), recorder)
			if err := tt.response.JSON(ctx); err != nil {
				t.Fatal(err)
			}
			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}
			var body struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tt.code {
				t.Fatalf("code = %q, want %q", body.Code, tt.code)
			}
		})
	}
}
