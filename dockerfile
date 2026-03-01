FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-w -s' -o app ./cmd/api/main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/app .
COPY config.yaml .

ENV TZ=Europe/Moscow \
    LOG_LEVEL=info

EXPOSE 8000
CMD ["./app"]
