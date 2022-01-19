package settings

import (
	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	EchoPort   string   `envconfig:"ECHO_PORT" default:"5000"`
	HTTPPort   string   `envconfig:"HTTP_PORT" default:"80"`
	KafkaHosts []string `envconfig:"KAFKA_HOSTS"`
	KafkaTopic string   `envconfig:"KAFKA_TOPIC", default:"echo-kafka"`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
