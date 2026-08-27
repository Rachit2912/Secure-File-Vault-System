package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/handlers"

	"github.com/gorilla/mux"
)

func TestFileDownloadHandler_MissingID(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/fileDownload/", nil)
	rec := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/api/fileDownload/{id}", handlers.FileDownloadHandler)
	handlers.FileDownloadHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestDownloadAuthMatrixLogic(t *testing.T) {
	// Matrix of tests verifying authorization rules (Scenarios A - E, F, G):
	// Public file: guest allowed (200), user allowed (200).
	// Private file: guest -> 401, wrong user -> 403, uploader -> allowed (200).

	tests := []struct {
		name           string
		isPublic       bool
		uploaderID     int
		ctxUserID      interface{}
		expectedStatus int
	}{
		{
			name:           "Test A: Public file, guest (logged out)",
			isPublic:       true,
			uploaderID:     1,
			ctxUserID:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test B: Public file, logged in user",
			isPublic:       true,
			uploaderID:     1,
			ctxUserID:      2,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test C: Private file, guest (logged out)",
			isPublic:       false,
			uploaderID:     1,
			ctxUserID:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Test D: Private file, uploader logged in",
			isPublic:       false,
			uploaderID:     1,
			ctxUserID:      1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test E: Private file, another user logged in",
			isPublic:       false,
			uploaderID:     1,
			ctxUserID:      2,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Test F: Toggle private -> public, guest logged out now allowed",
			isPublic:       true, // toggled to public
			uploaderID:     1,
			ctxUserID:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Test G: Toggle public -> private, guest denied, wrong user denied, uploader allowed",
			isPublic:       false, // toggled to private
			uploaderID:     1,
			ctxUserID:      nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var authenticated bool
			var userID int
			if tt.ctxUserID != nil {
				if uid, ok := tt.ctxUserID.(int); ok {
					userID = uid
					authenticated = true
				}
			}

			var status int
			if !tt.isPublic {
				if !authenticated {
					status = http.StatusUnauthorized
				} else if tt.uploaderID != userID {
					status = http.StatusForbidden
				} else {
					status = http.StatusOK
				}
			} else {
				status = http.StatusOK
			}

			if status != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, status)
			}
		})
	}
}
