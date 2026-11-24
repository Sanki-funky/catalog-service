# catalog-service

Šank ingredients, menu products, units and versioned recipes.

## Running

```sh
go run ./cmd/api
```

The server listens on port `8080` and logs JSON to stdout. Stop it with Ctrl+C: it finishes in-flight requests before exiting.

## Configuration

Settings come from environment variables. Every one is optional and has a working default, so the service starts with none set. An invalid value stops the service at startup with an error naming the variable.

| Variable           | Default | Meaning                                                 |
| ------------------ | ------- | ------------------------------------------------------- |
| `PORT`             | `8080`  | HTTP port for the API, health checks and metrics        |
| `LOG_LEVEL`        | `info`  | Minimum log level: `debug`, `info`, `warn` or `error`   |
| `SHUTDOWN_TIMEOUT` | `10s`   | How long in-flight requests get to finish after SIGTERM |

[.env.example](.env.example) lists them. For example:

```sh
PORT=9090 LOG_LEVEL=debug go run ./cmd/api
```

To add a setting, add a field to `Config` in [internal/config/config.go](internal/config/config.go) and one `read(...)` line in `FromLookup`, then document it here and in `.env.example`.

## Logs

The service writes JSON lines to stdout, so they show up wherever you run it:

```sh
cd catalog-service
LOG_LEVEL=debug go run ./cmd/api
```

Raw JSON is hard to read, so pipe it through `jq` to get pretty, colored output:

```sh
LOG_LEVEL=debug go run ./cmd/api | jq
```

To see only errors:

```sh
go run ./cmd/api | jq 'select(.level == "ERROR")'
```

## Endpoints

| Endpoint            | Purpose                                                                                                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `GET /health/live`  | `200` with an empty body while the process can answer HTTP. Never checks dependencies. Kubernetes liveness probe.                                                         |
| `GET /health/ready` | `200 {"status":"ok"}` when the service can take traffic. `503` with the failing checks, or `{"status":"shutting down"}` once shutdown starts. Kubernetes readiness probe. |
| `GET /metrics`      | Prometheus metrics: `http_requests_total` and `http_request_duration_seconds` per route, plus Go runtime and process metrics.                                             |

```sh
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
curl http://localhost:8080/metrics
```

Every request is logged with its method, path, route, status and duration. Health and metrics requests are logged at `debug` level so probes don't flood the logs.

## Layout

```text
cmd/api/              main: config, logger, wiring, signal handling
internal/config/      environment variables -> Config
internal/web/httpapi/ routes, health, metrics, request logging, server lifecycle
```

Business code will go in `internal/domain` and `internal/usecase`, and PostgreSQL/NATS code in `internal/adapters`, as described in the platform [development guide](https://github.com/Sanki-funky/platform/blob/main/docs/development.md).

## Tests

```sh
go test ./...
```
