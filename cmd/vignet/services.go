package main

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	vignet "github.com/ViBiOh/vignet/pkg/vignet"
)

type services struct {
	server           *server.Server
	streamHandler    *amqphandler.Service
	thumbnailHandler *amqphandler.Service
	vignet           vignet.Service
}

func newServices(config configuration, clients clients, adapters adapters) (services, error) {
	var output services
	var err error

	output.server = server.New(config.server)

	output.vignet = vignet.New(config.vignet, clients.amqp, adapters.storage, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())

	output.streamHandler, err = amqphandler.New(config.streamHandler, clients.amqp, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider(), output.vignet.AmqpStreamHandler)
	if err != nil {
		return output, fmt.Errorf("stream: %w", err)
	}

	output.thumbnailHandler, err = amqphandler.New(config.thumbnailHandler, clients.amqp, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider(), output.vignet.AmqpThumbnailHandler)
	if err != nil {
		return output, fmt.Errorf("thumbnail: %w", err)
	}

	return output, nil
}

func (s services) Start(ctx context.Context) {
	go s.streamHandler.Start(ctx)
	go s.thumbnailHandler.Start(ctx)
	go s.vignet.Start(ctx)
}
