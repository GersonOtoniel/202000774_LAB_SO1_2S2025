FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o rabbitmq-writer

FROM alpine
WORKDIR /app
COPY --from=builder /app/rabbitmq-writer .
CMD ["./rabbitmq-writer"]