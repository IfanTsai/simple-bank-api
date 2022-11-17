package gapi

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = b
	size, err := w.ResponseWriter.Write(b)

	return size, errors.Wrap(err, "failed to write response")
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func HTTPLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			writer := newResponseWriter(w)
			next.ServeHTTP(writer, r)

			end := time.Now()
			elapsed := end.Sub(start)

			fields := []zap.Field{
				zap.Int("status_code", writer.statusCode),
				zap.String("status", http.StatusText(writer.statusCode)),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("elapsed", elapsed),
			}

			if writer.statusCode != http.StatusOK {
				logger.With(fields...).Error("HTTP error", zap.ByteString("body", writer.body))
			} else {
				logger.With(fields...).Info("HTTP success")
			}
		})
	}
}

func GRPCLogger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		res, err := handler(ctx, req)

		end := time.Now()
		elapsed := end.Sub(start)

		statusCode := codes.Unknown
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		}

		fields := []zap.Field{
			zap.Int("status_code", int(statusCode)),
			zap.String("status", statusCode.String()),
			zap.String("method", info.FullMethod),
			zap.Duration("elapsed", elapsed),
		}

		if err != nil {
			logger.With(fields...).Error("gRPC error", zap.Error(err))
		} else {
			logger.With(fields...).Info("gRPC success")
		}

		return res, err
	}
}
