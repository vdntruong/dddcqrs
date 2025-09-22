# Order Management Service Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Disable Go workspace to avoid referencing modules not present in the image
ENV GOWORK=off
# Enable CGO for confluent-kafka-go and install build deps
ENV CGO_ENABLED=1
RUN apk add --no-cache build-base pkgconfig librdkafka-dev git
# Copy required sources
COPY shared/ ./shared/
COPY order-management-service/ ./order-management-service/

# Build the order management service
WORKDIR /app/order-management-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates librdkafka
WORKDIR /root/

COPY --from=builder /app/order-management-service/main .

CMD ["./main"]
