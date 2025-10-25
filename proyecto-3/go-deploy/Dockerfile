# Etapa de compilación
FROM golang:1.25.2-alpine AS builder

RUN apk add --no-cache git protobuf protobuf-dev build-base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Instala los plugins de Go para protoc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Asegúrate de que GOPATH/bin esté en el PATH
ENV PATH="$PATH:$(go env GOPATH)/bin"

RUN protoc --go_out=. --go-grpc_out=. proto/weathertweet.proto
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-go main.go

# Etapa final
FROM alpine:3.22

# Instala certificados
RUN apk add --no-cache ca-certificates

# Crear usuario no root
RUN adduser -D app_user
USER app_user

WORKDIR /app
COPY --from=builder /app/api-go .

EXPOSE 50051
CMD ["./api-go"]
