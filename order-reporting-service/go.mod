module github.com/vdntruong/dddcqrs/order-reporting-service

go 1.25

require (
	github.com/gorilla/mux v1.8.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/vdntruong/dddcqrs/shared v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace github.com/vdntruong/dddcqrs/shared => ../shared
