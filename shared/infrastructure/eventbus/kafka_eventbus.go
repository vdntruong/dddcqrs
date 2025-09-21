package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/vdntruong/dddcqrs/shared/domain/events"
)

type KafkaEventBus struct {
    producer *kafka.Producer
    consumer *kafka.Consumer
    brokers  string
}

func NewKafkaEventBus(brokers string) *KafkaEventBus {
    // Producer configuration
    producer, err := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": brokers,
        "client.id":        "order-management-service",
        "acks":            "all",
        "retries":         "3",
    })
    if err != nil {
        log.Fatalf("Failed to create Kafka producer: %v", err)
    }
    
    // Consumer configuration
    consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": brokers,
        "group.id":         "order-reporting-service",
        "auto.offset.reset": "earliest",
        "enable.auto.commit": "false",
    })
    if err != nil {
        log.Fatalf("Failed to create Kafka consumer: %v", err)
    }
    
    return &KafkaEventBus{
        producer: producer,
        consumer: consumer,
        brokers:  brokers,
    }
}

func (k *KafkaEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    topic := "orders"
    message := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Value: eventData,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(event.Type())},
            {Key: "aggregate-id", Value: []byte(event.AggregateID())},
        },
    }
    
    deliveryChan := make(chan kafka.Event)
    err = k.producer.Produce(message, deliveryChan)
    if err != nil {
        return fmt.Errorf("failed to produce message: %w", err)
    }
    
    // Wait for delivery confirmation
    select {
    case e := <-deliveryChan:
        m := e.(*kafka.Message)
        if m.TopicPartition.Error != nil {
            return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
        }
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (k *KafkaEventBus) Subscribe(ctx context.Context, topic string, handler func(events.DomainEvent) error) error {
    err := k.consumer.Subscribe(topic, nil)
    if err != nil {
        return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
    }
    
    go func() {
        defer k.consumer.Close()
        
        for {
            select {
            case <-ctx.Done():
                return
            default:
                msg, err := k.consumer.ReadMessage(-1)
                if err != nil {
                    log.Printf("Error reading message: %v", err)
                    continue
                }
                
                var event events.DomainEvent
                if err := json.Unmarshal(msg.Value, &event); err != nil {
                    log.Printf("Error unmarshaling event: %v", err)
                    continue
                }
                
                if err := handler(event); err != nil {
                    log.Printf("Error handling event: %v", err)
                    // In production, you might want to send to a dead letter queue
                    continue
                }
                
                // Commit the message after successful processing
                if _, err := k.consumer.CommitMessage(msg); err != nil {
                    log.Printf("Error committing message: %v", err)
                }
            }
        }
    }()
    
    return nil
}

func (k *KafkaEventBus) Close() error {
    if k.producer != nil {
        k.producer.Close()
    }
    if k.consumer != nil {
        k.consumer.Close()
    }
    return nil
}
