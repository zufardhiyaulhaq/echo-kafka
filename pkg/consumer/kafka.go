package consumer

import (
	"context"

	"github.com/Shopify/sarama"
	echoproto "github.com/zufardhiyaulhaq/echo-grpc/proto"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
	"google.golang.org/protobuf/proto"
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
					msg := &echoproto.Message{}
					err := proto.Unmarshal(message.Value, msg)
					if err != nil {
						information <- Information{
							Error: err,
						}
					} else {
						information <- Information{
							Message: string(msg.Message),
						}
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

type KafkaConsumerGroup struct {
	ConsumerGroup sarama.ConsumerGroup
}

func (k *KafkaConsumerGroup) Close() error {
	return k.ConsumerGroup.Close()
}

func (k KafkaConsumerGroup) Consume(topic string) chan Information {
	topics := []string{topic}
	ctx := context.Background()

	consumer := SaramaConsumer{
		information: make(chan Information),
	}

	go func() {
		for {
			err := k.ConsumerGroup.Consume(ctx, topics, &consumer)
			if err != nil {
				consumer.information <- Information{
					Error: err,
				}
			}
		}
	}()

	return consumer.information
}

type SaramaConsumer struct {
	information chan Information
}

func (consumer *SaramaConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *SaramaConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (consumer *SaramaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		consumer.information <- Information{
			Message: string(message.Value),
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func NewKafkaConsumerGroup(settings settings.Settings) (Consumer, error) {
	config := sarama.NewConfig()

	consumerClient, err := sarama.NewConsumerGroup(settings.KafkaHosts, settings.KafkaGroupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumerGroup{
		ConsumerGroup: consumerClient,
	}, nil
}
