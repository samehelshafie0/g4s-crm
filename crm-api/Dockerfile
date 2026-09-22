# ─── Build Stage ─────────────────────────────────────────────
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o g4s-crm-api ./cmd/server

# ─── Runtime Stage ───────────────────────────────────────────
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget python3 python3-venv libreoffice-writer libreoffice-calc fonts-dejavu-core fonts-liberation && rm -rf /var/lib/apt/lists/*
COPY pdf-requirements.txt ./
RUN python3 -m venv /opt/pdf-venv && /opt/pdf-venv/bin/pip install --no-cache-dir -r pdf-requirements.txt
ENV PDF_PYTHON=/opt/pdf-venv/bin/python
ENV HOME=/tmp

RUN groupadd --system appgroup && useradd --system --gid appgroup --home-dir /tmp appuser

COPY --from=builder /app/g4s-crm-api .
COPY --from=builder /app/migrations ./migrations

RUN mkdir -p storage keys && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/ready || exit 1

CMD ["./g4s-crm-api"]
