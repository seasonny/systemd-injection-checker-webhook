FROM golang:1.20 AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o systemd-injection-checker-webhook

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/systemd-injection-checker-webhook .
CMD ["./systemd-injection-checker-webhook"]
