package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
)

type Kafka struct {
	Consumer sarama.Consumer
}

func (k Kafka) Consume(topic string) chan Information {
	information := make(chan Information)
	partitions, _ := k.Consumer.Partitions(topic)

	for _, partition := range partitions {
		consumer, err := k.Consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)

		if nil != err {
			information <- Information{
				Error: err,
			}

			return information
		}

		go func(topic string, consumer sarama.PartitionConsumer) {
			for {
				select {
				case consumerError := <-consumer.Errors():
					information <- Information{
						Error: consumerError,
					}

				case message := <-consumer.Messages():
					information <- Information{
						Message: string(message.Value),
					}
				}
			}
		}(topic, consumer)
	}

	return information
}

func (k *Kafka) Close() error {
	return k.Consumer.Close()
}

func NewKafkaConsumer(settings settings.Settings) (Consumer, error) {
	config := sarama.NewConfig()

	consumerClient, err := sarama.NewConsumer(settings.KafkaHosts, config)
	if err != nil {
		return nil, err
	}

	return &Kafka{
		Consumer: consumerClient,
	}, nil
}
