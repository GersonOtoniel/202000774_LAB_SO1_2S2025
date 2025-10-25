FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o consumer-kafka

FROM alpine
WORKDIR /app
COPY --from=builder /app/consumer-kafka .
CMD ["./consumer-kafka"]