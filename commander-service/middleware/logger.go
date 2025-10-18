package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// ZeroLogger is a middleware that logs HTTP requests using zerolog
func ZeroLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		defer func() {
			duration := time.Since(start)
			requestID := middleware.GetReqID(r.Context())

			// Fully structured approach
			log.Info().
				Str("requestID", requestID).
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("proto", r.Proto).
				Str("remoteAddr", r.RemoteAddr).
				Int("status", ww.Status()).
				Int("size", ww.BytesWritten()).
				Dur("duration", duration).
				Msg("request completed")
		}()

		next.ServeHTTP(ww, r)
	})
}
