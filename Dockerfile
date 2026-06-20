# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately from source changes
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 for a static binary that runs in a minimal base image
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/photo_fetch ./cmd/photo_fetch

# ---- Run stage ----
FROM alpine:3.20

WORKDIR /app

# Optional: TLS root certs, needed if your service makes outbound HTTPS calls
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/photo_fetch .

EXPOSE 8080

ENTRYPOINT ["./photo_fetch"]