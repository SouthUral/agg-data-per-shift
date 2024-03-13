package main

import (
	"fmt"
	"os"
	"time"

	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	aggMileage "agg-data-per-shift/internal/services/aggMileageHours"
	storage "agg-data-per-shift/internal/storage"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {

	envRabbit := "amqp://test_user:rmpassword@localhost:5672/asd"
	pgUrl := "postgres://kovalenko:kovalenko@localhost:5435/report_bd"
	nameConsumer := "test_consumer"
	stream := "messages_stream"

	rb := amqp.InitRabbit(envRabbit, stream, nameConsumer, 30)
	st, ctxStorage := storage.InitStorageMessageHandler(pgUrl)
	ctxRb := rb.StartRb()

	aggEventRouter, ctxEventRouter := aggMileage.InitEventRouter(st.GetStorageCh(), 10)

	for {
		select {
		case <-ctxRb.Done():
			aggEventRouter.Shudown(fmt.Errorf("rabbitMQ закончил работу"))
			st.Shutdown(fmt.Errorf("rabbitMQ закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case <-ctxEventRouter.Done():
			rb.RabbitShutdown(fmt.Errorf("роутер закончил работу"))
			st.Shutdown(fmt.Errorf("роутер закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case <-ctxStorage.Done():
			aggEventRouter.Shudown(fmt.Errorf("ошибка storage"))
			rb.RabbitShutdown(fmt.Errorf("ошибка storage"))
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
				// message.GetReverceCh() <- answerEvent{
				// 	// offset: 500000,
				// 	// offset: 0,
				// 	offset: 646644,
				// }
				// TODO: нужно сделать отправку сообщения в storage
				st.GetStorageCh() <- transportStruct{
					sender:         "amqp",
					mesage:         message.GetTypeMsg(),
					reverseChannel: message.GetReverceCh(),
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

// TODO: нужно переделать все под универсальную структуру
type trunsportMes interface {
	GetSender() string                  // имя модуля отправителя сообщения
	GetMesage() interface{}             // сообщение от модуля
	GetChForResponse() chan interface{} // метод для отправки ответа
}

// транспортная структура (универсальный интерфейс)
type transportStruct struct {
	sender         string           // модуль отправитель сообщения
	mesage         interface{}      // сообщение отправителя
	reverseChannel chan interface{} // канал для отправки ответа от модуля storage
}

func (t transportStruct) GetSender() string {
	return t.sender
}

func (t transportStruct) GetMesage() interface{} {
	return t.mesage
}

// метод для отправки ответа от модуля storage
func (t transportStruct) GetChForResponse() chan interface{} {
	return t.reverseChannel
}

// ответное сообщение
type answerEvent struct {
	offset int
}

func (a answerEvent) GetOffset() int {
	return a.offset
}
