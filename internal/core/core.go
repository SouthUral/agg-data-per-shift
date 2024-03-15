package core

import (
	"fmt"
	"time"

	amqp "agg-data-per-shift/internal/amqp/amqp_client"
	aggMileage "agg-data-per-shift/internal/services/aggMileageHours"
	storage "agg-data-per-shift/internal/storage"
	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

// функция запускает сервис
func StartService() {
	// var targetCount int = 5000

	envs, err := getEnvs()
	if err != nil {
		log.Error(err)
		return
	}

	timeMeter := utils.InitProcessingTimeMeter()
	defer timeMeter.GetAnaliticsForAllProcess()
	defer timeMeter.Shudown()
	defer time.Sleep(5 * time.Second)

	// инициализация подключения к базам
	rb := amqp.InitRabbit(envs.rbEnvs, 30)
	db, dbCtx := storage.InitPgConn(envs.pgEnvs, 1000, 1000, 20)

	// инициализация логики
	st, ctxSt := storage.InitStorageMessageHandler(db)
	ag, ctxAg := aggMileage.InitEventRouter(st.GetStorageCh(), 10, timeMeter)

	// начало прослушивания очереди
	ctxRb := rb.StartRb()

	for {
		select {
		case <-dbCtx.Done():
			rb.RabbitShutdown(fmt.Errorf("db закончил работу"))
			ag.Shudown(fmt.Errorf("db закончил работу"))
			st.Shutdown(fmt.Errorf("db закончил работу"))
			return
		case <-ctxRb.Done():
			ag.Shudown(fmt.Errorf("rabbitMQ закончил работу"))
			st.Shutdown(fmt.Errorf("rabbitMQ закончил работу"))
			return
		case <-ctxAg.Done():
			rb.RabbitShutdown(fmt.Errorf("роутер закончил работу"))
			st.Shutdown(fmt.Errorf("роутер закончил работу"))
			return
		case <-ctxSt.Done():
			ag.Shudown(fmt.Errorf("ошибка storage"))
			rb.RabbitShutdown(fmt.Errorf("ошибка storage"))
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
				// counter, _ := timeMeter.GetCounterOnKey("eventHandling")
				// if counter >= targetCount {
				// 	ag.Shudown(fmt.Errorf("цель достигнута"))
				// 	st.Shutdown(fmt.Errorf("цель достигнута"))
				// 	rb.RabbitShutdown(fmt.Errorf("цель достигнута"))
				// 	log.Infof("counter = %d", counter)
				// 	return
				// }
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
