package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/consumer"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/producer"
	"github.com/zufardhiyaulhaq/echo-kafka/pkg/settings"
)

func main() {

	settings, err := settings.NewSettings()
	if err != nil {
		panic(err.Error())
	}

	log.Info().Msg("creating kafka producer client")
	producerClient, err := producer.NewKafkaProducer(settings)
	if err != nil {
		panic(err.Error())
	}

	var consumerClient consumer.Consumer

	if settings.KafkaEnableConsumerGroup {
		log.Info().Msg("creating kafka consumer group client")
		consumerClient, err = consumer.NewKafkaConsumerGroup(settings)
		if err != nil {
			panic(err.Error())
		}
	} else {
		log.Info().Msg("creating kafka consumer client")
		consumerClient, err = consumer.NewKafkaConsumer(settings)
		if err != nil {
			panic(err.Error())
		}
	}

	wg := new(sync.WaitGroup)
	wg.Add(3)

	log.Info().Msg("starting server")
	server := NewServer(settings, producerClient)

	go func() {
		log.Info().Msg("starting HTTP server")
		server.ServeHTTP()
		wg.Done()
	}()

	go func() {
		log.Info().Msg("starting echo server")
		server.ServeEcho()
		wg.Done()
	}()

	go func() {
		log.Info().Msg("starting kafka consumer client")
		information := consumerClient.Consume(settings.KafkaTopic)

		termChan := make(chan os.Signal)
		signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT, os.Interrupt)

		doneCh := make(chan struct{})
		go func() {
			for {
				select {
				case msg := <-information:
					if msg.Error != nil {
						log.Info().Msgf("error consuming: ", string(msg.Error.Error()))
					} else {
						log.Info().Msgf("consuming: ", string(msg.Message))
					}
				case <-termChan:
					log.Info().Msg("intrupted")
					doneCh <- struct{}{}
				}

			}
		}()
		<-doneCh

		wg.Done()
	}()

	wg.Wait()

	defer func() {
		log.Info().Msg("closing kafka producer")
		if err = producerClient.Close(); err != nil {
			panic(err)
		}

		log.Info().Msg("closing kafka consumer")
		if err = consumerClient.Close(); err != nil {
			panic(err)
		}
	}()
}
