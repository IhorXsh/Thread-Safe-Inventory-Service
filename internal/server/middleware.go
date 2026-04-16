package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const requestIDHeader = "X-Request-ID"

type requestContextKey string

const requestIDContextKey requestContextKey = "request_id"

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func (s *Server) observabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := resolveRequestID(r.Header.Get(requestIDHeader))
		w.Header().Set(requestIDHeader, requestID)

		reqLogger := s.logger.With("request_id", requestID)
		reqWithContext := r.WithContext(withRequestID(r.Context(), requestID))

		startedAt := timeNow()
		span, reqWithTrace := s.startSpan(reqWithContext)
		defer span.End()
		span.SetAttributes(attribute.String("http.request_id", requestID))

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, reqWithTrace)

		durationMs := time.Since(startedAt).Milliseconds()
		s.observeRequest(reqWithTrace, rec.statusCode, startedAt)
		if rec.statusCode >= http.StatusBadRequest {
			span.RecordError(fmt.Errorf("http status %d", rec.statusCode))
			reqLogger.Warn("request completed with error", "method", reqWithTrace.Method, "path", routeTemplate(reqWithTrace), "status", rec.statusCode, "duration_ms", durationMs)
		} else {
			reqLogger.Info("request completed", "method", reqWithTrace.Method, "path", routeTemplate(reqWithTrace), "status", rec.statusCode, "duration_ms", durationMs)
		}
		span.SetAttributes(attribute.Int("http.status_code", rec.statusCode))
	})
}

func (s *Server) startSpan(r *http.Request) (trace.Span, *http.Request) {
	path := routeTemplate(r)
	ctx, span := s.tracer.Start(r.Context(), r.Method+" "+path,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", path),
		),
	)
	return span, r.WithContext(ctx)
}

func routeTemplate(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return "unknown"
	}
	path, err := route.GetPathTemplate()
	if err != nil {
		return "unknown"
	}
	return path
}

func pathVar(r *http.Request, key string) string {
	return mux.Vars(r)[key]
}

func resolveRequestID(incoming string) string {
	if incoming == "" {
		return uuid.NewString()
	}
	if _, err := uuid.Parse(incoming); err != nil {
		return uuid.NewString()
	}
	return incoming
}

func withRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func requestIDFromContext(ctx context.Context) string {
	v := ctx.Value(requestIDContextKey)
	requestID, ok := v.(string)
	if !ok {
		return ""
	}
	return requestID
}

func (s *Server) requestLogger(r *http.Request) *slog.Logger {
	requestID := requestIDFromContext(r.Context())
	if requestID == "" {
		return s.logger
	}
	return s.logger.With("request_id", requestID)
}

var timeNow = func() time.Time {
	return time.Now()
}
