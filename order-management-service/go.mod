module github.com/vdntruong/dddcqrs/order-management-service

go 1.25

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/vdntruong/dddcqrs/shared v0.0.0
)

require github.com/confluentinc/confluent-kafka-go/v2 v2.3.0 // indirect

replace github.com/vdntruong/dddcqrs/shared => ../shared
