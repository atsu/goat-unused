package stream

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type StreamConsumer interface {
	Start(*StreamConfig, interface{}) error
	Message(*kafka.Message) error // error != nil, stop consumer
	Interval(time.Time) error     // error != nil, stop consumer
	Timeout(time.Time, bool) bool // bool != false, stop consumer
	Error(kafka.Error) bool       // bool != false, stop consumer
	Process() (bool, error)
	Finish() error // This value will be returned to the caller

	DoneCh() <-chan bool
}

func (sc *StreamConfig) Consume(consumer StreamConsumer, config interface{}) error {
	// Connect to Kafka
	c, err := sc.NewConsumer(nil)
	if err != nil {
		return err
	}
	defer c.Close()

	last := 0
	stalled := false

	timeTick := &time.Ticker{}
	// sc.timeout == 0 means NEVER timeout
	if sc.Timeout > 0 {
		timeTick = time.NewTicker(sc.Timeout)
		defer timeTick.Stop()
	}

	// sc.interval == 0 means NO INTERVAL
	intTick := &time.Ticker{}
	if sc.Interval > 0 {
		intTick = time.NewTicker(sc.Interval)
		defer intTick.Stop()
	}

	if err := consumer.Start(sc, config); err != nil {
		return err
	}

	run := true
	for run {
		select {
		case ev := <-c.Events():
			switch e := ev.(type) {
			case *kafka.Message:
				sc.Messages += 1
				sc.Bytes += len(e.Value)

				if err := consumer.Message(e); err != nil {
					run = false
				}
			case kafka.Error:
				// Consumer must handle all errors, including EOF
				if consumer.Error(e) {
					run = false
				}
			}
		case t := <-intTick.C:
			if err := consumer.Interval(t); err != nil {
				run = false
			}
		case t := <-timeTick.C:
			if last == sc.Messages {
				stalled = true
			} else {
				stalled = false
			}
			last = sc.Messages

			if consumer.Timeout(t, stalled) {
				run = false
			}
		case <-consumer.DoneCh():
			run = false
		}

		if stop, err := consumer.Process(); err != nil || stop {
			run = false
		}

	}
	return consumer.Finish()
}
