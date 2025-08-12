package models

import (
	"time"
)

// ExchangeRate represents a single exchange rate between two currencies
type ExchangeRate struct {
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	Rate         float64   `json:"rate"`
	Date         time.Time `json:"date"`
	LastUpdated  time.Time `json:"last_updated"`
}

// ConversionRequest represents a request to convert currency
type ConversionRequest struct {
	From   string  `json:"from" binding:"required"`
	To     string  `json:"to" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Date   string  `json:"date,omitempty"` // Optional, format: YYYY-MM-DD
}

// ConversionResponse represents the response for currency conversion
type ConversionResponse struct {
	From            string    `json:"from"`
	To              string    `json:"to"`
	Amount          float64   `json:"amount"`
	ConvertedAmount float64   `json:"converted_amount"`
	Rate            float64   `json:"rate"`
	Date            time.Time `json:"date"`
}

// HistoricalRateRequest represents a request for historical rates
type HistoricalRateRequest struct {
	From      string `json:"from" binding:"required"`
	To        string `json:"to" binding:"required"`
	StartDate string `json:"start_date" binding:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" binding:"required"`   // YYYY-MM-DD
}

// HistoricalRateResponse represents historical rate data
type HistoricalRateResponse struct {
	From  string                    `json:"from"`
	To    string                    `json:"to"`
	Rates map[string]HistoricalRate `json:"rates"` // date -> rate
}

// HistoricalRate represents a rate for a specific date
type HistoricalRate struct {
	Rate float64   `json:"rate"`
	Date time.Time `json:"date"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ExternalAPIResponse represents the response from external exchange rate API
type ExternalAPIResponse struct {
	Provider        string             `json:"provider"`
	Base            string             `json:"base"`
	Date            string             `json:"date"`
	TimeLastUpdated int64              `json:"time_last_updated"`
	Rates           map[string]float64 `json:"rates"`
}

// SupportedCurrencies lists all supported currencies
var SupportedCurrencies = map[string]bool{
	"USD": true, // United States Dollar
	"INR": true, // Indian Rupee
	"EUR": true, // Euro
	"JPY": true, // Japanese Yen
	"GBP": true, // British Pound Sterling
}
