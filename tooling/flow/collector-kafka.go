package flow

import (
	"log"

	"github.com/Shopify/sarama"
	"github.com/adrianco/spigo/tooling/archaius"
)

// defaultKafkaTopic sets the standard Kafka topic our Collector will publish
// on. The default topic for zipkin-receiver-kafka is "zipkin", see:
// https://github.com/openzipkin/zipkin/tree/master/zipkin-receiver-kafka
const defaultKafkaTopic = "zipkin"

// KafkaCollector implements Collector by publishing spans to a Kafka
// broker.
type KafkaCollector struct {
	producer sarama.AsyncProducer
	topic    string
}

// NewKafkaCollector returns a new Kafka-backed Collector. addrs should be a
// slice of TCP endpoints of the form "host:port".
func NewKafkaCollector(addrs []string) (*KafkaCollector, error) {
	c := &KafkaCollector{
		topic: defaultKafkaTopic,
	}

	p, err := sarama.NewAsyncProducer(addrs, nil)
	if err != nil {
		return nil, err
	}
	c.producer = p

	if archaius.Conf.Msglog {
		go c.logErrors()
	}

	return c, nil
}

func (c *KafkaCollector) logErrors() {
	for pe := range c.producer.Errors() {
		log.Printf("Kafka Error: %#v\n", pe)
	}
}

// Collect implements Collector.
func (c *KafkaCollector) Collect(s []byte) {
	c.producer.Input() <- &sarama.ProducerMessage{
		Topic: c.topic,
		Key:   nil,
		Value: sarama.ByteEncoder(s),
	}
}

// Close implements Collector.
func (c *KafkaCollector) Close() error {
	return c.producer.Close()
}
