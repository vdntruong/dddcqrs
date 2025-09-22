# Order Reporting Service Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Disable Go workspace to avoid referencing modules not present in the image
ENV GOWORK=off
# Enable CGO for confluent-kafka-go and install build deps
ENV CGO_ENABLED=1
RUN apk add --no-cache build-base pkgconfig librdkafka-dev
# Copy required sources
COPY shared/ ./shared/
COPY order-reporting-service/ ./order-reporting-service/

# Build the order reporting service
WORKDIR /app/order-reporting-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates librdkafka
WORKDIR /root/

COPY --from=builder /app/order-reporting-service/main .

CMD ["./main"]
