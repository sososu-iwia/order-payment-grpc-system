FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o service ./cmd/payment-service

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/service /app/service

EXPOSE 8082

CMD ["/app/service"]