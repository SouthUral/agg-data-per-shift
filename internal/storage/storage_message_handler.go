package storage

import (
	"context"
	"encoding/json"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type StorageMessageHandler struct {
	dbConn     *pgConn
	incomingCh chan interface{}
	cancel     func()
}

func InitStorageMessageHandler(url string, waitingTime int) (*StorageMessageHandler, context.Context) {
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
			s.handleRequests(ctx, message)
		}
	}

}

// обработчик запросов
func (s *StorageMessageHandler) handleRequests(ctx context.Context, message trunsportMes) {
	switch message.GetSender() {
	case aggMileageHours:
		go s.handlerMesAggMileageHours(ctx, message)
	default:
		log.Errorf("unknown sender: %s", message.GetSender())
	}
}

// метод обрабатывает сообщение от модуля aggMileageHours
func (s *StorageMessageHandler) handlerMesAggMileageHours(ctx context.Context, message trunsportMes) {
	mes, err := utils.TypeConversion[mesFromAggMileageHours](message.GetMesage())
	if err != nil {
		err = utils.Wrapper(handlerMesAggMileageHoursError{}, err)
		log.Error(err)
		return
	}

	switch mes.GetType() {
	case restoreShiftDataPerObj:
		// обработка сообщения восстановления состояния
		response := s.handlerRestoreShiftDataPerObj(ctx, mes.GetObjID())
		answer := answerForAggMileageHours{
			shiftData:   response.responseShift.data,
			sessionData: response.responseSession.data,
			err:         response.handlingErrors(),
		}
		message.SendAnswer(answer)
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

// метод производит два ассинхронных запроса на получение строк из БД
func (s *StorageMessageHandler) handlerRestoreShiftDataPerObj(ctx context.Context, objId int) responceShiftSession {
	var counterResponse int
	var result responceShiftSession
	numResponse := 2 // количество ответов, которые нужно получить

	chShift := make(chan responseDB)
	chSession := make(chan responseDB)

	go func() {
		response, err := s.dbConn.queryDB(getLastObjShift, objId)
		if err != nil {
			log.Error(err)
			return
		}
		shiftDBdata, err := convertingQueryRowResultIntoStructure[RowShiftObjData](response)
		if err != nil {
			log.Error(err)
			return
		}
		b, err := json.Marshal(shiftDBdata)
		if err != nil {
			log.Error(err)
			return
		}

		defer func() { chShift <- responseDB{data: b, err: err} }()
	}()

	go func() {
		response, err := s.dbConn.queryDB(getLastObjSession, objId)
		if err != nil {
			log.Error(err)
			return
		}
		sessionDBData, err := convertingQueryRowResultIntoStructure[RowSessionObjData](response)
		if err != nil {
			log.Error(err)
			return
		}
		b, err := json.Marshal(sessionDBData)
		if err != nil {
			log.Error(err)
		}

		defer func() { chSession <- responseDB{data: b, err: err} }()
	}()

	for {

		select {
		case <-ctx.Done():
			return result
		case shiftResponseData := <-chShift:
			result.responseShift = shiftResponseData
			counterResponse++
			if counterResponse == numResponse {
				return result
			}
		case sessionDataResponce := <-chSession:
			result.responseSession = sessionDataResponce
			counterResponse++
			if counterResponse == numResponse {
				return result
			}
		}
	}

}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}
