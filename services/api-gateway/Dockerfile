# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.24-bookworm AS builder

WORKDIR /build

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o api-gateway ./src/...

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="Growww API Gateway" \
      org.opencontainers.image.vendor="Growww Exchange" \
      org.opencontainers.image.licenses="UNLICENSED"

COPY --from=builder /build/api-gateway /usr/local/bin/api-gateway

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/api-gateway"]
