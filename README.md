# vignet

[![Build](https://github.com/ViBiOh/vignet/workflows/Build/badge.svg)](https://github.com/ViBiOh/vignet/actions)

## API

The HTTP API is pretty simple :

- `GET /health`: healthcheck of server, always respond [`okStatus (default 204)`](#usage)
- `GET /ready`: checks external dependencies availability and then respond [`okStatus (default 204)`](#usage) or `503` during [`graceDuration`](#usage) when close signal is received
- `GET /version`: value of `VERSION` environment variable
- `POST /`: generate thumbnail of the video passed in payload in binary

### Installation

Golang binary is built with static link. You can download it directly from the [GitHub Release page](https://github.com/ViBiOh/vignet/releases) or build it by yourself by cloning this repo and running `make`.

You can configure app by passing CLI args or environment variables (cf. [Usage](#usage) section). CLI override environment variables.

You'll find a Kubernetes exemple in the [`infra/`](infra) folder, using my [`app chart`](https://github.com/ViBiOh/charts/tree/main/app)

## Usage

The application can be configured by passing CLI args described below or their equivalent as environment variable. CLI values take precedence over environments variables.

Be careful when using the CLI values, if someone list the processes on the system, they will appear in plain-text. Pass secrets by environment variables: it's less easily visible.

```bash
Usage of vignet:
  --address                     string    [server] Listen address ${VIGNET_ADDRESS}
  --amqpPrefetch                int       [amqp] Prefetch count for QoS ${VIGNET_AMQP_PREFETCH} (default 1)
  --amqpURI                     string    [amqp] Address in the form amqps?://<user>:<password>@<address>:<port>/<vhost> ${VIGNET_AMQP_URI}
  --cert                        string    [server] Certificate file ${VIGNET_CERT}
  --exchange                    string    [thumbnail] AMQP Exchange Name ${VIGNET_EXCHANGE} (default "fibr")
  --graceDuration               duration  [http] Grace duration when signal received ${VIGNET_GRACE_DURATION} (default 30s)
  --idleTimeout                 duration  [server] Idle Timeout ${VIGNET_IDLE_TIMEOUT} (default 2m0s)
  --key                         string    [server] Key file ${VIGNET_KEY}
  --loggerJson                            [logger] Log format as JSON ${VIGNET_LOGGER_JSON} (default false)
  --loggerLevel                 string    [logger] Logger level ${VIGNET_LOGGER_LEVEL} (default "INFO")
  --loggerLevelKey              string    [logger] Key for level in JSON ${VIGNET_LOGGER_LEVEL_KEY} (default "level")
  --loggerMessageKey            string    [logger] Key for message in JSON ${VIGNET_LOGGER_MESSAGE_KEY} (default "msg")
  --loggerTimeKey               string    [logger] Key for timestamp in JSON ${VIGNET_LOGGER_TIME_KEY} (default "time")
  --name                        string    [server] Name ${VIGNET_NAME} (default "http")
  --okStatus                    int       [http] Healthy HTTP Status code ${VIGNET_OK_STATUS} (default 204)
  --port                        uint      [server] Listen port (0 to disable) ${VIGNET_PORT} (default 1080)
  --pprofAgent                  string    [pprof] URL of the Datadog Trace Agent (e.g. http://datadog.observability:8126) ${VIGNET_PPROF_AGENT}
  --pprofPort                   int       [pprof] Port of the HTTP server (0 to disable) ${VIGNET_PPROF_PORT} (default 0)
  --readTimeout                 duration  [server] Read Timeout ${VIGNET_READ_TIMEOUT} (default 2m0s)
  --routingKey                  string    [thumbnail] AMQP Routing Key to fibr ${VIGNET_ROUTING_KEY} (default "thumbnail_output")
  --shutdownTimeout             duration  [server] Shutdown Timeout ${VIGNET_SHUTDOWN_TIMEOUT} (default 10s)
  --storageFileSystemDirectory  /data     [storage] Path to directory. Default is dynamic. /data on a server and Current Working Directory in a terminal. ${VIGNET_STORAGE_FILE_SYSTEM_DIRECTORY}
  --storageObjectAccessKey      string    [storage] Storage Object Access Key ${VIGNET_STORAGE_OBJECT_ACCESS_KEY}
  --storageObjectBucket         string    [storage] Storage Object Bucket ${VIGNET_STORAGE_OBJECT_BUCKET}
  --storageObjectClass          string    [storage] Storage Object Class ${VIGNET_STORAGE_OBJECT_CLASS}
  --storageObjectEndpoint       string    [storage] Storage Object endpoint ${VIGNET_STORAGE_OBJECT_ENDPOINT}
  --storageObjectRegion         string    [storage] Storage Object Region ${VIGNET_STORAGE_OBJECT_REGION}
  --storageObjectSSL                      [storage] Use SSL ${VIGNET_STORAGE_OBJECT_SSL} (default true)
  --storageObjectSecretAccess   string    [storage] Storage Object Secret Access ${VIGNET_STORAGE_OBJECT_SECRET_ACCESS}
  --storagePartSize             uint      [storage] PartSize configuration ${VIGNET_STORAGE_PART_SIZE} (default 5242880)
  --streamExchange              string    [stream] Exchange name ${VIGNET_STREAM_EXCHANGE} (default "fibr")
  --streamExclusive                       [stream] Queue exclusive mode (for fanout exchange) ${VIGNET_STREAM_EXCLUSIVE} (default false)
  --streamInactiveTimeout       duration  [stream] When inactive during the given timeout, stop listening ${VIGNET_STREAM_INACTIVE_TIMEOUT} (default 0s)
  --streamMaxRetry              uint      [stream] Max send retries ${VIGNET_STREAM_MAX_RETRY} (default 3)
  --streamQueue                 string    [stream] Queue name ${VIGNET_STREAM_QUEUE} (default "stream")
  --streamRetryInterval         duration  [stream] Interval duration when send fails ${VIGNET_STREAM_RETRY_INTERVAL} (default 1h0m0s)
  --streamRoutingKey            string    [stream] RoutingKey name ${VIGNET_STREAM_ROUTING_KEY} (default "stream")
  --telemetryRate               string    [telemetry] OpenTelemetry sample rate, 'always', 'never' or a float value ${VIGNET_TELEMETRY_RATE} (default "always")
  --telemetryURL                string    [telemetry] OpenTelemetry gRPC endpoint (e.g. otel-exporter:4317) ${VIGNET_TELEMETRY_URL}
  --telemetryUint64                       [telemetry] Change OpenTelemetry Trace ID format to an unsigned int 64 ${VIGNET_TELEMETRY_UINT64} (default true)
  --thumbnailExchange           string    [thumbnail] Exchange name ${VIGNET_THUMBNAIL_EXCHANGE} (default "fibr")
  --thumbnailExclusive                    [thumbnail] Queue exclusive mode (for fanout exchange) ${VIGNET_THUMBNAIL_EXCLUSIVE} (default false)
  --thumbnailInactiveTimeout    duration  [thumbnail] When inactive during the given timeout, stop listening ${VIGNET_THUMBNAIL_INACTIVE_TIMEOUT} (default 0s)
  --thumbnailMaxRetry           uint      [thumbnail] Max send retries ${VIGNET_THUMBNAIL_MAX_RETRY} (default 3)
  --thumbnailQueue              string    [thumbnail] Queue name ${VIGNET_THUMBNAIL_QUEUE} (default "thumbnail")
  --thumbnailRetryInterval      duration  [thumbnail] Interval duration when send fails ${VIGNET_THUMBNAIL_RETRY_INTERVAL} (default 1h0m0s)
  --thumbnailRoutingKey         string    [thumbnail] RoutingKey name ${VIGNET_THUMBNAIL_ROUTING_KEY} (default "thumbnail")
  --tmpFolder                   string    [vignet] Folder used for temporary files storage ${VIGNET_TMP_FOLDER} (default "/tmp")
  --url                         string    [alcotest] URL to check ${VIGNET_URL}
  --userAgent                   string    [alcotest] User-Agent for check ${VIGNET_USER_AGENT} (default "Alcotest")
  --writeTimeout                duration  [server] Write Timeout ${VIGNET_WRITE_TIMEOUT} (default 2m0s)
```
