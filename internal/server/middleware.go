package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
		startedAt := timeNow()
		span, reqWithTrace := s.startSpan(r)
		defer span.End()

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, reqWithTrace)

		s.observeRequest(reqWithTrace, rec.statusCode, startedAt)
		if rec.statusCode >= http.StatusBadRequest {
			span.RecordError(fmt.Errorf("http status %d", rec.statusCode))
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

var timeNow = func() time.Time {
	return time.Now()
}
