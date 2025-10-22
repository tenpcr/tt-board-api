FROM golang:1.24.5 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ttboard_api

FROM scratch

WORKDIR /app

COPY --from=builder /app/ttboard_api .

CMD ["/app/ttboard_api"]