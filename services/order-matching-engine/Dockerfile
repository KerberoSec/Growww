# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM rust:1.78-slim-bookworm AS builder

WORKDIR /build

# Cache dependencies
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo fetch

# Build actual source
COPY . .
RUN cargo build --release --locked

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM gcr.io/distroless/cc-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="Growww Order Matching Engine" \
      org.opencontainers.image.vendor="Growww Exchange" \
      org.opencontainers.image.licenses="UNLICENSED"

COPY --from=builder /build/target/release/order-matching-engine /usr/local/bin/order-matching-engine

USER nonroot:nonroot

EXPOSE 9090

ENTRYPOINT ["/usr/local/bin/order-matching-engine"]
