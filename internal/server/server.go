package server

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/IhorXsh/Thread-Safe-Inventory-Service/internal/inventory"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	svc    inventory.InventoryService
	logger *slog.Logger
	tracer trace.Tracer

	reqDuration *prometheus.HistogramVec
	reqTotal    *prometheus.CounterVec
	reqErrors   *prometheus.CounterVec

	router *mux.Router
}

func New(svc inventory.InventoryService, logger *slog.Logger, tracer trace.Tracer) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	reqDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "inventory_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	reqTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	reqErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "inventory_http_request_errors_total",
			Help: "Total HTTP requests completed with error status (>=400).",
		},
		[]string{"method", "path"},
	)

	registerCollector(reqDuration)
	registerCollector(reqTotal)
	registerCollector(reqErrors)

	s := &Server{
		svc:         svc,
		logger:      logger,
		tracer:      tracer,
		reqDuration: reqDuration,
		reqTotal:    reqTotal,
		reqErrors:   reqErrors,
		router:      mux.NewRouter(),
	}

	s.router.Use(s.observabilityMiddleware)
	s.router.HandleFunc("/healthz", s.handleHealthz).Methods(http.MethodGet)
	s.router.HandleFunc("/stock/{productID}", s.handleGetStock).Methods(http.MethodGet)
	s.router.HandleFunc("/reserve", s.handleReserve).Methods(http.MethodPost)
	s.router.HandleFunc("/reserve-multiple", s.handleReserveMultiple).Methods(http.MethodPost)

	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) observeRequest(r *http.Request, statusCode int, startedAt time.Time) {
	path := routeTemplate(r)
	status := strconv.Itoa(statusCode)

	s.reqDuration.WithLabelValues(r.Method, path, status).Observe(time.Since(startedAt).Seconds())
	s.reqTotal.WithLabelValues(r.Method, path, status).Inc()

	if statusCode >= http.StatusBadRequest {
		s.reqErrors.WithLabelValues(r.Method, path).Inc()
	}
}

func registerCollector(collector prometheus.Collector) {
	if err := prometheus.Register(collector); err != nil {
		if _, ok := err.(prometheus.AlreadyRegisteredError); ok {
			return
		}
		panic(err)
	}
}
