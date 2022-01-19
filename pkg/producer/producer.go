package producer

type Producer interface {
	SendMessage(topic, message string) error
	Close() error
}
