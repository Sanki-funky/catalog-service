# catalog-service
Sank ingredients, menu products, units and versioned recipes

## Running

```sh
go run ./cmd/server
```

The HTTP server listens on port `8080`. Set `PORT` to use another one:

```sh
PORT=9090 go run ./cmd/server
```

## Docker

```sh
docker build -t catalog-service .
docker run --rm -p 8080:8080 catalog-service
```

The image is built for your machine's architecture. The cluster runs on
`linux/amd64`, so build for that when the image is going there:

```sh
docker build --platform linux/amd64 -t catalog-service .
```

The final image is based on distroless and contains only the binary. It has
no shell, so use `kubectl logs` or `docker logs` instead of exec-ing into it.

## Health checks

### Liveness

`GET /health/live` returns `200 OK` with an empty body when the
service can handle HTTP requests. It does not check databases or other
dependencies.

```sh
curl -i http://localhost:8080/health/live
```

### Readiness

`GET /health/ready` returns `200 OK` with an empty body when the service
should receive traffic, and `503 Service Unavailable` when it shouldn't:
before the server is listening and once it starts shutting down. Checks on
dependencies such as the database will be added here.

```sh
curl -i http://localhost:8080/health/ready
```

## Shutdown

On `SIGTERM` or `Ctrl+C` the service reports not ready, stops accepting new
connections and gives in-flight requests up to 10 seconds to finish. A second
signal stops it immediately.
