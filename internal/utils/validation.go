package utils

import (
	"fmt"
	"time"

	"exchange-rate-service/internal/models"
)

const (
	DateFormat      = "2006-01-02"
	MaxLookbackDays = 90
)

// ValidateCurrency checks if a currency is supported
func ValidateCurrency(currency string) error {
	if !models.SupportedCurrencies[currency] {
		return fmt.Errorf("unsupported currency: %s. Supported currencies: USD, INR, EUR, JPY, GBP", currency)
	}
	return nil
}

// ValidateCurrencyPair checks if both currencies in a pair are supported
func ValidateCurrencyPair(from, to string) error {
	if err := ValidateCurrency(from); err != nil {
		return fmt.Errorf("invalid 'from' currency: %w", err)
	}
	if err := ValidateCurrency(to); err != nil {
		return fmt.Errorf("invalid 'to' currency: %w", err)
	}
	return nil
}

// ValidateDate validates date format and checks if it's within the allowed range
func ValidateDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}

	// Parse the date
	parsedDate, err := time.Parse(DateFormat, dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format. Expected YYYY-MM-DD, got: %s", dateStr)
	}

	// Check if date is in the future
	now := time.Now()
	if parsedDate.After(now) {
		return time.Time{}, fmt.Errorf("date cannot be in the future: %s", dateStr)
	}

	// Check if date is beyond the maximum lookback period
	maxLookbackDate := now.AddDate(0, 0, -MaxLookbackDays)
	if parsedDate.Before(maxLookbackDate) {
		return time.Time{}, fmt.Errorf("date is beyond the maximum lookback period of %d days. Earliest allowed date: %s",
			MaxLookbackDays, maxLookbackDate.Format(DateFormat))
	}

	return parsedDate, nil
}

// ValidateDateRange validates a date range for historical data requests
func ValidateDateRange(startDateStr, endDateStr string) (time.Time, time.Time, error) {
	// Validate start date
	startDate, err := ValidateDate(startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start date: %w", err)
	}

	// Validate end date
	endDate, err := ValidateDate(endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end date: %w", err)
	}

	// Check if start date is after end date
	if startDate.After(endDate) {
		return time.Time{}, time.Time{}, fmt.Errorf("start date (%s) cannot be after end date (%s)",
			startDateStr, endDateStr)
	}

	// Check if the date range is reasonable (not more than 90 days)
	if endDate.Sub(startDate) > time.Duration(MaxLookbackDays)*24*time.Hour {
		return time.Time{}, time.Time{}, fmt.Errorf("date range cannot exceed %d days", MaxLookbackDays)
	}

	return startDate, endDate, nil
}

// ValidateAmount checks if the amount is valid for conversion
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0, got: %f", amount)
	}
	if amount > 1e15 { // Reasonable upper limit
		return fmt.Errorf("amount too large: %f", amount)
	}
	return nil
}

// FormatDate formats a time.Time to string in the required format
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// ParseDateSafe safely parses a date string, returning current time if empty
func ParseDateSafe(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Now(), nil
	}
	return time.Parse(DateFormat, dateStr)
}

// IsValidDateString checks if a string is a valid date without full validation
func IsValidDateString(dateStr string) bool {
	if dateStr == "" {
		return true
	}
	_, err := time.Parse(DateFormat, dateStr)
	return err == nil
}

// GetDateRangeList returns a slice of date strings between start and end dates
func GetDateRangeList(startDate, endDate time.Time) []string {
	var dates []string
	current := startDate

	for !current.After(endDate) {
		dates = append(dates, FormatDate(current))
		current = current.AddDate(0, 0, 1)
	}

	return dates
}

// ValidateConversionRequest validates a complete conversion request
func ValidateConversionRequest(req *models.ConversionRequest) error {
	// Validate currency pair
	if err := ValidateCurrencyPair(req.From, req.To); err != nil {
		return err
	}

	// Validate amount
	if err := ValidateAmount(req.Amount); err != nil {
		return err
	}

	// Validate date if provided
	if req.Date != "" {
		_, err := ValidateDate(req.Date)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateHistoricalRequest validates a historical rate request
func ValidateHistoricalRequest(req *models.HistoricalRateRequest) error {
	// Validate currency pair
	if err := ValidateCurrencyPair(req.From, req.To); err != nil {
		return err
	}

	// Validate date range
	_, _, err := ValidateDateRange(req.StartDate, req.EndDate)
	return err
}
