package external

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"exchange-rate-service/internal/models"
)

const (
	BaseURL         = "https://api.exchangerate-api.com/v4"
	LatestEndpoint  = "/latest"
	HistoryEndpoint = "/history"
	RequestTimeout  = 10 * time.Second
)

type ExchangeRateClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewExchangeRateClient() *ExchangeRateClient {
	return &ExchangeRateClient{
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
		baseURL: BaseURL,
	}
}

func (c *ExchangeRateClient) GetLatestRates(baseCurrency string) (*models.ExternalAPIResponse, error) {
	url := fmt.Sprintf("%s%s/%s", c.baseURL, LatestEndpoint, baseCurrency)
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var apiResponse models.ExternalAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResponse.Rates == nil {
		return nil, fmt.Errorf("API request was not successful - no rates received")
	}

	return &apiResponse, nil
}

func (c *ExchangeRateClient) GetHistoricalRates(baseCurrency, date string) (*models.ExternalAPIResponse, error) {

	return nil, fmt.Errorf("historical data not available with current API - upgrade to paid tier for historical data")
}

func (c *ExchangeRateClient) GetRateForPair(from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	// Get latest rates with 'from' currency as base
	apiResponse, err := c.GetLatestRates(from)
	if err != nil {
		return 0, err
	}

	rate, exists := apiResponse.Rates[to]
	if !exists {
		return 0, fmt.Errorf("rate not found for currency pair %s/%s", from, to)
	}

	return rate, nil
}

func (c *ExchangeRateClient) GetHistoricalRateForPair(from, to, date string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	apiResponse, err := c.GetHistoricalRates(from, date)
	if err != nil {
		return 0, err
	}

	rate, exists := apiResponse.Rates[to]
	if !exists {
		return 0, fmt.Errorf("historical rate not found for currency pair %s/%s on %s", from, to, date)
	}

	return rate, nil
}
