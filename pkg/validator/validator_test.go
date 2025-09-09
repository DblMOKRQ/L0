package validator

import (
	"L0/internal/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		phone    string
		expected bool
	}{
		{"+79999999999", true},
		{"79999999999", true},
		{"+712345", false},
		{"not-phone", false},
		{"", false},
		{"+1234567890", true},
		{"+123456789012345", true},
		{"+1234567890123456", false},
		{"+441234567890", true},
		{"+1-800-123-4567", false},
		{"+7 (999) 123-45-67", false},
		{"+7 999 123 45 67", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			result := isValidPhone(tt.phone)
			assert.Equal(t, tt.expected, result, "Phone: %s", tt.phone)
		})
	}
}

func TestValidateItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []models.Item
		wantErr  bool
		errorMsg string
	}{
		{
			name:     "No items",
			items:    []models.Item{},
			wantErr:  true,
			errorMsg: "at least one item is required",
		},
		{
			name: "Negative chrtID",
			items: []models.Item{
				{ChrtID: -1},
			},
			wantErr: true,
		},
		{
			name: "Negative price",
			items: []models.Item{
				{ChrtID: 1, Price: -1},
			},
			wantErr: true,
		},
		{
			name: "No name",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: ""},
			},
			wantErr: true,
		},
		{
			name: "Negative total price",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: -1},
			},
			wantErr: true,
		},
		{
			name: "Negative sale",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: -1},
			},
			wantErr: true,
		},
		{
			name: "No brand",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: 1, Brand: ""},
			},
			wantErr: true,
		},
		{
			name: "Negative status",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: 1, Brand: "test", Status: -1},
			},
			wantErr: true,
		},
		{
			name: "No rid",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: 1, Brand: "test", Status: 1, Rid: ""},
			},
			wantErr: true,
		},
		{
			name: "No trackNumber",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: 1, Brand: "test", Status: 1, Rid: "12313", TrackNumber: ""},
			},
			wantErr: true,
		},
		{
			name: "Negative NmID",
			items: []models.Item{
				{ChrtID: 1, Price: 1, Name: "test", TotalPrice: 1, Sale: 1, Brand: "test", Status: 1, Rid: "12313", TrackNumber: "123123", NmID: -1},
			},
			wantErr: true,
		},
		{
			name: "Correct item",
			items: []models.Item{
				{
					ChrtID:      1,
					TrackNumber: "1223213",
					Price:       10,
					Rid:         "123123",
					Name:        "test",
					Sale:        1,
					TotalPrice:  10, // Исправлено: должно быть >= Price
					NmID:        213213,
					Brand:       "test",
					Status:      1,
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple items with one invalid",
			items: []models.Item{
				{
					ChrtID:      1,
					TrackNumber: "1223213",
					Price:       10,
					Rid:         "123123",
					Name:        "test",
					Sale:        1,
					TotalPrice:  10,
					NmID:        213213,
					Brand:       "test",
					Status:      1,
				},
				{
					ChrtID:      -1, // Невалидный
					TrackNumber: "1223213",
					Price:       10,
					Rid:         "123123",
					Name:        "test",
					Sale:        1,
					TotalPrice:  10,
					NmID:        213213,
					Brand:       "test",
					Status:      1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateItems(tt.items)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePayment(t *testing.T) {
	tests := []struct {
		name    string
		payment models.Payment
		wantErr bool
	}{
		{
			name: "Invalid transaction",
			payment: models.Payment{
				Transaction: "",
			},
			wantErr: true,
		},
		{
			name: "Not currency ",
			payment: models.Payment{
				Transaction: "123",
				Currency:    "",
			},
			wantErr: true,
		},
		{
			name: "No provider",
			payment: models.Payment{
				Transaction: "123",
				Currency:    "us",
				Provider:    "",
			},
			wantErr: true,
		},
		{
			name: "Negative amount",
			payment: models.Payment{
				Transaction: "123",
				Currency:    "us",
				Provider:    "123",
				Amount:      -1,
			},
			wantErr: true,
		},
		{
			name: "Negative delivery cost",
			payment: models.Payment{
				Transaction:  "123",
				Currency:     "us",
				Provider:     "123",
				Amount:       10,
				DeliveryCost: -1,
			},
			wantErr: true,
		},
		{
			name: "Negative goods total",
			payment: models.Payment{
				Transaction:  "123",
				Currency:     "us",
				Provider:     "123",
				Amount:       10,
				DeliveryCost: 10,
				GoodsTotal:   -1,
			},
			wantErr: true,
		},
		{
			name: "Negative custom fee",
			payment: models.Payment{
				Transaction:  "123",
				Currency:     "us",
				Provider:     "123",
				Amount:       10,
				DeliveryCost: 10,
				GoodsTotal:   10,
				CustomFee:    -1,
			},
			wantErr: true,
		},
		{
			name: "Correct payment",
			payment: models.Payment{
				Transaction:  "123",
				Currency:     "us",
				Provider:     "123",
				Amount:       10,
				DeliveryCost: 10,
				GoodsTotal:   10,
				CustomFee:    1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePayment(&tt.payment)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDelivery(t *testing.T) {
	tests := []struct {
		name      string
		delivery  models.Delivery
		wantErr   bool
		errorMsgs []string
	}{
		{
			name:      "Empty delivery",
			delivery:  models.Delivery{},
			wantErr:   true,
			errorMsgs: []string{"name is required", "phone is required", "address is required", "city is required"},
		},
		{
			name: "Invalid phone format",
			delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "invalid-phone",
				Address: "Street 1",
				City:    "Moscow",
			},
			wantErr:   true,
			errorMsgs: []string{"invalid phone format"},
		},
		{
			name: "Invalid email format",
			delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "+79999999999",
				Email:   "invalid-email",
				Address: "Street 1",
				City:    "Moscow",
			},
			wantErr:   true,
			errorMsgs: []string{"invalid email format"},
		},
		{
			name: "Missing required fields",
			delivery: models.Delivery{
				Name:    "",
				Phone:   "",
				Address: "",
				City:    "",
			},
			wantErr:   true,
			errorMsgs: []string{"name is required", "phone is required", "address is required", "city is required"},
		},
		{
			name: "Valid delivery without email",
			delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "+79999999999",
				Address: "Street 1",
				City:    "Moscow",
			},
			wantErr: false,
		},
		{
			name: "Valid delivery with email",
			delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "+79999999999",
				Email:   "test@example.com",
				Address: "Street 1",
				City:    "Moscow",
			},
			wantErr: false,
		},
		{
			name: "Valid international phone",
			delivery: models.Delivery{
				Name:    "Test User",
				Phone:   "+441234567890",
				Address: "Street 1",
				City:    "London",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDelivery(&tt.delivery)

			if tt.wantErr {
				assert.Error(t, err)
				if len(tt.errorMsgs) > 0 {
					for _, msg := range tt.errorMsgs {
						assert.Contains(t, err.Error(), msg)
					}
				}
				t.Logf("Error: %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOrder(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		order   models.Order
		wantErr bool
	}{
		{
			name:    "Empty order",
			order:   models.Order{},
			wantErr: true,
		},
		{
			name: "Missing required fields",
			order: models.Order{
				OrderUID:    "",
				TrackNumber: "",
				Entry:       "",
			},
			wantErr: true,
		},
		{
			name: "Future date",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "Negative SmID",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(-24 * time.Hour),
				SmID:        -1,
			},
			wantErr: true,
		},
		{
			name: "Invalid delivery",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(-24 * time.Hour),
				SmID:        1,
				Delivery:    models.Delivery{Name: ""},
			},
			wantErr: true,
		},
		{
			name: "Invalid payment",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(-24 * time.Hour),
				SmID:        1,
				Delivery: models.Delivery{
					Name:    "Test User",
					Phone:   "+79999999999",
					Address: "Street 1",
					City:    "Moscow",
				},
				Payment: models.Payment{Transaction: ""},
			},
			wantErr: true,
		},
		{
			name: "Invalid items",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(-24 * time.Hour),
				SmID:        1,
				Delivery: models.Delivery{
					Name:    "Test User",
					Phone:   "+79999999999",
					Address: "Street 1",
					City:    "Moscow",
				},
				Payment: models.Payment{
					Transaction:  "test123",
					Currency:     "USD",
					Provider:     "stripe",
					Amount:       100,
					DeliveryCost: 10,
					GoodsTotal:   90,
					CustomFee:    0,
				},
				Items: []models.Item{},
			},
			wantErr: true,
		},
		{
			name: "Valid order",
			order: models.Order{
				OrderUID:    "test123",
				TrackNumber: "TRACK123",
				Entry:       "WBIL",
				DateCreated: now.Add(-24 * time.Hour),
				SmID:        1,
				Delivery: models.Delivery{
					Name:    "Test User",
					Phone:   "+79999999999",
					Email:   "test@example.com",
					Address: "Street 1",
					City:    "Moscow",
				},
				Payment: models.Payment{
					Transaction:  "test123",
					Currency:     "USD",
					Provider:     "stripe",
					Amount:       100,
					DeliveryCost: 10,
					GoodsTotal:   90,
					CustomFee:    0,
				},
				Items: []models.Item{
					{
						ChrtID:      1,
						TrackNumber: "TRACK123",
						Price:       100,
						Rid:         "rid123",
						Name:        "Test Item",
						Sale:        0,
						TotalPrice:  100,
						NmID:        123,
						Brand:       "Test Brand",
						Status:      1,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOrder(&tt.order)

			if tt.wantErr {
				assert.Error(t, err)
				t.Logf("Expected error: %v", err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
