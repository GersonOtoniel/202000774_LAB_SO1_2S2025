FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/api .

# Expone el puerto 8081, 8082 y 8083
EXPOSE 8082

CMD ["./api"]