package stream

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/atsu/goat/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kelseyhightower/envconfig"
)

const DefaultFlushInterval = 100

// AtsuUnsetGroupId is used by library callers, but also manually set as default below.
const AtsuUnsetGroupId = "atsu-unset-group-id"

// StreamConfig provides Kafka-related configuration
type StreamConfig struct {
	Brokers  string `json:"brokers" yaml:"brokers" default:"kafka-atsu-prod-01:9092,kafka-atsu-prod-02:9092,kafka-atsu-prod-03:9092"`
	Prefix   string `json:"prefix" yaml:"prefix" default:"atsu"` // defaults to atsu
	Topic    string `json:"topic" yaml:"topic" default:"unset"`  // defaults to unset
	Messages int    `json:"messages" yaml:"messages"`
	Bytes    int    `json:"bytes" yaml:"bytes"`

	Timeout  time.Duration `ignored:"true" json:"-"` // Must be set explicitly
	Interval time.Duration `ignored:"true" json:"-"` // Must be set explicitly

	Offset          string `default:"latest" json:"offset" yaml:"offset"`
	GroupId         string `default:"atsu-unset-group-id" split_words:"true" json:"group_id" yaml:"group_id"`
	Glob            bool   `default:"false" json:"glob" yaml:"glob"`
	DeliveryReports bool   `default:"false" json:"reports" yaml:"reports"`

	Codec string `default:"none" json:"codec" yaml:"codec"`

	producer      *kafka.Producer
	consumer      *kafka.Consumer
	deliveryError func(*kafka.Message)
}

// String returns JSON representation
func (sc StreamConfig) String() string {
	b, _ := json.Marshal(&sc)

	return string(b)
}

// SetFlags to install various command-line flag(s)
func (sc *StreamConfig) SetFlags() {
	flag.StringVar(&sc.Brokers, "brokers", sc.Brokers, "Brokers.")
	flag.StringVar(&sc.Topic, "topic", sc.Topic, "Kafka topic.")
	flag.StringVar(&sc.Prefix, "prefix", sc.Prefix, "Stream prefix.")
	flag.StringVar(&sc.Offset, "offset", sc.Offset, "Topic offset.")
	flag.StringVar(&sc.GroupId, "groupid", sc.GroupId, "Group ID")
	flag.StringVar(&sc.Codec, "codec", sc.Codec, "Compression")
	flag.BoolVar(&sc.Glob, "glob", sc.Glob, "Add glob .* to topic")

	envconfig.Process(config.AtsuConfigEnvPrefix, sc)
}

func (sc *StreamConfig) SetDeliveryError(f func(*kafka.Message)) {
	sc.deliveryError = f
}

const SessionTimeoutDefault = 6000 // ms

// consumerDefaults returns a *kafka.ConfigMap with sane defaults
func (sc StreamConfig) consumerDefaults() *kafka.ConfigMap {
	return &kafka.ConfigMap{
		"bootstrap.servers":        sc.Brokers,
		"group.id":                 sc.GroupId,
		"session.timeout.ms":       SessionTimeoutDefault,
		"go.events.channel.enable": true,
		"enable.auto.commit":       false,
		"default.topic.config":     kafka.ConfigMap{"auto.offset.reset": sc.Offset},

		//"auto.commit.interval.ms":  60000,
	}
}

// NewConsumer() creates a new Kafka consumer and subscribes to the underlying topic
func (sc *StreamConfig) NewConsumer(km *kafka.ConfigMap) (*kafka.Consumer, error) {
	if km == nil {
		km = sc.consumerDefaults()
	}
	if c, err := kafka.NewConsumer(km); err == nil {
		topic := sc.FullTopic("")
		if sc.Glob {
			topic := fmt.Sprintf("^%s", sc.FullTopic(""))
			topic = topic + ".*"
		}

		// sanity check broker communications early
		_, err = c.GetMetadata(nil, true, SessionTimeoutDefault)

		if err != nil {
			return nil, err
		}

		c.Subscribe(topic, nil)

		sc.consumer = c

		return c, nil
	} else {
		return nil, err
	}
}

// producerDefaults returns a *kafka.ConfigMap with sane defaults
func (sc StreamConfig) ProducerDefaults() *kafka.ConfigMap {
	if sc.Codec == "" {
		sc.Codec = "none"
	}

	return &kafka.ConfigMap{
		"bootstrap.servers":   sc.Brokers,
		"compression.codec":   sc.Codec,
		"go.delivery.reports": sc.DeliveryReports,
		"session.timeout.ms":  SessionTimeoutDefault,
	}
}

// NewProducer() creates a new Kafka producer
func (sc *StreamConfig) NewProducer(km *kafka.ConfigMap) (*kafka.Producer, error) {
	if km == nil {
		km = sc.ProducerDefaults()
	}
	p, err := kafka.NewProducer(km)
	if err != nil {
		return nil, err
	} else {
		sc.producer = p
	}

	if sc.DeliveryReports {
		if err := sc.deliverReports(); err != nil {
			return p, err
		}
	}

	// sanity check broker communications early
	_, err = p.GetMetadata(nil, true, SessionTimeoutDefault)

	if err != nil {
		return p, err
	}

	return sc.producer, nil
}

func (sc *StreamConfig) SetTopic(topic string) {
	sc.Topic = topic
}

func (sc *StreamConfig) SetPrefix(prefix string) {
	sc.Prefix = prefix
}

func (sc *StreamConfig) SetBrokers(brokers string) {
	sc.Brokers = brokers
}

func (sc *StreamConfig) Close() error {
	if sc.producer != nil {
		sc.producer.Close()
	}
	if sc.consumer != nil {
		if err := sc.consumer.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (sc *StreamConfig) Produce(topic *string, value []byte) error {
	if sc.producer == nil {
		panic("internal failure, no producer set")
	}

	return sc.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: topic,
			Partition: kafka.PartitionAny},
		Value: value}, nil)
}

func (sc StreamConfig) Flush(ms int) int {
	if sc.producer == nil {
		panic("internal failure, no producer set")
	}

	if ms < 1 {
		ms = DefaultFlushInterval
	}
	return sc.producer.Flush(ms)
}

// FullTopic returns prefix.topic (XXX don't like this naming yet)
// if t == "" the sc.Topic will be used
func (sc StreamConfig) FullTopic(t string) string {
	if t == "" {
		t = sc.Topic
	}
	return fmt.Sprintf("%s.%s", sc.Prefix, t)
}

func (sc StreamConfig) deliverReports() error {
	if sc.producer == nil {
		return errors.New("internal failure, no producer set")
	}

	go func() {
		for e := range sc.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil && sc.deliveryError != nil {
					sc.deliveryError(ev)
				}
			}
		}
	}()
	return nil
}

func (sc StreamConfig) ChannelProduce(topic *string, value []byte) {
	if sc.producer == nil {
		panic("internal failure, no producer set")
	}

	sc.producer.ProduceChannel() <- &kafka.Message{TopicPartition: kafka.TopicPartition{Topic: topic, Partition: kafka.PartitionAny}, Value: value}
}

func (sc StreamConfig) GetBrokers() string {
	return sc.Brokers
}

func (sc StreamConfig) GetPrefix() string {
	return sc.Prefix
}

/// XXX Temporary functions to allow more advanced usage
func (sc StreamConfig) GetConsumer() *kafka.Consumer {
	return sc.consumer
}
func (sc StreamConfig) GetProducer() *kafka.Producer {
	return sc.producer
}
