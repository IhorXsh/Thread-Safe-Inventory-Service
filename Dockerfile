# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/inventory-service ./cmd/inventory-service

FROM gcr.io/distroless/static-debian12 AS runtime
WORKDIR /app

COPY --from=builder /out/inventory-service /app/inventory-service

ENV PORT=8080
ENV METRICS_PORT=9090

EXPOSE 8080
EXPOSE 9090

USER nonroot:nonroot
ENTRYPOINT ["/app/inventory-service"]
