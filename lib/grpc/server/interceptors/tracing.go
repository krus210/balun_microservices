package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	traceIDKey = "x-trace-id"
)

// DebugOpenTelemetryUnaryServerInterceptor - OpenTelemetry interceptor для логирования запросов/ответов
func DebugOpenTelemetryUnaryServerInterceptor(logRequest, logResponse bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		tracer := otel.Tracer("grpc-server")

		// Создаем или получаем span
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		// Добавляем trace ID в metadata
		if span.SpanContext().HasTraceID() {
			traceID := span.SpanContext().TraceID().String()
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceIDKey, traceID))

			header := metadata.New(map[string]string{traceIDKey: traceID})
			err := grpc.SendHeader(ctx, header)
			if err != nil {
				return nil, err
			}
		}

		// Логируем запрос если нужно
		if pbMsg, ok := req.(proto.Message); ok && logRequest {
			if jsonRequest, err := protojson.Marshal(pbMsg); err == nil {
				span.AddEvent("grpc_request", trace.WithAttributes(
					attribute.String("request", string(jsonRequest)),
				))
			}
		}

		// Вызываем handler
		res, err := handler(ctx, req)

		// Обрабатываем ошибку
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			span.SetAttributes(
				attribute.Int("grpc.status_code", int(status.Code(err))),
			)
		} else {
			span.SetStatus(codes.Ok, "")
			// Логируем ответ если нужно
			if pbMsg, ok := res.(proto.Message); ok && logResponse {
				if jsonResponse, err := protojson.Marshal(pbMsg); err == nil {
					span.AddEvent("grpc_response", trace.WithAttributes(
						attribute.String("response", string(jsonResponse)),
					))
				}
			}
		}

		return res, err
	}
}
