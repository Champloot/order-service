package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"context"

	"order-service/internal/cache"
	"order-service/internal/database"
	"order-service/internal/models"
)

type Server struct {
	cache *cache.RedisCache
	db    *database.PostgresRepository
}

func NewServer(cache *cache.RedisCache, db *database.PostgresRepository) *Server {
	return &Server{
		cache: cache,
		db:    db,
	}
}

func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/order/", s.getOrderHandler)
	mux.HandleFunc("/api/benchmark", s.benchmarkHandler) // Новый эндпоинт для бенчмарка

	// Serve static files
	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/", fs)

	log.Printf("Starting HTTP server on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract order_uid from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	orderUID := pathParts[3]
	if orderUID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching order: %s", orderUID)

	var order *models.Order
	var err error
	var source string
	var duration time.Duration

	// Try cache first
	cacheStart := time.Now()
	order, err = s.cache.GetOrder(r.Context(), orderUID)
	cacheDuration := time.Since(cacheStart)

	if err != nil {
		log.Printf("Error accessing cache: %v", err)
	} else if order != nil {
		source = "cache"
		duration = cacheDuration
	} else {
		// If not in cache, try database
		log.Printf("Order %s not found in cache, checking database", orderUID)
		dbStart := time.Now()
		order, err = s.db.GetOrder(r.Context(), orderUID)
		dbDuration := time.Since(dbStart)

		if err != nil {
			log.Printf("Error retrieving order from database: %v", err)
			http.Error(w, "Error retrieving order", http.StatusInternalServerError)
			return
		}

		if order == nil {
			log.Printf("Order %s not found in database", orderUID)
			http.Error(w, "Order not found", http.StatusNotFound)
			return
		}

		source = "database"
		duration = dbDuration

		// Save to cache for future requests
		go func() {
			if err := s.cache.SetOrder(context.Background(), order); err != nil {
				log.Printf("Failed to set order in cache: %v", err)
			}
		}()
	}

	totalDuration := time.Since(start)

	// Add info about data source and time
	response := map[string]interface{}{
		"order":    order,
		"source":   source,
		"timing": map[string]interface{}{
			"total":    totalDuration.String(),
			"fetch":    duration.String(),
			"source":   source,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Order %s fetched from %s in %s (fetch: %s)", orderUID, source, totalDuration.String(), duration.String())
}

// New benchmark handler
func (s *Server) benchmarkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// orders ID for test (5 in cache, 5 in db)
	cachedOrderIDs := []string{"test-order-1", "test-order-2", "test-order-3", "test-order-4", "test-order-5"}
	dbOnlyOrderIDs := []string{"test-order-6", "test-order-7", "test-order-8", "test-order-9", "test-order-10"}

	// Preload first 5
	ctx := context.Background()
	for _, id := range cachedOrderIDs {
		if order, err := s.db.GetOrder(ctx, id); err == nil && order != nil {
			s.cache.SetOrder(ctx, order)
		}
	}

	results := make(map[string]interface{})
	totalCacheTime := time.Duration(0)
	totalDBTime := time.Duration(0)

	// test cache orders
	for _, id := range cachedOrderIDs {
		start := time.Now()
		order, err := s.cache.GetOrder(ctx, id)
		duration := time.Since(start)

		if err == nil && order != nil {
			totalCacheTime += duration
			results[id] = map[string]interface{}{
				"source":   "cache",
				"duration": duration.String(),
				"success":  true,
			}
		}
	}

	// test db orders
	for _, id := range dbOnlyOrderIDs {
		start := time.Now()
		order, err := s.db.GetOrder(ctx, id)
		duration := time.Since(start)

		if err == nil && order != nil {
			totalDBTime += duration
			results[id] = map[string]interface{}{
				"source":   "database",
				"duration": duration.String(),
				"success":  true,
			}
		}
	}

	// calc mid time
	avgCacheTime := totalCacheTime / time.Duration(len(cachedOrderIDs))
	avgDBTime := totalDBTime / time.Duration(len(dbOnlyOrderIDs))

	response := map[string]interface{}{
		"results": results,
		"summary": map[string]interface{}{
			"cache_requests": len(cachedOrderIDs),
			"db_requests":    len(dbOnlyOrderIDs),
			"avg_cache_time": avgCacheTime.String(),
			"avg_db_time":    avgDBTime.String(),
			"speed_ratio":    float64(avgDBTime.Nanoseconds()) / float64(avgCacheTime.Nanoseconds()),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Benchmark completed: Cache avg=%s, DB avg=%s, Ratio=%.2fx faster",
		avgCacheTime.String(), avgDBTime.String(),
		float64(avgDBTime.Nanoseconds())/float64(avgCacheTime.Nanoseconds()))
}
