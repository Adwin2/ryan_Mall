package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ryan-mall-microservices/internal/shared/infrastructure"

	"github.com/segmentio/kafka-go"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	Handle(ctx context.Context, message *Message) error
	Topic() string
}

// Message 消息结构
type Message struct {
	ID        string            `json:"id"`
	Topic     string            `json:"topic"`
	Key       string            `json:"key"`
	Value     []byte            `json:"value"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	Partition int               `json:"partition"`
	Offset    int64             `json:"offset"`
}

// Producer Kafka生产者接口
type Producer interface {
	SendMessage(ctx context.Context, topic, key string, value interface{}) error
	SendMessageWithHeaders(ctx context.Context, topic, key string, value interface{}, headers map[string]string) error
	Close() error
}

// Consumer Kafka消费者接口
type Consumer interface {
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error
	Close() error
}

// KafkaProducer Kafka生产者实现
type KafkaProducer struct {
	writers map[string]*kafka.Writer
	mutex   sync.RWMutex
	brokers []string
	logger  infrastructure.Logger
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(brokers []string) *KafkaProducer {
	return &KafkaProducer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
		logger:  infrastructure.GetLogger(),
	}
}

// SendMessage 发送消息
func (p *KafkaProducer) SendMessage(ctx context.Context, topic, key string, value interface{}) error {
	return p.SendMessageWithHeaders(ctx, topic, key, value, nil)
}

// SendMessageWithHeaders 发送带头部的消息
func (p *KafkaProducer) SendMessageWithHeaders(ctx context.Context, topic, key string, value interface{}, headers map[string]string) error {
	// 序列化消息值
	var valueBytes []byte
	var err error

	switch v := value.(type) {
	case []byte:
		valueBytes = v
	case string:
		valueBytes = []byte(v)
	default:
		valueBytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message value: %w", err)
		}
	}

	// 获取或创建writer
	writer := p.getWriter(topic)

	// 构建Kafka消息
	kafkaMessage := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
		Time:  time.Now(),
	}

	// 添加头部
	if headers != nil {
		kafkaMessage.Headers = make([]kafka.Header, 0, len(headers))
		for k, v := range headers {
			kafkaMessage.Headers = append(kafkaMessage.Headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	// 发送消息
	err = writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		p.logger.Error("Failed to send message",
			infrastructure.String("topic", topic),
			infrastructure.String("key", key),
			infrastructure.Error(err),
		)
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.logger.Debug("Message sent successfully",
		infrastructure.String("topic", topic),
		infrastructure.String("key", key),
	)

	return nil
}

// getWriter 获取或创建writer
func (p *KafkaProducer) getWriter(topic string) *kafka.Writer {
	p.mutex.RLock()
	writer, exists := p.writers[topic]
	p.mutex.RUnlock()

	if exists {
		return writer
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 双重检查
	if writer, exists := p.writers[topic]; exists {
		return writer
	}

	// 创建新的writer
	writer = &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne,
		Async:        false,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    100,
	}

	p.writers[topic] = writer
	return writer
}

// Close 关闭生产者
func (p *KafkaProducer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var lastErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Error("Failed to close writer",
				infrastructure.String("topic", topic),
				infrastructure.Error(err),
			)
			lastErr = err
		}
		delete(p.writers, topic)
	}

	return lastErr
}

// KafkaConsumer Kafka消费者实现
type KafkaConsumer struct {
	readers map[string]*kafka.Reader
	mutex   sync.RWMutex
	brokers []string
	groupID string
	logger  infrastructure.Logger
	cancel  context.CancelFunc
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(brokers []string, groupID string) *KafkaConsumer {
	return &KafkaConsumer{
		readers: make(map[string]*kafka.Reader),
		brokers: brokers,
		groupID: groupID,
		logger:  infrastructure.GetLogger(),
	}
}

// Subscribe 订阅主题
func (c *KafkaConsumer) Subscribe(ctx context.Context, topics []string, handler MessageHandler) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	for _, topic := range topics {
		reader := c.getReader(topic)
		
		// 启动消费协程
		go c.consumeMessages(ctx, reader, handler)
	}

	return nil
}

// getReader 获取或创建reader
func (c *KafkaConsumer) getReader(topic string) *kafka.Reader {
	c.mutex.RLock()
	reader, exists := c.readers[topic]
	c.mutex.RUnlock()

	if exists {
		return reader
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 双重检查
	if reader, exists := c.readers[topic]; exists {
		return reader
	}

	// 创建新的reader
	reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.brokers,
		Topic:          topic,
		GroupID:        c.groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	c.readers[topic] = reader
	return reader
}

// consumeMessages 消费消息
func (c *KafkaConsumer) consumeMessages(ctx context.Context, reader *kafka.Reader, handler MessageHandler) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 读取消息
			kafkaMessage, err := reader.ReadMessage(ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				c.logger.Error("Failed to read message",
					infrastructure.String("topic", reader.Config().Topic),
					infrastructure.Error(err),
				)
				continue
			}

			// 转换消息格式
			message := &Message{
				ID:        fmt.Sprintf("%s-%d-%d", reader.Config().Topic, kafkaMessage.Partition, kafkaMessage.Offset),
				Topic:     kafkaMessage.Topic,
				Key:       string(kafkaMessage.Key),
				Value:     kafkaMessage.Value,
				Headers:   make(map[string]string),
				Timestamp: kafkaMessage.Time,
				Partition: kafkaMessage.Partition,
				Offset:    kafkaMessage.Offset,
			}

			// 转换头部
			for _, header := range kafkaMessage.Headers {
				message.Headers[header.Key] = string(header.Value)
			}

			// 处理消息
			if err := handler.Handle(ctx, message); err != nil {
				c.logger.Error("Failed to handle message",
					infrastructure.String("topic", message.Topic),
					infrastructure.String("key", message.Key),
					infrastructure.Int64("offset", message.Offset),
					infrastructure.Error(err),
				)
				// 这里可以实现重试逻辑或死信队列
				continue
			}

			c.logger.Debug("Message processed successfully",
				infrastructure.String("topic", message.Topic),
				infrastructure.String("key", message.Key),
				infrastructure.Int64("offset", message.Offset),
			)
		}
	}
}

// Close 关闭消费者
func (c *KafkaConsumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var lastErr error
	for topic, reader := range c.readers {
		if err := reader.Close(); err != nil {
			c.logger.Error("Failed to close reader",
				infrastructure.String("topic", topic),
				infrastructure.Error(err),
			)
			lastErr = err
		}
		delete(c.readers, topic)
	}

	return lastErr
}

// EventBus 基于Kafka的事件总线
type EventBus struct {
	producer Producer
	consumer Consumer
	handlers map[string]MessageHandler
	mutex    sync.RWMutex
	logger   infrastructure.Logger
}

// NewEventBus 创建事件总线
func NewEventBus(producer Producer, consumer Consumer) *EventBus {
	return &EventBus{
		producer: producer,
		consumer: consumer,
		handlers: make(map[string]MessageHandler),
		logger:   infrastructure.GetLogger(),
	}
}

// PublishEvent 发布事件
func (e *EventBus) PublishEvent(ctx context.Context, topic string, event interface{}) error {
	// 生成事件ID作为key
	eventID := fmt.Sprintf("event-%d", time.Now().UnixNano())
	
	return e.producer.SendMessage(ctx, topic, eventID, event)
}

// SubscribeEvent 订阅事件
func (e *EventBus) SubscribeEvent(ctx context.Context, topic string, handler MessageHandler) error {
	e.mutex.Lock()
	e.handlers[topic] = handler
	e.mutex.Unlock()

	return e.consumer.Subscribe(ctx, []string{topic}, handler)
}

// Close 关闭事件总线
func (e *EventBus) Close() error {
	var lastErr error
	
	if err := e.producer.Close(); err != nil {
		lastErr = err
	}
	
	if err := e.consumer.Close(); err != nil {
		lastErr = err
	}
	
	return lastErr
}
