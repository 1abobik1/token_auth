FROM golang:1.23.1-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/auth-service .

COPY .env .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./auth-service", "--config=.env"]
