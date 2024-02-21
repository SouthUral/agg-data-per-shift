package main

import (
	"fmt"
	"os"
	"time"

	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	aggMileage "agg-data-per-shift/internal/services/aggMileageHours"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {

	envRabbit := "amqp://test_user:rmpassword@localhost:5672/asd"
	nameConsumer := "test_consumer"
	stream := "messages_stream"

	rb := amqp.InitRabbit(envRabbit, stream, nameConsumer, 30)
	ctxRb := rb.StartRb()

	aggEventRouter, ctxEventRouter := aggMileage.InitEventRouter()

	for {
		select {
		case <-ctxRb.Done():
			aggEventRouter.Shudown(fmt.Errorf("rabbitMQ закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case <-ctxEventRouter.Done():
			rb.RabbitShutdown(fmt.Errorf("роутер закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case msg := <-rb.GetChan():
			message, ok := msg.(msgEvent)
			if !ok {
				rb.RabbitShutdown(fmt.Errorf("ошибка приведения типа"))
				return
			}
			switch message.GetTypeMsg() {
			case "GetOffset":
				message.GetReverceCh() <- answerEvent{
					offset: 0,
				}
			case "InputMSG":
				log.Debugf("offset: %d", message.GetOffset())
				message.GetReverceCh() <- answerEvent{}
				aggEventRouter.EventReception(message)

			default:
				rb.RabbitShutdown(fmt.Errorf("неизвестный тип событий"))
				return
			}
		}
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
