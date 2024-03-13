package core

import (
	"fmt"
	"time"

	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	aggMileage "agg-data-per-shift/internal/services/aggMileageHours"
	storage "agg-data-per-shift/internal/storage"

	log "github.com/sirupsen/logrus"
)

// функция запускает сервис
func StartService() {
	envs, err := getEnvs()
	if err != nil {
		log.Error(err)
		return
	}
	// envRabbit := "amqp://test_user:rmpassword@localhost:5672/asd"
	// pgUrl := "postgres://kovalenko:kovalenko@localhost:5435/report_bd"
	// nameConsumer := "test_consumer"
	// stream := "messages_stream"

	rb := amqp.InitRabbit(envs.rbEnvs, 30)
	st, ctxSt := storage.InitStorageMessageHandler(envs.pgEnvs)
	ctxRb := rb.StartRb()

	ag, ctxAg := aggMileage.InitEventRouter(st.GetStorageCh(), 10)

	for {
		select {
		case <-ctxRb.Done():
			ag.Shudown(fmt.Errorf("rabbitMQ закончил работу"))
			st.Shutdown(fmt.Errorf("rabbitMQ закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case <-ctxAg.Done():
			rb.RabbitShutdown(fmt.Errorf("роутер закончил работу"))
			st.Shutdown(fmt.Errorf("роутер закончил работу"))
			time.Sleep(5 * time.Second)
			return
		case <-ctxSt.Done():
			ag.Shudown(fmt.Errorf("ошибка storage"))
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
				st.GetStorageCh() <- transportStruct{
					sender:         "amqp",
					mesage:         message.GetTypeMsg(),
					reverseChannel: message.GetReverceCh(),
				}
			case "InputMSG":
				log.Debugf("offset: %d", message.GetOffset())
				message.GetReverceCh() <- answerEvent{}
				ag.EventReception(message)

			default:
				rb.RabbitShutdown(fmt.Errorf("неизвестный тип событий"))
				return
			}
		}
	}
}
