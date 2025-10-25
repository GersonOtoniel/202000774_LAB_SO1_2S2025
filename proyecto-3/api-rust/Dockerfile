FROM rust:alpine AS builder

WORKDIR /app

RUN apk add --no-cache protobuf gcc musl-dev


# Copia todo el código
COPY . .

# Compila la aplicación en release
RUN cargo build --release

FROM alpine

WORKDIR /app

RUN adduser -D app_user
USER app_user

COPY --from=builder --chown=app_user:app_user /app/target/release/api-rust ./api-rust

# Expone el puerto Actix
EXPOSE 8080

ENV GRPC_SERVER_URL=http://server-go-service:50051


CMD ["./api-rust"]
