package services

import (
	"fmt"
	"time"

	"exchange-rate-service/internal/cache"
	"exchange-rate-service/internal/external"
	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/utils"
)

type ExchangeService struct {
	cache       cache.CacheInterface
	rateFetcher *RateFetcher
	client      *external.ExchangeRateClient
}

func NewExchangeService(cache cache.CacheInterface, rateFetcher *RateFetcher, client *external.ExchangeRateClient) *ExchangeService {
	return &ExchangeService{
		cache:       cache,
		rateFetcher: rateFetcher,
		client:      client,
	}
}

func (s *ExchangeService) ConvertCurrency(req *models.ConversionRequest) (*models.ConversionResponse, error) {
	if err := utils.ValidateConversionRequest(req); err != nil {
		return nil, err
	}

	var conversionDate time.Time
	var err error
	if req.Date != "" {
		conversionDate, err = utils.ValidateDate(req.Date)
		if err != nil {
			return nil, err
		}
	} else {
		conversionDate = time.Now()
	}

	var rate float64
	if req.Date != "" {
		rate, err = s.getHistoricalRate(req.From, req.To, req.Date)
	} else {
		rate, err = s.getLatestRate(req.From, req.To)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	convertedAmount := req.Amount * rate

	return &models.ConversionResponse{
		From:            req.From,
		To:              req.To,
		Amount:          req.Amount,
		ConvertedAmount: convertedAmount,
		Rate:            rate,
		Date:            conversionDate,
	}, nil
}

func (s *ExchangeService) GetLatestRate(from, to string) (float64, error) {
	if err := utils.ValidateCurrencyPair(from, to); err != nil {
		return 0, err
	}

	return s.getLatestRate(from, to)
}

func (s *ExchangeService) GetHistoricalRates(req *models.HistoricalRateRequest) (*models.HistoricalRateResponse, error) {
	if err := utils.ValidateHistoricalRequest(req); err != nil {
		return nil, err
	}

	startDate, endDate, err := utils.ValidateDateRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	dates := utils.GetDateRangeList(startDate, endDate)
	rates := make(map[string]models.HistoricalRate)

	for _, dateStr := range dates {
		rate, err := s.getHistoricalRate(req.From, req.To, dateStr)
		if err != nil {
			continue
		}

		parsedDate, _ := time.Parse(utils.DateFormat, dateStr)
		rates[dateStr] = models.HistoricalRate{
			Rate: rate,
			Date: parsedDate,
		}
	}

	return &models.HistoricalRateResponse{
		From:  req.From,
		To:    req.To,
		Rates: rates,
	}, nil
}

func (s *ExchangeService) getLatestRate(from, to string) (float64, error) {
	// Same currency
	if from == to {
		return 1.0, nil
	}

	if rate, found := s.cache.Get(from, to, ""); found {
		return rate, nil
	}

	rate, err := s.rateFetcher.FetchRateOnDemand(from, to)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rate from API: %w", err)
	}

	return rate, nil
}

func (s *ExchangeService) getHistoricalRate(from, to, date string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	if rate, found := s.cache.Get(from, to, date); found {
		return rate, nil
	}

	rate, err := s.rateFetcher.FetchHistoricalRateOnDemand(from, to, date)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch historical rate from API: %w", err)
	}

	return rate, nil
}

func (s *ExchangeService) GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(models.SupportedCurrencies))
	for currency := range models.SupportedCurrencies {
		currencies = append(currencies, currency)
	}
	return currencies
}

func (s *ExchangeService) GetCacheStats() map[string]interface{} {
	return s.rateFetcher.GetCacheStats()
}

func (s *ExchangeService) GetServiceHealth() map[string]interface{} {
	return map[string]interface{}{
		"status":               "healthy",
		"rate_fetcher":         s.rateFetcher.IsRunning(),
		"supported_currencies": s.GetSupportedCurrencies(),
		"cache_stats":          s.GetCacheStats(),
		"timestamp":            time.Now().Format(time.RFC3339),
	}
}
