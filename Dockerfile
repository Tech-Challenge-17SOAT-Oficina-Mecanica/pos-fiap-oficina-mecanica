FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o oficina-mecanica \
    ./cmd/oficina-mecanica

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app/oficina-mecanica .

EXPOSE 8080

CMD ["./oficina-mecanica"]
