package services

import (
	"context"
	"log"
	"sync"
	"time"

	"exchange-rate-service/internal/cache"
	"exchange-rate-service/internal/external"
	"exchange-rate-service/internal/models"
)

type RateFetcher struct {
	client        *external.ExchangeRateClient
	cache         cache.CacheInterface
	currencies    []string
	fetchInterval time.Duration
	mu            sync.RWMutex
	isRunning     bool
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewRateFetcher(client *external.ExchangeRateClient, cache cache.CacheInterface) *RateFetcher {
	currencies := make([]string, 0, len(models.SupportedCurrencies))
	for currency := range models.SupportedCurrencies {
		currencies = append(currencies, currency)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &RateFetcher{
		client:        client,
		cache:         cache,
		currencies:    currencies,
		fetchInterval: 1 * time.Hour, 
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (rf *RateFetcher) Start() {
	rf.mu.Lock()
	if rf.isRunning {
		rf.mu.Unlock()
		return
	}
	rf.isRunning = true
	rf.mu.Unlock()

	log.Println("Starting rate fetcher service...")

	go rf.fetchAllRates()

	go rf.periodicFetch()
}

func (rf *RateFetcher) Stop() {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if !rf.isRunning {
		return
	}

	log.Println("Stopping rate fetcher service...")
	rf.cancel()
	rf.isRunning = false
}

func (rf *RateFetcher) IsRunning() bool {
	rf.mu.RLock()
	defer rf.mu.RUnlock()
	return rf.isRunning
}

func (rf *RateFetcher) periodicFetch() {
	ticker := time.NewTicker(rf.fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rf.ctx.Done():
			log.Println("Rate fetcher stopped")
			return
		case <-ticker.C:
			rf.fetchAllRates()
		}
	}
}

func (rf *RateFetcher) fetchAllRates() {
	log.Println("Fetching latest exchange rates...")
	start := time.Now()

	var wg sync.WaitGroup
	rateChan := make(chan rateResult, len(rf.currencies)*len(rf.currencies))

	for _, baseCurrency := range rf.currencies {
		wg.Add(1)
		go func(base string) {
			defer wg.Done()
			rf.fetchRatesForBase(base, rateChan)
		}(baseCurrency)
	}

	go func() {
		wg.Wait()
		close(rateChan)
	}()

	successCount := 0
	errorCount := 0

	for result := range rateChan {
		if result.err != nil {
			log.Printf("Error fetching rate %s/%s: %v", result.from, result.to, result.err)
			errorCount++
		} else {
			rf.cache.Set(result.from, result.to, "", result.rate)
			successCount++
		}
	}

	duration := time.Since(start)
	log.Printf("Rate fetch completed in %v. Success: %d, Errors: %d", duration, successCount, errorCount)
}

type rateResult struct {
	from string
	to   string
	rate float64
	err  error
}

func (rf *RateFetcher) fetchRatesForBase(baseCurrency string, resultChan chan<- rateResult) {
	apiResponse, err := rf.client.GetLatestRates(baseCurrency)
	if err != nil {
		for _, toCurrency := range rf.currencies {
			if toCurrency != baseCurrency {
				resultChan <- rateResult{
					from: baseCurrency,
					to:   toCurrency,
					err:  err,
				}
			}
		}
		return
	}

	for toCurrency, rate := range apiResponse.Rates {
		if models.SupportedCurrencies[toCurrency] && toCurrency != baseCurrency {
			resultChan <- rateResult{
				from: baseCurrency,
				to:   toCurrency,
				rate: rate,
			}
		}
	}

	resultChan <- rateResult{
		from: baseCurrency,
		to:   baseCurrency,
		rate: 1.0,
	}
}

func (rf *RateFetcher) FetchRateOnDemand(from, to string) (float64, error) {
	log.Printf("Fetching on-demand rate for %s/%s", from, to)

	rate, err := rf.client.GetRateForPair(from, to)
	if err != nil {
		return 0, err
	}

	rf.cache.Set(from, to, "", rate)

	return rate, nil
}

func (rf *RateFetcher) FetchHistoricalRateOnDemand(from, to, date string) (float64, error) {
	log.Printf("Fetching historical rate for %s/%s on %s", from, to, date)

	rate, err := rf.client.GetHistoricalRateForPair(from, to, date)
	if err != nil {
		return 0, err
	}

	// Cache the fetched historical rate
	rf.cache.Set(from, to, date, rate)

	return rate, nil
}

func (rf *RateFetcher) GetCacheStats() map[string]interface{} {
	return rf.cache.GetStats()
}
