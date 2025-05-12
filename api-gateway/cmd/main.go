package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	failures         int
	lastFailure      time.Time
	state            string
	mutex            sync.RWMutex
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: threshold,
		resetTimeout:     timeout,
		failures:         0,
		state:            "closed",
	}
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = "half-open"
			cb.mutex.Unlock()
			cb.mutex.RLock()
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == "half-open" {
		cb.state = "closed"
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.failureThreshold {
		cb.state = "open"
	}
}

var serviceCircuitBreakers = make(map[string]*CircuitBreaker)

var (
	// Define metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Count of all HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	serviceStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_status",
			Help: "Status of microservices (1=up, 0=down)",
		},
		[]string{"service"},
	)
)

func main() {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			ip := c.IP()
			log.Printf("Request from IP: %s", ip)
			return ip
		},
	}))

	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()
		method := c.Method()
		endpoint := c.Route().Path

		httpRequestsTotal.WithLabelValues(method, endpoint, fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		return err
	})

	initCircuitBreakers()

	setupRoutes(app)

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	log.Fatal(app.Listen("0.0.0.0:8080"))
}

func initCircuitBreakers() {
	services := []string{"user-service", "auth-service", "menu-service", "order-service", "payment-service", "review-service"}
	for _, service := range services {
		serviceCircuitBreakers[service] = NewCircuitBreaker(3, 30*time.Second)
	}
}

func setupRoutes(app *fiber.App) {
	app.Get("/health", healthCheck)

	api := app.Group("/api")

	api.All("/users/*", createServiceProxy("user-service", 8081))
	api.All("/auth/*", createServiceProxy("auth-service", 8082))
	api.All("/menu/*", createServiceProxy("menu-service", 8083))
	api.All("/orders/*", createServiceProxy("order-service", 8084))
	api.All("/payments/*", createServiceProxy("payment-service", 8085))
	api.All("/reviews/*", createServiceProxy("review-service", 8086))
}

func healthCheck(c *fiber.Ctx) error {
	health := map[string]string{"api_gateway": "ok"}

	serviceStatus.WithLabelValues("api_gateway").Set(1)

	services := []string{"user-service", "auth-service", "menu-service", "order-service", "payment-service", "review-service"}
	for _, service := range services {
		serviceURL := os.Getenv(strings.ToUpper(service) + "_URL")
		if serviceURL == "" {
			health[service] = "unknown"
			serviceStatus.WithLabelValues(service).Set(0)
			continue
		}

		client := &http.Client{Timeout: 2 * time.Second}
		_, err := client.Get(serviceURL + "/health")
		if err != nil {
			health[service] = "unavailable"
			serviceStatus.WithLabelValues(service).Set(0)
		} else {
			health[service] = "ok"
			serviceStatus.WithLabelValues(service).Set(1)
		}
	}

	return c.Status(fiber.StatusOK).JSON(health)
}

func createServiceProxy(service string, port int) fiber.Handler {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	cb := serviceCircuitBreakers[service]

	return func(c *fiber.Ctx) error {
		if !cb.AllowRequest() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fmt.Sprintf("%s is temporarily unavailable", service),
			})
		}

		serviceURL := os.Getenv(strings.ToUpper(service) + "_URL")
		if serviceURL == "" {
			serviceURL = fmt.Sprintf("http://%s:%d", service, port)
		}

		originalPath := c.Path()
		var strippedPath string

		switch {
		case strings.HasPrefix(originalPath, "/api/users/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/users")
		case strings.HasPrefix(originalPath, "/api/auth/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/auth")
		case strings.HasPrefix(originalPath, "/api/menu/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/menu")
		case strings.HasPrefix(originalPath, "/api/orders/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/orders")
		case strings.HasPrefix(originalPath, "/api/payments/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/payments")
		case strings.HasPrefix(originalPath, "/api/reviews/"):
			strippedPath = strings.TrimPrefix(originalPath, "/api/reviews")
		default:
			// 3ashan el health checks and other paths
			parts := strings.Split(originalPath, "/")
			if len(parts) >= 3 {
				strippedPath = "/" + strings.Join(parts[3:], "/")
			} else {
				strippedPath = "/"
			}
		}

		if strippedPath == "" || strippedPath[0] != '/' {
			strippedPath = "/" + strippedPath
		}

		// the final target URL
		targetURL := serviceURL + strippedPath

		// debug logging
		log.Printf("Forwarding request: %s %s -> %s", c.Method(), originalPath, targetURL)

		// create a new request
		req, err := http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create request",
				"details": err.Error(),
			})
		}

		// copying headers
		for key, values := range c.GetReqHeaders() {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		// executing the request with retries
		var resp *http.Response
		var retryCount int = 0
		maxRetries := 2

		for retryCount <= maxRetries {
			resp, err = client.Do(req)
			if err == nil {
				break
			}

			retryCount++
			if retryCount <= maxRetries {
				log.Printf("Retry %d for %s: %v", retryCount, service, err)
				time.Sleep(time.Duration(retryCount*200) * time.Millisecond)
				// need to create a new request with body for retry
				req, _ = http.NewRequest(c.Method(), targetURL, bytes.NewReader(c.Body()))
				for key, values := range c.GetReqHeaders() {
					for _, value := range values {
						req.Header.Add(key, value)
					}
				}
			}
		}

		if err != nil {
			// record failure in circuit breaker
			cb.RecordFailure()
			log.Printf("Service %s is unreachable after %d retries: %v", service, retryCount, err)
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "Service temporarily unavailable",
			})
		}
		defer resp.Body.Close()

		// record success in circuit breaker
		cb.RecordSuccess()

		// read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read response",
			})
		}

		// set response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Set(key, value)
			}
		}

		// return response
		return c.Status(resp.StatusCode).Send(body)
	}
}
