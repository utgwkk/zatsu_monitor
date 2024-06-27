package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

type LogHandler struct {
	slog.Handler
}

func (h *LogHandler) Handle(ctx context.Context, r slog.Record) error {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return h.Handler.Handle(ctx, r)
	}

	r.AddAttrs(slog.String("requestId", lc.AwsRequestID))
	return h.Handler.Handle(ctx, r)
}
