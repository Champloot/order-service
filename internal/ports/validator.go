package ports

import "order-service/internal/validation"
import "order-service/internal/models"

//go:generate mockgen -source=validator.go -destination=../mocks/mock_validator.go -package=mocks

type Validator interface {
	Validate(order *models.Order) validation.ValidationResult
}