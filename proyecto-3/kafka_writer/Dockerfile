FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o kafka-writer

FROM alpine
WORKDIR /app
COPY --from=builder /app/kafka-writer .
CMD ["./kafka-writer"]