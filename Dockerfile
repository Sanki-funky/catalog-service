# syntax=docker/dockerfile:1

# Build stage. It runs on the build machine's own architecture and
# cross-compiles for the target, so building an amd64 image for the cluster
# on an Apple Silicon Mac doesn't need slow emulation.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /src

# Dependencies get their own layer, so changing the code doesn't
# download them again. go.sum is optional until we have dependencies.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
# CGO is off so the binary is fully static and runs without a C library.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/catalog-service ./cmd/server

# Runtime stage: no shell or package manager, only the binary,
# CA certificates and timezone data. Runs as an unprivileged user.
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /out/catalog-service /catalog-service

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/catalog-service"]
