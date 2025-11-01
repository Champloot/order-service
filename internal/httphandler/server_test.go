package httphandler_test

import (
	// "bytes"
	// "context"
	"encoding/json"
	// "fmt"
	"net/http"
	"net/http/httptest"
	// "strings"
	"testing"
	// "time"

	"order-service/internal/httphandler"
	// "order-service/internal/models"
	"order-service/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHealthHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockOrderCache(ctrl)
	mockRepo := mocks.NewMockOrderRepository(ctrl)

	server := httphandler.NewServer(mockCache, mockRepo)
	handler := server.GetHandler()

	req, err := http.NewRequest("GET", "/api/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response["timestamp"], "T")
}