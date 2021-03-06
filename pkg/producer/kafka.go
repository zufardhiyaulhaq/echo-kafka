package producer

import (
	"github.com/Shopify/sarama"
	"github.com/golang/protobuf/proto"
	echoproto "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
)

type Kafka struct {
	Producer sarama.SyncProducer
}

func (k *Kafka) SendMessage(topic, message string) error {
	msg := echoproto.Message{
		Message: message,
	}

	bytes, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	producerMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(bytes),
	}

	_, _, err = k.Producer.SendMessage(producerMessage)
	if err != nil {
		return err
	}
	return nil
}

func (k *Kafka) Close() error {
	return k.Producer.Close()
}

func NewKafkaProducer(settings settings.Settings) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(settings.KafkaHosts, config)
	if err != nil {
		return nil, err
	}

	return &Kafka{
		Producer: syncProducer,
	}, nil
}
