# catalog-service
Sank ingredients, menu products, units and versioned recipes

## Running

```sh
go run ./cmd/server
```

The HTTP server listens on port `8080`.

## Liveness

`GET health/live` returns `200 OK` with an empty body when the
service can handle HTTP requests. It does not check databases or other
dependencies.

```sh
curl -i http://localhost:8080/health/live
```

