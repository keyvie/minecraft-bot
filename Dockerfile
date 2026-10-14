FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/minecraft-bot .

FROM alpine:3.24
RUN adduser -D -u 10001 bot
WORKDIR /app
COPY --from=builder /app/minecraft-bot ./minecraft-bot
USER bot
ENTRYPOINT ["./minecraft-bot"]