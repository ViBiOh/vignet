package vignet

import (
	"bytes"
	"context"
	"flag"
	"log/slog"
	"sync"

	absto "github.com/ViBiOh/absto/pkg/model"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/vignet/pkg/model"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	SmallSize = 150

	hlsExtension = ".m3u8"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 32*1024))
	},
}

type Config struct {
	TmpFolder string

	AmqpExchange   string
	AmqpRoutingKey string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("TmpFolder", "Folder used for temporary files storage").Prefix(prefix).DocPrefix("vignet").StringVar(fs, &config.TmpFolder, "/tmp", overrides)
	flags.New("Exchange", "AMQP Exchange Name").Prefix(prefix).DocPrefix("thumbnail").StringVar(fs, &config.AmqpExchange, "fibr", overrides)
	flags.New("RoutingKey", "AMQP Routing Key to fibr").Prefix(prefix).DocPrefix("thumbnail").StringVar(fs, &config.AmqpRoutingKey, "thumbnail_output", overrides)

	return &config
}

type Service struct {
	done               chan struct{}
	stop               chan struct{}
	streamRequestQueue chan model.Request
	storage            absto.Storage
	tracer             trace.Tracer
	amqpClient         *amqp.Client
	metric             metric.Int64Counter
	tmpFolder          string
	amqpExchange       string
	amqpRoutingKey     string
}

func New(config *Config, amqpClient *amqp.Client, storageService absto.Storage, meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) Service {
	service := Service{
		tmpFolder: config.TmpFolder,
		storage:   storageService,

		amqpClient:     amqpClient,
		amqpExchange:   config.AmqpExchange,
		amqpRoutingKey: config.AmqpRoutingKey,

		streamRequestQueue: make(chan model.Request, 4),
		stop:               make(chan struct{}),
		done:               make(chan struct{}),
	}

	if meterProvider != nil {
		meter := meterProvider.Meter("github.com/ViBiOh/vignet/pkg/vignet")

		var err error

		service.metric, err = meter.Int64Counter("vignet.item")
		if err != nil {
			slog.LogAttrs(context.Background(), slog.LevelError, "create vignet counter", slog.Any("error", err))
		}
	}

	if tracerProvider != nil {
		service.tracer = tracerProvider.Tracer("vignet")
	}

	return service
}
