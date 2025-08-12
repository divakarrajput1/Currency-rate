# Exchange Rate Service

A high-performance, scalable currency exchange rate service built in Go that provides real-time and historical exchange rates with intelligent caching.

## Features

- **Real-time Exchange Rates**: Fetches latest rates every hour from external APIs
- **Historical Data**: Supports historical exchange rates up to 90 days back
- **Currency Conversion**: Convert amounts between supported currencies
- **Smart Caching**: In-memory caching with TTL to reduce API calls
- **Thread-Safe**: Handles concurrent requests gracefully
- **Error Handling**: Robust error handling for API failures
- **Health Monitoring**: Built-in health checks and statistics
- **Docker Support**: Containerized for easy deployment
- **Clean Architecture**: Well-structured, testable codebase

## Supported Currencies

- **USD** - United States Dollar
- **INR** - Indian Rupee  
- **EUR** - Euro
- **JPY** - Japanese Yen
- **GBP** - British Pound Sterling

## Quick Start

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd exchange-rate-service
   ```

2. **Run with Docker Compose**
   ```bash
   docker-compose up --build
   ```

3. **Test the service**
   ```bash
   curl http://localhost:8080/health
   ```

### Local Development

1. **Prerequisites**
   - Go 1.21 or higher
   - Git

2. **Setup**
   ```bash
   git clone <repository-url>
   cd exchange-rate-service
   go mod download
   ```

3. **Run the service**
   ```bash
   go run cmd/server/main.go
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Endpoints

#### 1. Currency Conversion

**POST /convert**
```bash
curl -X POST http://localhost:8080/api/v1/convert \
  -H "Content-Type: application/json" \
  -d '{
    "from": "USD",
    "to": "INR", 
    "amount": 100
  }'
```

**GET /convert (Query Parameters)**
```bash
curl "http://localhost:8080/api/v1/convert?from=USD&to=INR&amount=100"
```

**Response:**
```json
{
  "from": "USD",
  "to": "INR",
  "amount": 100,
  "converted_amount": 8312.50,
  "rate": 83.125,
  "date": "2025-01-16T10:30:00Z"
}
```

#### 2. Latest Exchange Rates

**GET /rates/latest**
```bash
curl "http://localhost:8080/api/v1/rates/latest?from=USD&to=INR"
```

**Response:**
```json
{
  "from": "USD",
  "to": "INR", 
  "rate": 83.125
}
```

#### 3. Historical Exchange Rates

> **Note**: Historical data requires a paid API subscription. The current free tier implementation returns an error for historical rate requests.

**POST /rates/historical**
```bash
curl -X POST http://localhost:8080/api/v1/rates/historical \
  -H "Content-Type: application/json" \
  -d '{
    "from": "USD",
    "to": "INR",
    "start_date": "2025-01-01", 
    "end_date": "2025-01-07"
  }'
```

**Current Response (Free Tier):**
```json
{
  "error": "Failed to get historical rates",
  "message": "historical data not available with current API - upgrade to paid tier for historical data",
  "code": 400
}
```

#### 4. Historical Conversion

**POST /convert (with date)**
```bash
curl -X POST http://localhost:8080/api/v1/convert \
  -H "Content-Type: application/json" \
  -d '{
    "from": "USD",
    "to": "INR",
    "amount": 100,
    "date": "2025-01-01"
  }'
```

#### 5. Utility Endpoints

**Get Supported Currencies**
```bash
curl http://localhost:8080/api/v1/currencies
```

**Health Check**
```bash
curl http://localhost:8080/health
```

**Cache Statistics**
```bash
curl http://localhost:8080/api/v1/stats/cache
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `GIN_MODE` | `release` | Gin framework mode |

### Cache Configuration

- **TTL**: 1 hour for all cached rates
- **Cleanup**: Expired entries cleaned every 5 minutes
- **Thread-Safe**: Uses RWMutex for concurrent access

### Rate Fetching

- **Interval**: Every 1 hour
- **Source**: exchangerate-api.com API
- **Timeout**: 10 seconds per request
- **Retry**: Handles API failures gracefully

## Architecture

```
┌─────────────────┐    ┌──────────────────┐
│     Client      │    │   External API   │
│                 │    │ (exchangerate.   │
└─────────┬───────┘    │     host)        │
          │            └──────────┬───────┘
          │                       │
          │ HTTP                  │ HTTP
          │                       │
          ▼                       ▼
┌─────────────────────────────────────────┐
│           API Layer                     │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │  Handlers   │  │  Middleware     │   │
│  └─────────────┘  └─────────────────┘   │
└─────────┬───────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────┐
│         Service Layer                   │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Exchange    │  │ Rate Fetcher    │   │
│  │ Service     │  │ Service         │   │
│  └─────────────┘  └─────────────────┘   │
└─────────┬───────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────┐
│         Data Layer                      │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │ Memory      │  │ External API    │   │
│  │ Cache       │  │ Client          │   │
│  └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────┘
```

## Performance Considerations

### Caching Strategy
- **Cache Hit**: O(1) lookup for cached rates
- **Cache Miss**: Falls back to API call
- **Cache Warming**: Proactive fetching every hour
- **Memory Efficient**: TTL-based expiration

### Concurrency
- **Thread-Safe Cache**: RWMutex for concurrent access
- **Parallel API Calls**: Concurrent fetching for multiple currencies
- **Graceful Degradation**: Continues operation during API failures

### Scalability
- **Horizontal Scaling**: Stateless design allows multiple instances
- **Vertical Scaling**: Efficient memory usage and CPU optimization
- **Load Balancing**: Ready for load balancer deployment

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test -cover ./...
```

### Run Specific Test Package
```bash
go test ./internal/utils
go test ./internal/cache
```

### Test Examples

The service includes comprehensive tests for:
- ✅ Input validation (currencies, dates, amounts)
- ✅ Cache operations (thread safety, expiration)
- ✅ Date validation (90-day limit enforcement)
- ✅ Currency conversion logic
- ✅ Concurrent access patterns

## Deployment

### Docker Production Build
```bash
docker build -t exchange-rate-service .
docker run -p 8080:8080 exchange-rate-service
```

### Docker Compose with Monitoring
```bash
# Run with Prometheus and Grafana
docker-compose --profile monitoring up
```

Access points:
- **Service**: http://localhost:8080
- **Prometheus**: http://localhost:9090  
- **Grafana**: http://localhost:3000 (admin/admin)

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exchange-rate-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: exchange-rate-service
  template:
    metadata:
      labels:
        app: exchange-rate-service
    spec:
      containers:
      - name: exchange-rate-service
        image: exchange-rate-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Error Handling

### Common Error Responses

**400 Bad Request**
```json
{
  "error": "Validation failed",
  "message": "unsupported currency: XYZ. Supported currencies: USD, INR, EUR, JPY, GBP",
  "code": 400
}
```

**400 Bad Request - Date Validation**
```json
{
  "error": "Invalid date",
  "message": "date is beyond the maximum lookback period of 90 days",
  "code": 400
}
```

### Error Categories
- **Validation Errors**: Invalid currencies, dates, amounts
- **API Errors**: External service failures
- **Rate Limit**: Too many requests (if implemented)
- **Service Errors**: Internal service failures

## Monitoring and Observability

### Health Check Response
```json
{
  "status": "healthy",
  "rate_fetcher": true,
  "supported_currencies": ["USD", "INR", "EUR", "JPY", "GBP"],
  "cache_stats": {
    "total_items": 25,
    "valid_items": 25,
    "expired_items": 0,
    "ttl_seconds": 3600
  },
  "timestamp": "2025-01-16T10:30:00Z"
}
```

### Metrics Available
- Cache hit/miss ratios
- API response times
- Request volumes
- Error rates
- Service uptime

## Assumptions Made

1. **Rate Source**: Using exchangerate-api.com as primary data source (free, reliable)
2. **Cache Duration**: 1-hour TTL balances freshness with API efficiency
3. **Date Validation**: 90-day lookback limit for performance and data availability
4. **Currency Set**: Fixed set of 5 major currencies for MVP
5. **Error Handling**: Graceful degradation when external APIs fail
6. **Concurrency**: Service designed for high-concurrency read operations
7. **Time Zone**: All dates processed in UTC for consistency

## Future Enhancements

- [ ] **Cryptocurrency Support**: BTC, ETH, USDT conversions
- [ ] **Rate Alerts**: WebSocket notifications for rate changes
- [ ] **Database Persistence**: PostgreSQL for historical data storage
- [ ] **Rate Limiting**: API throttling and quotas
- [ ] **Authentication**: JWT-based API authentication
- [ ] **Metrics Export**: Prometheus metrics endpoint
- [ ] **Multiple Data Sources**: Failover between API providers
- [ ] **Batch Operations**: Convert multiple currency pairs at once

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Built with ❤️ using Go, Gin, Docker, and best practices for production-ready services.** 