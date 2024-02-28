package storage

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type StorageMessageHandler struct {
	dbConn     *pgConn
	incomingCh chan interface{}
	cancel     func()
}

func InitPsql(url string, waitingTime int) (*StorageMessageHandler, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &StorageMessageHandler{
		cancel: cancel,
		dbConn: initPgConn(url, 10),
	}

	go s.listenAndServe(ctx)

	return s, ctx
}

// процесс получения и обработки сообщений от других модулей
func (s *StorageMessageHandler) listenAndServe(ctx context.Context) {
	defer log.Warning("listenAndServe is closed")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.incomingCh:
			message, err := utils.TypeConversion[trunsportMes](msg)
			if err != nil {
				err = utils.Wrapper(listenAndServeError{}, err)
				s.Shutdown(err)
			}
			s.handleRequests(message)
		}
	}

}

// обработчик запросов
func (s *StorageMessageHandler) handleRequests(message trunsportMes) {
	switch message.GetSender() {
	case aggMileageHours:
		go s.handlerMesAggMileageHours(message)
	default:
		log.Errorf("unknown sender: %s", message.GetSender())
	}
}

// метод обрабатывает сообщение от модуля aggMileageHours
func (s *StorageMessageHandler) handlerMesAggMileageHours(message trunsportMes) {
	mes, err := utils.TypeConversion[mesFromAggMileageHours](message.GetMesage())
	if err != nil {
		err = utils.Wrapper(handlerMesAggMileageHoursError{}, err)
		log.Error(err)
		return
	}

	switch mes.GetType() {
	case restoreShiftDataPerObj:
		// обработка сообщения восстановления состояния
		// TODO: нужно отправить два запроса в БД (чтобы забрать данные смены и сессии)
		// дождаться выполнения обоих запросов
		// ответом будет строка из таблицы смен и строка из таблицы сессий
		// ответ от БД преобразовать в структуру (под интерфейс) и отправить обратно
	case addNewShiftAndSession:
		// обработка сообщения добавление новых смены и сессии
		// ответом будет id смены и id сессии

	case updateShiftAndAddNewSession:
		// обработка сообщения, обновление смены и добавление новой сессии
		// ответом будет id сессии

	case updateShiftAndSession:
		// обработка сообщения, обновления смены и сессии
		// ответ не нужен (пока)

	default:
		log.Error("an unknown message type was received")
	}
}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}
