FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o oficina-mecanica \
    ./cmd/oficina-mecanica

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && wget -O /etc/ssl/certs/aws-rds-global-bundle.pem \
       https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem \
    && addgroup -S oficina && adduser -S -G oficina oficina

COPY --from=build --chown=oficina:oficina /app/oficina-mecanica .

USER oficina

EXPOSE 8080

CMD ["./oficina-mecanica"]
