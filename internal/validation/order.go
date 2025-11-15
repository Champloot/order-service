package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"order-service/internal/models"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

type Validator interface {
	Validate(order *models.Order) ValidationResult
}

type OrderValidator struct {
	config ValidationConfig
}

type ValidationConfig struct {
	RequireOrderUID      bool `json:"require_order_uid"`
	RequireTrackNumber   bool `json:"require_track_number"`
	RequireEntry         bool `json:"require_entry"`
	RequireCustomerID    bool `json:"require_customer_id"`
	RequireDelivery      bool `json:"require_delivery"`
	RequirePayment       bool `json:"require_payment"`
	RequireItems         bool `json:"require_items"`
	
	ValidateEmail       bool `json:"validate_email"`
	ValidatePhone       bool `json:"validate_phone"`
	ValidateAmounts     bool `json:"validate_amounts"`
	ValidateDates       bool `json:"validate_dates"`
	
	MaxItems          int           `json:"max_items"`
	MinAmount         int           `json:"min_amount"`
	MaxAmount         int           `json:"max_amount"`
	MaxDateFuture     time.Duration `json:"max_date_future"`
	MaxDatePast       time.Duration `json:"max_date_past"`
}

func DefaultConfig() ValidationConfig {
	return ValidationConfig{
		RequireOrderUID:    true,
		RequireTrackNumber: true,
		RequireEntry:       true,
		RequireCustomerID:  true,
		RequireDelivery:    true,
		RequirePayment:     true,
		RequireItems:       true,
		
		ValidateEmail:   true,
		ValidatePhone:   true,
		ValidateAmounts: true,
		ValidateDates:   true,
		
		MaxItems:      100,
		MinAmount:     0,
		MaxAmount:     1_000_000_00,
		MaxDateFuture: 24 * time.Hour,
		MaxDatePast:   365 * 24 * time.Hour,
	}
}

func NewOrderValidator(config ValidationConfig) *OrderValidator {
	return &OrderValidator{
		config: config,
	}
}

func (v *OrderValidator) Validate(order *models.Order) ValidationResult {
	var errors []ValidationError

	errors = append(errors, v.validateBasicFields(order)...)
	
	if v.config.RequireDelivery {
		errors = append(errors, v.validateDelivery(&order.Delivery)...)
	}
	
	if v.config.RequirePayment {
		errors = append(errors, v.validatePayment(&order.Payment)...)
	}
	
	if v.config.RequireItems {
		errors = append(errors, v.validateItems(order.Items)...)
	}
	
	if v.config.ValidateDates {
		errors = append(errors, v.validateDates(order)...)
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

func (v *OrderValidator) validateBasicFields(order *models.Order) []ValidationError {
	var errors []ValidationError

	if v.config.RequireOrderUID && strings.TrimSpace(order.OrderUID) == "" {
		errors = append(errors, ValidationError{
			Field:   "order_uid",
			Message: "order UID is required",
		})
	}

	if v.config.RequireTrackNumber && strings.TrimSpace(order.TrackNumber) == "" {
		errors = append(errors, ValidationError{
			Field:   "track_number",
			Message: "track number is required",
		})
	}

	if v.config.RequireEntry && strings.TrimSpace(order.Entry) == "" {
		errors = append(errors, ValidationError{
			Field:   "entry",
			Message: "entry is required",
		})
	}

	if v.config.RequireCustomerID && strings.TrimSpace(order.CustomerID) == "" {
		errors = append(errors, ValidationError{
			Field:   "customer_id",
			Message: "customer ID is required",
		})
	}

	if len(order.OrderUID) > 255 {
		errors = append(errors, ValidationError{
			Field:   "order_uid",
			Message: "order UID too long (max 255 characters)",
			Value:   len(order.OrderUID),
		})
	}

	return errors
}

func (v *OrderValidator) validateDelivery(delivery *models.Delivery) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(delivery.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "delivery.name",
			Message: "delivery name is required",
		})
	}

	if strings.TrimSpace(delivery.Phone) == "" {
		errors = append(errors, ValidationError{
			Field:   "delivery.phone",
			Message: "delivery phone is required",
		})
	} else if v.config.ValidatePhone {
		if !isValidPhone(delivery.Phone) {
			errors = append(errors, ValidationError{
				Field:   "delivery.phone",
				Message: "invalid phone format",
				Value:   delivery.Phone,
			})
		}
	}

	if strings.TrimSpace(delivery.Zip) == "" {
		errors = append(errors, ValidationError{
			Field:   "delivery.zip",
			Message: "delivery zip code is required",
		})
	}

	if strings.TrimSpace(delivery.City) == "" {
		errors = append(errors, ValidationError{
			Field:   "delivery.city",
			Message: "delivery city is required",
		})
	}

	if strings.TrimSpace(delivery.Address) == "" {
		errors = append(errors, ValidationError{
			Field:   "delivery.address",
			Message: "delivery address is required",
		})
	}

	if v.config.ValidateEmail && strings.TrimSpace(delivery.Email) != "" {
		if !isValidEmail(delivery.Email) {
			errors = append(errors, ValidationError{
				Field:   "delivery.email",
				Message: "invalid email format",
				Value:   delivery.Email,
			})
		}
	}

	return errors
}

func (v *OrderValidator) validatePayment(payment *models.Payment) []ValidationError {
	var errors []ValidationError

	if strings.TrimSpace(payment.Transaction) == "" {
		errors = append(errors, ValidationError{
			Field:   "payment.transaction",
			Message: "payment transaction is required",
		})
	}

	if strings.TrimSpace(payment.Currency) == "" {
		errors = append(errors, ValidationError{
			Field:   "payment.currency",
			Message: "payment currency is required",
		})
	} else if len(payment.Currency) != 3 {
		errors = append(errors, ValidationError{
			Field:   "payment.currency",
			Message: "currency must be 3 characters (ISO 4217)",
			Value:   payment.Currency,
		})
	}

	if strings.TrimSpace(payment.Provider) == "" {
		errors = append(errors, ValidationError{
			Field:   "payment.provider",
			Message: "payment provider is required",
		})
	}

	if v.config.ValidateAmounts {
		if payment.Amount < v.config.MinAmount {
			errors = append(errors, ValidationError{
				Field:   "payment.amount",
				Message: fmt.Sprintf("amount too small (min %d)", v.config.MinAmount),
				Value:   payment.Amount,
			})
		}

		if payment.Amount > v.config.MaxAmount {
			errors = append(errors, ValidationError{
				Field:   "payment.amount",
				Message: fmt.Sprintf("amount too large (max %d)", v.config.MaxAmount),
				Value:   payment.Amount,
			})
		}

		calculatedTotal := payment.DeliveryCost + payment.GoodsTotal + payment.CustomFee
		if payment.Amount != calculatedTotal {
			errors = append(errors, ValidationError{
				Field:   "payment.amount_consistency",
				Message: fmt.Sprintf("amount (%d) doesn't match sum of delivery_cost + goods_total + custom_fee (%d)", 
					payment.Amount, calculatedTotal),
				Value: map[string]int{
					"amount":       payment.Amount,
					"delivery_cost": payment.DeliveryCost,
					"goods_total":   payment.GoodsTotal,
					"custom_fee":    payment.CustomFee,
				},
			})
		}

		if payment.DeliveryCost < 0 {
			errors = append(errors, ValidationError{
				Field:   "payment.delivery_cost",
				Message: "delivery cost cannot be negative",
				Value:   payment.DeliveryCost,
			})
		}

		if payment.GoodsTotal < 0 {
			errors = append(errors, ValidationError{
				Field:   "payment.goods_total",
				Message: "goods total cannot be negative",
				Value:   payment.GoodsTotal,
			})
		}

		if payment.CustomFee < 0 {
			errors = append(errors, ValidationError{
				Field:   "payment.custom_fee",
				Message: "custom fee cannot be negative",
				Value:   payment.CustomFee,
			})
		}
	}

	if payment.PaymentDt <= 0 {
		errors = append(errors, ValidationError{
			Field:   "payment.payment_dt",
			Message: "payment date must be positive",
			Value:   payment.PaymentDt,
		})
	} else {
		paymentTime := time.Unix(payment.PaymentDt, 0)
		if paymentTime.After(time.Now().Add(v.config.MaxDateFuture)) {
			errors = append(errors, ValidationError{
				Field:   "payment.payment_dt",
				Message: "payment date is too far in the future",
				Value:   paymentTime.Format(time.RFC3339),
			})
		}
	}

	return errors
}

func (v *OrderValidator) validateItems(items []models.Item) []ValidationError {
	var errors []ValidationError

	if len(items) == 0 {
		errors = append(errors, ValidationError{
			Field:   "items",
			Message: "at least one item is required",
		})
		return errors
	}

	if len(items) > v.config.MaxItems {
		errors = append(errors, ValidationError{
			Field:   "items",
			Message: fmt.Sprintf("too many items (max %d)", v.config.MaxItems),
			Value:   len(items),
		})
	}

	for i, item := range items {
		itemPrefix := fmt.Sprintf("items[%d].", i)

		if strings.TrimSpace(item.TrackNumber) == "" {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "track_number",
				Message: "item track number is required",
			})
		}

		if strings.TrimSpace(item.Name) == "" {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "name",
				Message: "item name is required",
			})
		}

		if item.Price < 0 {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "price",
				Message: "item price cannot be negative",
				Value:   item.Price,
			})
		}

		if item.TotalPrice < 0 {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "total_price",
				Message: "item total price cannot be negative",
				Value:   item.TotalPrice,
			})
		}

		if item.Price > 0 && item.TotalPrice > 0 && item.TotalPrice > item.Price {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "price_consistency",
				Message: "total price cannot be greater than unit price",
				Value: map[string]int{
					"price":       item.Price,
					"total_price": item.TotalPrice,
				},
			})
		}

		if item.Sale < 0 || item.Sale > 100 {
			errors = append(errors, ValidationError{
				Field:   itemPrefix + "sale",
				Message: "sale must be between 0 and 100",
				Value:   item.Sale,
			})
		}
	}

	return errors
}

func (v *OrderValidator) validateDates(order *models.Order) []ValidationError {
	var errors []ValidationError

	now := time.Now()

	if order.DateCreated.IsZero() {
		errors = append(errors, ValidationError{
			Field:   "date_created",
			Message: "date created is required",
		})
	} else {
		if order.DateCreated.After(now.Add(v.config.MaxDateFuture)) {
			errors = append(errors, ValidationError{
				Field:   "date_created",
				Message: "date created is too far in the future",
				Value:   order.DateCreated.Format(time.RFC3339),
			})
		}

		if order.DateCreated.Before(now.Add(-v.config.MaxDatePast)) {
			errors = append(errors, ValidationError{
				Field:   "date_created",
				Message: "date created is too far in the past",
				Value:   order.DateCreated.Format(time.RFC3339),
			})
		}
	}

	return errors
}


func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

func isValidPhone(phone string) bool {
	phoneRegex := `^[\+]?[0-9\s\-\(\)]{10,15}$`
	matched, _ := regexp.MatchString(phoneRegex, phone)
	return matched
}

func ValidateOrder(order *models.Order) ValidationResult {
	validator := NewOrderValidator(DefaultConfig())
	return validator.Validate(order)
}