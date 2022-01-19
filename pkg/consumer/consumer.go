package consumer

type Information struct {
	Message string
	Error   error
}

type Consumer interface {
	Consume(topic string) chan Information
	Close() error
}
