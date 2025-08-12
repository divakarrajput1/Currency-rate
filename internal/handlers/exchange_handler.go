package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/services"
)

type ExchangeHandler struct {
	exchangeService *services.ExchangeService
}

func NewExchangeHandler(exchangeService *services.ExchangeService) *ExchangeHandler {
	return &ExchangeHandler{
		exchangeService: exchangeService,
	}
}

func (h *ExchangeHandler) ConvertCurrency(c *gin.Context) {
	var req models.ConversionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.exchangeService.ConvertCurrency(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Conversion failed",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /convert?from=USD&to=INR&amount=100&date=2025-01-01
func (h *ExchangeHandler) ConvertCurrencyQuery(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	amountStr := c.Query("amount")
	date := c.Query("date")

	if from == "" || to == "" || amountStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing required parameters",
			Message: "from, to, and amount parameters are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid amount",
			Message: "amount must be a valid number",
			Code:    http.StatusBadRequest,
		})
		return
	}

	req := models.ConversionRequest{
		From:   from,
		To:     to,
		Amount: amount,
		Date:   date,
	}

	result, err := h.exchangeService.ConvertCurrency(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Conversion failed",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /rates/latest?from=USD&to=INR
func (h *ExchangeHandler) GetLatestRate(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing required parameters",
			Message: "from and to parameters are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	rate, err := h.exchangeService.GetLatestRate(from, to)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to get exchange rate",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from": from,
		"to":   to,
		"rate": rate,
	})
}

// POST /rates/historical
func (h *ExchangeHandler) GetHistoricalRates(c *gin.Context) {
	var req models.HistoricalRateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	result, err := h.exchangeService.GetHistoricalRates(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to get historical rates",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /rates/historical?from=USD&to=INR&start_date=2025-01-01&end_date=2025-01-07
func (h *ExchangeHandler) GetHistoricalRatesQuery(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if from == "" || to == "" || startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Missing required parameters",
			Message: "from, to, start_date, and end_date parameters are required",
			Code:    http.StatusBadRequest,
		})
		return
	}

	req := models.HistoricalRateRequest{
		From:      from,
		To:        to,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := h.exchangeService.GetHistoricalRates(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to get historical rates",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GET /currencies
func (h *ExchangeHandler) GetSupportedCurrencies(c *gin.Context) {
	currencies := h.exchangeService.GetSupportedCurrencies()
	c.JSON(http.StatusOK, gin.H{
		"currencies": currencies,
	})
}

// GET /health
func (h *ExchangeHandler) GetHealth(c *gin.Context) {
	health := h.exchangeService.GetServiceHealth()
	c.JSON(http.StatusOK, health)
}

// GET /stats/cache
func (h *ExchangeHandler) GetCacheStats(c *gin.Context) {
	stats := h.exchangeService.GetCacheStats()
	c.JSON(http.StatusOK, stats)
}
