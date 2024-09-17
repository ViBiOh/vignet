package main

import (
	"flag"
	"os"
	"time"

	"github.com/ViBiOh/absto/pkg/absto"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/amqp"
	"github.com/ViBiOh/httputils/v4/pkg/amqphandler"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	vignet "github.com/ViBiOh/vignet/pkg/vignet"
)

type configuration struct {
	alcotest  *alcotest.Config
	logger    *logger.Config
	telemetry *telemetry.Config
	pprof     *pprof.Config
	server    *server.Config
	health    *health.Config

	vignet           *vignet.Config
	absto            *absto.Config
	amqp             *amqp.Config
	streamHandler    *amqphandler.Config
	thumbnailHandler *amqphandler.Config
}

func newConfig() configuration {
	fs := flag.NewFlagSet("vignet", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:    logger.Flags(fs, "logger"),
		alcotest:  alcotest.Flags(fs, ""),
		telemetry: telemetry.Flags(fs, "telemetry"),
		pprof:     pprof.Flags(fs, "pprof"),
		health:    health.Flags(fs, ""),

		server: server.Flags(fs, "", flags.NewOverride("ReadTimeout", 2*time.Minute), flags.NewOverride("WriteTimeout", 2*time.Minute)),

		vignet:           vignet.Flags(fs, ""),
		absto:            absto.Flags(fs, "storage", flags.NewOverride("FileSystemDirectory", "")),
		amqp:             amqp.Flags(fs, "amqp"),
		streamHandler:    amqphandler.Flags(fs, "stream", flags.NewOverride("Exchange", "fibr"), flags.NewOverride("Queue", "stream"), flags.NewOverride("RoutingKey", "stream")),
		thumbnailHandler: amqphandler.Flags(fs, "thumbnail", flags.NewOverride("Exchange", "fibr"), flags.NewOverride("Queue", "thumbnail"), flags.NewOverride("RoutingKey", "thumbnail")),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
