# Order Reporting Service Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY order-reporting-service/ ./order-reporting-service/

# Build the order reporting service
WORKDIR /app/order-reporting-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/order-reporting-service/main .

CMD ["./main"]
