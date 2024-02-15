package main

import (
	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	envRabbit := "amqp://test_user:rmpassword@localhost:5672/asd"
	nameConsumer := "test_consumer"
	stream := "messages_stream"

	rb := amqp.InitRabbit(envRabbit, stream, nameConsumer, 30)
	ctx := rb.StartRb()

	ch := rb.GetChan()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:
			message, ok := msg.(msgEvent)
			if !ok {
				rb.RabbitShutdown(fmt.Errorf("ошибка приведения типа"))
				return
			}
			switch message.GetTypeMsg() {
			case "GetOffset":
				message.GetReverceCh() <- answerEvent{
					offset: 10,
				}
			case "InputMSG":
				log.Infof("message: %v, offset: %d", message.GetMsg(), message.GetOffset())
				message.GetReverceCh() <- answerEvent{}
			default:
				rb.RabbitShutdown(fmt.Errorf("неизвестный тип событий"))
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
}

type msgEvent interface {
	GetTypeMsg() string
	GetReverceCh() chan interface{}
	GetMsg() []byte
	GetOffset() int64
}

// ответное сообщение
type answerEvent struct {
	offset int
}

func (a answerEvent) GetOffset() int {
	return a.offset
}
