package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"ryan-mall-microservices/internal/shared/events"

	"github.com/segmentio/kafka-go"
)

// KafkaEventBus Kafka事件总线实现
type KafkaEventBus struct {
	writer   *kafka.Writer
	readers  map[string]*kafka.Reader
	handlers map[string][]events.EventHandler
	brokers  []string
	topic    string
}

// NewKafkaEventBus 创建Kafka事件总线
func NewKafkaEventBus(brokers []string, topic string) *KafkaEventBus {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		Compression:  kafka.Snappy,
		BatchTimeout: 10 * time.Millisecond,
	}

	return &KafkaEventBus{
		writer:   writer,
		readers:  make(map[string]*kafka.Reader),
		handlers: make(map[string][]events.EventHandler),
		brokers:  brokers,
		topic:    topic,
	}
}

// Publish 发布事件到Kafka
func (k *KafkaEventBus) Publish(ctx context.Context, events ...events.Event) error {
	messages := make([]kafka.Message, len(events))
	
	for i, event := range events {
		// 序列化事件
		eventData, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// 创建Kafka消息
		messages[i] = kafka.Message{
			Key:   []byte(event.EventType()),
			Value: eventData,
			Headers: []kafka.Header{
				{Key: "event_type", Value: []byte(event.EventType())},
				{Key: "event_id", Value: []byte(event.EventID())},
				{Key: "aggregate_id", Value: []byte(event.AggregateID())},
				{Key: "timestamp", Value: []byte(event.OccurredAt().Format(time.RFC3339))},
			},
		}
	}

	// 发送消息到Kafka
	return k.writer.WriteMessages(ctx, messages...)
}

// Subscribe 订阅事件类型
func (k *KafkaEventBus) Subscribe(handler events.EventHandler) error {
	eventType := handler.EventType()
	
	// 添加处理器
	k.handlers[eventType] = append(k.handlers[eventType], handler)
	
	// 如果是第一个订阅该事件类型的处理器，创建Reader
	if len(k.handlers[eventType]) == 1 {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     k.brokers,
			Topic:       k.topic,
			GroupID:     fmt.Sprintf("%s-consumer", eventType),
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
			MaxWait:     1 * time.Second,
			StartOffset: kafka.LastOffset,
		})
		
		k.readers[eventType] = reader
		
		// 启动消费者协程
		go k.consumeEvents(eventType, reader)
	}
	
	return nil
}

// Unsubscribe 取消订阅事件类型
func (k *KafkaEventBus) Unsubscribe(eventType string) error {
	// 移除处理器
	delete(k.handlers, eventType)
	
	// 关闭并移除Reader
	if reader, exists := k.readers[eventType]; exists {
		reader.Close()
		delete(k.readers, eventType)
	}
	
	return nil
}

// consumeEvents 消费事件
func (k *KafkaEventBus) consumeEvents(eventType string, reader *kafka.Reader) {
	for {
		// 读取消息
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message for event type %s: %v", eventType, err)
			continue
		}

		// 检查事件类型
		var messageEventType string
		for _, header := range message.Headers {
			if header.Key == "event_type" {
				messageEventType = string(header.Value)
				break
			}
		}

		// 只处理匹配的事件类型
		if messageEventType != eventType {
			continue
		}

		// 反序列化事件
		var baseEvent events.BaseEvent
		if err := json.Unmarshal(message.Value, &baseEvent); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// 调用处理器
		handlers := k.handlers[eventType]
		for _, handler := range handlers {
			go func(h events.EventHandler) {
				if err := h.Handle(context.Background(), &baseEvent); err != nil {
					log.Printf("Error handling event %s: %v", eventType, err)
				}
			}(handler)
		}
	}
}

// Close 关闭Kafka连接
func (k *KafkaEventBus) Close() error {
	// 关闭Writer
	if err := k.writer.Close(); err != nil {
		return err
	}

	// 关闭所有Readers
	for _, reader := range k.readers {
		if err := reader.Close(); err != nil {
			return err
		}
	}

	return nil
}

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}

// NewKafkaConfigFromEnv 从环境变量创建Kafka配置
func NewKafkaConfigFromEnv() *KafkaConfig {
	return &KafkaConfig{
		Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		Topic:   getEnv("KAFKA_TOPIC", "ryan-mall-events"),
	}
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
