# Order Management Service Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY order-management-service/ ./order-management-service/

# Build the order management service
WORKDIR /app/order-management-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/order-management-service/main .

CMD ["./main"]
