package stream

import "github.com/confluentinc/confluent-kafka-go/kafka"

type KafkaStreamConfig interface {
	NewProducer(km *kafka.ConfigMap) (*kafka.Producer, error)
	NewConsumer(km *kafka.ConfigMap) (*kafka.Consumer, error)
	SetDeliveryError(f func(*kafka.Message))
	SetTopic(topic string)
	SetPrefix(prefix string)
	SetBrokers(brokers string)
	SetFlags()
	ProducerDefaults() *kafka.ConfigMap
	Produce(topic *string, value []byte) error
	Flush(ms int) int
	FullTopic(t string) string
	ChannelProduce(topic *string, value []byte)
	GetConsumer() *kafka.Consumer
	GetProducer() *kafka.Producer
	GetBrokers() string
	GetPrefix() string
	Close() error
}
