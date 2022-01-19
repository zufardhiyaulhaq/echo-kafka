package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
)

type KafkaConsumer struct {
	Client   sarama.Client
	Consumer sarama.Consumer
}

func (k KafkaConsumer) Consume(topic string) chan Information {
	information := make(chan Information)
	partitions, _ := k.Consumer.Partitions(topic)

	for _, partition := range partitions {
		offset, err := k.Client.GetOffset(topic, partition, sarama.ReceiveTime)
		if err != nil {
			information <- Information{
				Error: err,
			}

			return information
		}

		consumer, err := k.Consumer.ConsumePartition(topic, partition, offset)
		if err != nil {
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

func (k *KafkaConsumer) Close() error {
	return k.Consumer.Close()
}

func NewKafkaConsumer(settings settings.Settings) (Consumer, error) {
	config := sarama.NewConfig()

	consumerClient, err := sarama.NewConsumer(settings.KafkaHosts, config)
	if err != nil {
		return nil, err
	}

	client, err := sarama.NewClient(settings.KafkaHosts, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{
		Consumer: consumerClient,
		Client:   client,
	}, nil
}
