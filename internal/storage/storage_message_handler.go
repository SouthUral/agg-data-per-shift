package storage

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type StorageMessageHandler struct {
	dbConn     *PgConn
	incomingCh chan interface{}
	cancel     func()
}

func (s *StorageMessageHandler) GetStorageCh() chan interface{} {
	return s.incomingCh
}

func InitStorageMessageHandler(url string) (*StorageMessageHandler, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())

	s := &StorageMessageHandler{
		cancel:     cancel,
		dbConn:     initPgConn(url, 10),
		incomingCh: make(chan interface{}),
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
	var responseErr error
	mes, err := utils.TypeConversion[mesFromAggMileageHours](message.GetMesage())
	if err != nil {
		err = utils.Wrapper(handlerMesAggMileageHoursError{}, err)
		log.Error(err)
		return
	}

	switch mes.GetType() {
	case restoreShiftDataPerObj:
		log.Infof("принято сообщения от обработчика объекта: %d", mes.GetObjID())
		// обработка сообщения восстановления состояния
		response := s.handlerRestoreShiftDataPerObj(ctx, mes.GetObjID())
		// нужно обработать ошибки
		// возвращется две ошибки, нужно обработать каждую, каждая ошибка проверяется на тип
		typeCriticalErr, err := handlingErrors(response.responseSession.err, response.responseShift.err)
		switch typeCriticalErr {
		case criticalError:
			log.Error(err)
			s.Shutdown(err)
			return
		case modulError:
			log.Error(err)
			responseErr = err
		}
		answer := answerForAggMileageHours{
			shiftData:   response.responseShift.data,
			sessionData: response.responseSession.data,

			err: responseErr,
			// модуль должен возвращать только критические ошибки
			// ошибка связанная с подключением не является критической
			// ошибка связанная с отсутствием значений в rows тоже не является критической
		}
		message.GetChForResponse() <- answer
		log.Infof("Ответ отправлен, ObjID: %d", mes.GetObjID())
	case addNewShiftAndSession:
		// обработка сообщения добавление новых смены и сессии
		// ответом будет id смены и id сессии
		s.handlerAddNewShiftAndSession(mes.GetShiftData(), mes.GetSessionData())

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
	defer log.Infof("Закончена обработка запроса на восстановления данных для объекта: %d", objId)
	var counterResponse int
	var result responceShiftSession
	numResponse := 2 // количество ответов, которые нужно получить

	chShift := makeRequestAndProcessShift(s.dbConn, getLastObjShift, objId)
	chSession := makeRequestAndProcessSession(s.dbConn, getLastObjSession, objId)

	for {
		select {
		case <-ctx.Done():
			return result
		case shiftResponseData := <-chShift:
			log.Info("Принято сообщение chShift")
			result.responseShift = shiftResponseData
			counterResponse++
			if counterResponse == numResponse {
				return result
			}
		case sessionDataResponce := <-chSession:
			log.Info("Принято сообщение chSession")
			result.responseSession = sessionDataResponce
			counterResponse++
			if counterResponse == numResponse {
				return result
			}
		}
	}
}

func (s *StorageMessageHandler) handlerAddNewShiftAndSession(shiftData, sessionData interface{}) error {
	// TODO: нужно преобразовать данные по смене и по сессии в структуры
	// Сформировать запрос в БД, одновременно данные внести не получится, т.к нужен id смены для записи сессии
	// сначала инсертить смену, получить id затем инсертить сессию
	shiftStructData := shiftDataFromModule{}
	sessionStructData := sessionDataFromModule{}

	err := shiftStructData.loadData(shiftData)
	if err != nil {
		return err
	}
	err = sessionStructData.loadData(sessionData)
	if err != nil {
		return err
	}

	return err

}

// функция отправляет запрос в БД, получает ответ, конвертирует его в переданный тип, и из типа конвертирует его в json.
// Функция обрабатывает ответ в одну строку.
func makeRequestAndProcessSession(dbConn *PgConn, request string, objectId int) chan responseSessionDB {
	responseCh := make(chan responseSessionDB)
	go func() {
		response, err := dbConn.QueryDB(request, objectId)
		if err != nil {
			log.Error(err)
			responseCh <- responseSessionDB{err: err}
			return
		}

		sessionData, engData, mileage, mileageGPS, err := convertingQueryRowResultIntoStructure[RowSessionObjData](response)
		if err != nil {
			log.Error(err)
			responseCh <- responseSessionDB{err: err}
			return
		}

		sessionData.engHours = engData
		sessionData.mileageData = mileage
		sessionData.mileageGPSData = mileageGPS

		responseCh <- responseSessionDB{data: sessionData, err: err}
		defer response.Close()
	}()
	return responseCh
}

func makeRequestAndProcessShift(dbConn *PgConn, request string, objectId int) chan responseShiftDB {
	responseCh := make(chan responseShiftDB)
	go func() {
		response, err := dbConn.QueryDB(request, objectId)
		if err != nil {
			log.Error(err)
			responseCh <- responseShiftDB{err: err}
			return
		}

		shiftData, engData, mileage, mileageGPS, err := convertingQueryRowResultIntoStructure[RowShiftObjData](response)
		if err != nil {
			log.Error(err)
			responseCh <- responseShiftDB{err: err}
			return
		}

		shiftData.engHours = engData
		shiftData.mileageData = mileage
		shiftData.mileageGPSData = mileageGPS

		responseCh <- responseShiftDB{data: shiftData, err: err}
		defer response.Close()
	}()
	return responseCh
}

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.Shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
}
