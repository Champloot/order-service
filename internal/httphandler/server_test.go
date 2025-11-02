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
	"sync"
	"time"

	"order-service/internal/httphandler"
	"order-service/internal/models"
	"order-service/internal/mocks"
	"order-service/internal/ports"


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

func TestGetOrderHandler_FromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockOrderCache(ctrl)
	mockRepo := mocks.NewMockOrderRepository(ctrl)

	order := &models.Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "b563feb7b2b84b6test",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
		Locale:            "en",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Now(),
		OofShard:          "1",
	}

	mockCache.EXPECT().
		GetOrder(gomock.Any(), "test-order-123").
		Return(order, nil).
		Times(1)

	mockRepo.EXPECT().
		GetOrder(gomock.Any(), "test-order-123").
		Times(0)

	server := httphandler.NewServer(mockCache, mockRepo)
	handler := server.GetHandler()

	req, err := http.NewRequest("GET", "/api/order/test-order-123", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "cache", response["source"])
	assert.Contains(t, response, "order")
	assert.Contains(t, response, "timing")

	orderData := response["order"].(map[string]interface{})
	assert.Equal(t, "test-order-123", orderData["order_uid"])
}

func TestGetOrderHandler_FromDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockOrderCache(ctrl)
	mockRepo := mocks.NewMockOrderRepository(ctrl)

	order := &models.Order{
		OrderUID:		"test-order-456",
		TrackNumber:	"WBILMTESTTRACK456",
		Entry:			"WBIL",
		DateCreated:	time.Now(),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// not found
	mockCache.EXPECT().
		GetOrder(gomock.Any(), "test-order-456").
		Return(nil, nil).
		Times(1)

	mockRepo.EXPECT().
		GetOrder(gomock.Any(), "test-order-456").
		Return(order, nil).
		Times(1)

	mockCache.EXPECT().
		SetOrder(gomock.Any(), order).
		DoAndReturn(func(ctx interface{}, order interface{}) error {
			defer wg.Done()
			return nil
		}).
		Times(1)

	server := httphandler.NewServer(mockCache, mockRepo)
	handler := server.GetHandler()

	req, err := http.NewRequest("GET", "/api/order/test-order-456", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	wg.Wait()

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "database", response["source"])
}

func TestGetOrderHandler_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCache := mocks.NewMockOrderCache(ctrl)
	mockRepo := mocks.NewMockOrderRepository(ctrl)

	// not found
	mockCache.EXPECT().
		GetOrder(gomock.Any(), "non-existent-order").
		Return(nil, nil).
		Times(1)

	mockRepo.EXPECT().
		GetOrder(gomock.Any(), "non-existent-order").
		Return(nil, ports.ErrOrderNotFound).
		Times(1)

	server := httphandler.NewServer(mockCache, mockRepo)
	handler := server.GetHandler()

	req, err := http.NewRequest("GET", "/api/order/non-existent-order", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Order not found")
}