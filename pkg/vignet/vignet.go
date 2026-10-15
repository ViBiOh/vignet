package vignet

import (
	"bytes"
	"context"
	"flag"
	"log/slog"
	"os/exec"
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
	TmpFolder      string
	FfmpegPath     string
	FfprobePath    string
	AmqpExchange   string
	AmqpRoutingKey string
	Hwaccel        bool
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("TmpFolder", "Folder used for temporary files storage").Prefix(prefix).DocPrefix("vignet").StringVar(fs, &config.TmpFolder, "/tmp", overrides)
	flags.New("Exchange", "AMQP Exchange Name").Prefix(prefix).DocPrefix("thumbnail").StringVar(fs, &config.AmqpExchange, "fibr", overrides)
	flags.New("RoutingKey", "AMQP Routing Key to fibr").Prefix(prefix).DocPrefix("thumbnail").StringVar(fs, &config.AmqpRoutingKey, "thumbnail_output", overrides)
	flags.New("FfmpegPath", "ffmpeg path").Prefix(prefix).DocPrefix("vignet").StringVar(fs, &config.FfmpegPath, "/usr/bin/ffmpeg", overrides)
	flags.New("FfprobePath", "ffmpeg path").Prefix(prefix).DocPrefix("vignet").StringVar(fs, &config.FfprobePath, "/usr/bin/ffprobe", overrides)

	flags.New("Hwaccel", "Enable ffmpeg hardware acceleration (-hwaccel auto), requires the matching libraries on the host").Prefix(prefix).DocPrefix("vignet").BoolVar(fs, &config.Hwaccel, false, overrides)

	return &config
}

type GatedQueue[T any] struct {
	ch     chan T
	mutex  sync.RWMutex
	closed bool
}

func NewGatedQueue[T any](ch chan T) *GatedQueue[T] {
	return &GatedQueue[T]{ch: ch}
}

func (gq *GatedQueue[T]) Send(ctx context.Context, content T) bool {
	gq.mutex.RLock()
	defer gq.mutex.RUnlock()

	if gq.closed {
		return false
	}

	select {
	case gq.ch <- content:
		return true
	case <-ctx.Done():
		return false
	}
}

func (gq *GatedQueue[T]) Close() {
	gq.mutex.Lock()
	defer gq.mutex.Unlock()

	if gq.closed {
		return
	}

	gq.closed = true
	close(gq.ch)
}

type Service struct {
	storage            absto.Storage
	tracer             trace.Tracer
	metric             metric.Int64Counter
	done               chan struct{}
	streamRequestQueue *GatedQueue[model.Request]
	amqpClient         *amqp.Client
	tmpFolder          string
	ffmpegPath         string
	ffprobePath        string
	amqpExchange       string
	amqpRoutingKey     string
	hwaccel            bool
}

func New(config *Config, amqpClient *amqp.Client, storageService absto.Storage, meterProvider metric.MeterProvider, tracerProvider trace.TracerProvider) Service {
	service := Service{
		tmpFolder:   config.TmpFolder,
		ffmpegPath:  config.FfmpegPath,
		ffprobePath: config.FfprobePath,
		hwaccel:     config.Hwaccel,
		storage:     storageService,

		amqpClient:     amqpClient,
		amqpExchange:   config.AmqpExchange,
		amqpRoutingKey: config.AmqpRoutingKey,

		streamRequestQueue: NewGatedQueue(make(chan model.Request, 4)),
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

func (s Service) ffmpegCommand(ctx context.Context, args ...string) *exec.Cmd {
	if s.hwaccel {
		args = append([]string{"-hwaccel", "auto"}, args...)
	}

	return exec.CommandContext(ctx, s.ffmpegPath, args...)
}
