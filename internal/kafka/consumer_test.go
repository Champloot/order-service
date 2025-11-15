package kafka_test
import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"order-service/internal/kafka"
	"order-service/internal/models"
	"order-service/internal/mocks"
	"order-service/internal/ports"
	"order-service/internal/validation"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestConsumer_ProcessMessage_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockValidator := mocks.NewMockValidator(ctrl)

    mockValidator.EXPECT().
        Validate(gomock.Any()).
        Return(validation.ValidationResult{IsValid: true, Errors: nil}).
        Times(1)

	order := &models.Order{
		OrderUID:		"test-order-123",
		TrackNumber:	"WBILMTESTTRACK",
		Entry:			"WBIL",
		DateCreated:	time.Now(),
	}

	orderJSON, err := json.Marshal(order)
	require.NoError(t, err)

	mockRepo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		Return(nil).
		Times(1)	

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"orders",
		mockRepo,
		10*time.Second,
		10240,
		10*1024*1024,
		1*time.Second,
		5*time.Second,
		mockValidator,
	)

	err = consumer.ProcessMessage(context.Background(), orderJSON)
	assert.NoError(t, err)
}

func TestConsumer_ProcessMessage_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockValidator := mocks.NewMockValidator(ctrl)


	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"orders",
		mockRepo,
		10*time.Second,
		10240,
		10*1024*1024,
		1*time.Second,
		5*time.Second,
        mockValidator,
	)

	invalidJSON := []byte(`{"invalid": json`)

	err := consumer.ProcessMessage(context.Background(), invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to unmarshal order")
}

func TestConsumer_ProcessMessage_EmptyOrderUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockValidator := mocks.NewMockValidator(ctrl)


	  mockValidator.EXPECT().
        Validate(gomock.Any()).
        Return(validation.ValidationResult{
            IsValid: false,
            Errors: []validation.ValidationError{
                {
                    Field:   "order_uid",
                    Message: "order UID is required",
                },
            },
        }).
        Times(1)

    consumer := kafka.NewConsumer(
        []string{"localhost:9092"},
        "orders",
        mockRepo,
        10*time.Second,
        10240,
        10*1024*1024,
        1*time.Second,
        5*time.Second,
        mockValidator,
    )

	order := &models.Order{
		OrderUID: "",
	}
	orderJSON, err := json.Marshal(order)
	require.NoError(t, err)

	err = consumer.ProcessMessage(context.Background(), orderJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order validation failed")
}

func TestConsumer_ProcessMessage_TransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockValidator := mocks.NewMockValidator(ctrl)

    mockValidator.EXPECT().
        Validate(gomock.Any()).
        Return(validation.ValidationResult{IsValid: true, Errors: nil}).
        Times(1)

	order := &models.Order{
		OrderUID:    "test-order-456",
		TrackNumber: "WBILMTESTTRACK456",
		Entry:       "WBIL",
		DateCreated: time.Now(),
	}

	orderJSON, err := json.Marshal(order)
	require.NoError(t, err)

	mockRepo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		Return(ports.ErrTxFailed).
		Times(1)

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"orders",
		mockRepo,
		10*time.Second,
		10240,
		10*1024*1024,
		1*time.Second,
		5*time.Second,
        mockValidator,
	)

	err = consumer.ProcessMessage(context.Background(), orderJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save order")
}

func TestConsumer_ProcessMessage_SaveErrorInTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockOrderRepository(ctrl)
    mockValidator := mocks.NewMockValidator(ctrl)

    mockValidator.EXPECT().
        Validate(gomock.Any()).
        Return(validation.ValidationResult{IsValid: true, Errors: nil}).
        Times(1)

	order := &models.Order{
		OrderUID:    "test-order-789",
		TrackNumber: "WBILMTESTTRACK789",
		Entry:       "WBIL",
		DateCreated: time.Now(),
	}

	orderJSON, err := json.Marshal(order)
	require.NoError(t, err)

	mockRepo.EXPECT().
		WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(tx ports.OrderTx) error) error {
			return ports.ErrTxFailed
		}).
		Times(1)

	consumer := kafka.NewConsumer(
		[]string{"localhost:9092"},
		"orders",
		mockRepo,
		10*time.Second,
		10240,
		10*1024*1024,
		1*time.Second,
		5*time.Second,
        mockValidator,
	)

	err = consumer.ProcessMessage(context.Background(), orderJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to save order")
}