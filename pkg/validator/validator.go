package validator

import (
	"L0/models"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func ValidateOrder(order *models.Order) error {

	var errors []string

	// Проверка обязательных полей верхнего уровня
	if order.OrderUID == "" {
		errors = append(errors, "order_uid is required")
	}
	if order.TrackNumber == "" {
		errors = append(errors, "track_number is required")
	}
	if order.Entry == "" {
		errors = append(errors, "entry is required")
	}

	// Проверка вложенных структур
	if err := validateDelivery(&order.Delivery); err != nil {
		errors = append(errors, fmt.Sprintf("delivery: %v", err))
	}
	if err := validatePayment(&order.Payment, order.OrderUID); err != nil { // order_uid должен ли совпадать??
		errors = append(errors, fmt.Sprintf("payment: %v", err))
	}
	if err := validateItems(order.Items); err != nil { // track_number должен совпадать?
		errors = append(errors, fmt.Sprintf("items: %v", err))
	}

	// Проверка даты
	if order.DateCreated.IsZero() {
		errors = append(errors, "date_created is required")
	}
	if order.DateCreated.After(time.Now()) {
		errors = append(errors, "date_created cannot be in the future")
	}

	// Проверка числовых полей
	if order.SmID < 0 {
		errors = append(errors, "sm_id cannot be negative")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

func validateDelivery(delivery *models.Delivery) error {
	var errors []string

	if delivery.Name == "" {
		errors = append(errors, "name is required")
	}
	if delivery.Phone == "" {
		errors = append(errors, "phone is required")
	} else if !isValidPhone(delivery.Phone) {
		errors = append(errors, "invalid phone format")
	}
	if delivery.Email != "" {
		if _, err := mail.ParseAddress(delivery.Email); err != nil {
			errors = append(errors, "invalid email format")
		}
	}
	if delivery.Address == "" {
		errors = append(errors, "address is required")
	}
	if delivery.City == "" {
		errors = append(errors, "city is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

func validatePayment(payment *models.Payment, orderUID string) error {
	var errors []string

	if payment.Transaction == "" {
		errors = append(errors, "transaction is required")
	} else if payment.Transaction != orderUID {
		errors = append(errors, "transaction must match order_uid")
	}
	if payment.Currency == "" {
		errors = append(errors, "currency is required")
	}
	if payment.Provider == "" {
		errors = append(errors, "provider is required")
	}
	if payment.Amount <= 0 {
		errors = append(errors, "amount must be positive")
	}
	if payment.DeliveryCost < 0 {
		errors = append(errors, "delivery_cost cannot be negative")
	}
	if payment.GoodsTotal < 0 {
		errors = append(errors, "goods_total cannot be negative")
	}
	if payment.CustomFee < 0 {
		errors = append(errors, "custom_fee cannot be negative")
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

func validateItems(items []models.Item) error {
	if len(items) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	var errors []string
	for i, item := range items {
		if item.ChrtID <= 0 {
			errors = append(errors, fmt.Sprintf("item[%d]: chrt_id must be positive", i))
		}
		if item.Price <= 0 {
			errors = append(errors, fmt.Sprintf("item[%d]: price must be positive", i))
		}
		if item.Name == "" {
			errors = append(errors, fmt.Sprintf("item[%d]: name is required", i))
		}
		if item.TotalPrice <= 0 {
			errors = append(errors, fmt.Sprintf("item[%d]: total_price must be positive", i))
		}
		if item.Sale < 0 {
			errors = append(errors, fmt.Sprintf("item[%d]: sale cannot be negative", i))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

func isValidPhone(phone string) bool {
	// Простая проверка формата телефона
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}
