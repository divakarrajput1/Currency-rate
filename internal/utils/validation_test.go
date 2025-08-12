package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"exchange-rate-service/internal/models"
)

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		wantErr  bool
	}{
		{"Valid USD", "USD", false},
		{"Valid INR", "INR", false},
		{"Valid EUR", "EUR", false},
		{"Valid JPY", "JPY", false},
		{"Valid GBP", "GBP", false},
		{"Invalid currency", "XYZ", true},
		{"Empty currency", "", true},
		{"Lowercase currency", "usd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrency(tt.currency)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCurrencyPair(t *testing.T) {
	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{"Valid pair USD-INR", "USD", "INR", false},
		{"Valid pair EUR-JPY", "EUR", "JPY", false},
		{"Same currency", "USD", "USD", false},
		{"Invalid from currency", "XYZ", "USD", true},
		{"Invalid to currency", "USD", "XYZ", true},
		{"Both invalid", "XYZ", "ABC", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrencyPair(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDate(t *testing.T) {
	now := time.Now()
	validDate := now.AddDate(0, 0, -30).Format(DateFormat)
	futureDate := now.AddDate(0, 0, 1).Format(DateFormat)
	oldDate := now.AddDate(0, 0, -100).Format(DateFormat)

	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"Empty date", "", false},
		{"Valid recent date", validDate, false},
		{"Future date", futureDate, true},
		{"Too old date", oldDate, true},
		{"Invalid format", "2023-13-01", true},
		{"Invalid format 2", "01-01-2023", true},
		{"Invalid format 3", "2023/01/01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateDate(tt.date)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDateRange(t *testing.T) {
	now := time.Now()
	validStart := now.AddDate(0, 0, -30).Format(DateFormat)
	validEnd := now.AddDate(0, 0, -20).Format(DateFormat)
	futureDate := now.AddDate(0, 0, 1).Format(DateFormat)
	oldDate := now.AddDate(0, 0, -100).Format(DateFormat)

	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantErr   bool
	}{
		{"Valid range", validStart, validEnd, false},
		{"Start after end", validEnd, validStart, true},
		{"Future start date", futureDate, validEnd, true},
		{"Old start date", oldDate, validEnd, true},
		{"Future end date", validStart, futureDate, true},
		{"Range too large", now.AddDate(0, 0, -89).Format(DateFormat), now.AddDate(0, 0, -1).Format(DateFormat), false},
		{"Range exceeds limit", now.AddDate(0, 0, -91).Format(DateFormat), now.AddDate(0, 0, -1).Format(DateFormat), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ValidateDateRange(tt.startDate, tt.endDate)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{"Valid amount", 100.0, false},
		{"Valid small amount", 0.01, false},
		{"Valid large amount", 1000000.0, false},
		{"Zero amount", 0.0, true},
		{"Negative amount", -100.0, true},
		{"Too large amount", 1e16, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConversionRequest(t *testing.T) {
	validDate := time.Now().AddDate(0, 0, -30).Format(DateFormat)

	tests := []struct {
		name    string
		req     *models.ConversionRequest
		wantErr bool
	}{
		{
			"Valid request",
			&models.ConversionRequest{From: "USD", To: "INR", Amount: 100.0},
			false,
		},
		{
			"Valid request with date",
			&models.ConversionRequest{From: "USD", To: "INR", Amount: 100.0, Date: validDate},
			false,
		},
		{
			"Invalid from currency",
			&models.ConversionRequest{From: "XYZ", To: "INR", Amount: 100.0},
			true,
		},
		{
			"Invalid to currency",
			&models.ConversionRequest{From: "USD", To: "XYZ", Amount: 100.0},
			true,
		},
		{
			"Invalid amount",
			&models.ConversionRequest{From: "USD", To: "INR", Amount: -100.0},
			true,
		},
		{
			"Invalid date",
			&models.ConversionRequest{From: "USD", To: "INR", Amount: 100.0, Date: "invalid-date"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConversionRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetDateRangeList(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC)

	dates := GetDateRangeList(start, end)

	expected := []string{
		"2023-01-01",
		"2023-01-02",
		"2023-01-03",
	}

	assert.Equal(t, expected, dates)
}

func TestFormatDate(t *testing.T) {
	date := time.Date(2023, 1, 15, 12, 30, 45, 0, time.UTC)
	formatted := FormatDate(date)
	assert.Equal(t, "2023-01-15", formatted)
}

func TestIsValidDateString(t *testing.T) {
	tests := []struct {
		name   string
		date   string
		expect bool
	}{
		{"Valid date", "2023-01-15", true},
		{"Empty date", "", true},
		{"Invalid format", "01-15-2023", false},
		{"Invalid date", "2023-13-01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidDateString(tt.date)
			assert.Equal(t, tt.expect, result)
		})
	}
}
