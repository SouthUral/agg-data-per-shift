package main

import (
	amqp "agg-data-per-shift/internal/amqp/stream_client"
	"time"
)

func main() {
	envRabbit := "rabbitmq-stream://test_user:rmpassword@localhost:5552/asd"
	nameConsumer := "test_consumer"
	stream := "messages_stream"

	rb := amqp.InitRabbit(envRabbit, nameConsumer, stream)
	time.Sleep(10 * time.Minute)
	rb.Shutdown()
}
