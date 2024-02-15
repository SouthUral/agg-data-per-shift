package streamclient

type connectStreamError struct {
}

func (e connectStreamError) Error() string {
	return "connect stream RabbitMQ error"
}
