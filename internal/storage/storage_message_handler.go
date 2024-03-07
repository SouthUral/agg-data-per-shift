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

// метод прекращает работу модуля psql (завершает все активные горутины, разрывает коннект с БД)
func (s *StorageMessageHandler) Shutdown(err error) {
	log.Errorf("psql is terminated for a reason: %s", err)
	s.cancel()
	s.dbConn.Shutdown(utils.Wrapper(fmt.Errorf("psql is terminated"), err))
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
		log.Debugf("восстановление состояния для объекта: %d", mes.GetObjID())
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
		responce := answerForAggMileageHours{
			shiftData:   response.responseShift.data,
			sessionData: response.responseSession.data,

			err: responseErr,
		}
		message.GetChForResponse() <- responce
		log.Infof("Ответ по восстановлению состояния отправлен, ObjID: %d", mes.GetObjID())
	case addNewShiftAndSession:
		log.Debugf("Добавление новых записей смены и сессии для объекта: %d", mes.GetObjID())
		shiftId, sessionId, err := s.handlerAddNewShiftAndSession(mes.GetObjID(), mes.GetShiftData(), mes.GetSessionData())
		// возможно придется добавить обработку ошибок
		if err != nil {
			log.Error(err)
			s.Shutdown(err)
			return
		}
		message.GetChForResponse() <- responceAggMileageHoursAddNewShiftAndSession{
			shiftId:   shiftId,
			sessionId: sessionId,
		}
	case updateShiftAndAddNewSession:
		log.Debugf("Добавление новой записи сессии, обновление записи смены для объекта : %d", mes.GetObjID())
		sessionId, err := s.handlerUpdateShiftAndAddNewSession(mes.GetObjID(), mes.GetShiftData(), mes.GetSessionData())
		if err != nil {
			log.Error(err)
			s.Shutdown(err)
			return
		}
		message.GetChForResponse() <- responceAggMileageHoursAddNewShiftAndSession{
			sessionId: sessionId,
		}
	case updateShiftAndSession:
		err := s.handlerUpdateShiftAndSession(mes.GetShiftData(), mes.GetSessionData())
		if err != nil {
			log.Error(err)
			s.Shutdown(err)
			return
		}
		// пока это бесполезное действие, т.к. ошибка все равно не отправится обратно, т.к. программа завершится
		message.GetChForResponse() <- responceAggMileageHoursAddNewShiftAndSession{
			err: err,
		}
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

func (s *StorageMessageHandler) handlerAddNewShiftAndSession(objId int, shiftData, sessionData interface{}) (int, int, error) {
	var (
		shifId, sessionId int
	)

	shiftStructData := shiftDataFromModule{}
	sessionStructData := sessionDataFromModule{}

	err := shiftStructData.loadData(shiftData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации смены"), err)
		return shifId, sessionId, err
	}
	err = sessionStructData.loadData(sessionData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации сессии"), err)
		return shifId, sessionId, err
	}
	err = s.makeRquestAddNewShift(shiftStructData, objId, &shifId)
	if err != nil {
		return shifId, sessionId, err
	}
	err = s.makeRquestAddNewSession(sessionStructData, objId, shifId, &sessionId)
	if err != nil {
		return shifId, sessionId, err
	}
	return shifId, sessionId, err
}

func (s *StorageMessageHandler) handlerUpdateShiftAndAddNewSession(objId int, shiftData, sessionData interface{}) (int, error) {
	var (
		sessionId int
	)

	shiftStructData := shiftDataFromModule{}
	sessionStructData := sessionDataFromModule{}

	err := shiftStructData.loadData(shiftData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации смены"), err)
		return sessionId, err
	}
	err = sessionStructData.loadData(sessionData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации сессии"), err)
		return sessionId, err
	}
	err = s.makeRquestAddNewSession(sessionStructData, objId, shiftStructData.mainData.GetShiftId(), &sessionId)
	if err != nil {
		return sessionId, err
	}
	err = s.makeRequestUpdateShift(shiftStructData)
	if err != nil {
		return sessionId, err
	}
	return sessionId, err
}

func (s *StorageMessageHandler) handlerUpdateShiftAndSession(shiftData, sessionData interface{}) error {
	shiftStructData := shiftDataFromModule{}
	sessionStructData := sessionDataFromModule{}

	err := shiftStructData.loadData(shiftData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации смены"), err)
		return err
	}
	err = sessionStructData.loadData(sessionData)
	if err != nil {
		err = utils.Wrapper(fmt.Errorf("ошибка конвертации сессии"), err)
		return err
	}
	err = s.makeRequestUpdateShift(shiftStructData)
	if err != nil {
		return err
	}
	err = s.makeRequestUpdateSession(sessionStructData)
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

		sessionData, err := converQuery[RowSessionObjData](response)
		if err != nil {
			log.Error(err)
			responseCh <- responseSessionDB{err: err}
			return
		}

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

		shiftData, err := converQuery[RowShiftObjData](response)
		if err != nil {
			log.Error(err)
			responseCh <- responseShiftDB{err: err}
			return
		}

		// shiftData.engHours = engData
		// shiftData.mileageData = mileage
		// shiftData.mileageGPSData = mileageGPS

		responseCh <- responseShiftDB{data: shiftData, err: err}
		defer response.Close()
	}()
	return responseCh
}

// метод делает запрос в БД на добавлении новой смены в таблицу
func (s *StorageMessageHandler) makeRquestAddNewShift(shiftStructData shiftDataFromModule, objId int, shifId *int) error {
	err := s.dbConn.QueryRowDB(addNewShift,
		shiftStructData.mainData.GetShiftNum(),
		objId,
		shiftStructData.mainData.GetShiftDateStart(),
		shiftStructData.mainData.GetShiftDateEnd(),
		shiftStructData.mainData.GetShiftDate(),
		shiftStructData.mainData.GetUpdatedTime(),
		shiftStructData.mainData.GetOffset(),
		shiftStructData.mainData.GetCurrentDriverId(),
		shiftStructData.mainData.GetStatusLoaded(),
		shiftStructData.engHoursData.GetEngHoursStart(),
		shiftStructData.engHoursData.GetEngHoursCurrent(),
		shiftStructData.engHoursData.GetEngHoursEnd(),
		shiftStructData.mileageData.GetMileageStart(),
		shiftStructData.mileageData.GetMileageCurrent(),
		shiftStructData.mileageData.GetMileageEnd(),
		shiftStructData.mileageData.GetMileageLoaded(),
		shiftStructData.mileageData.GetMileageAtBeginningOfLoading(),
		shiftStructData.mileageData.GetMileageEmpty(),
		shiftStructData.mileageGPSData.GetMileageStart(),
		shiftStructData.mileageGPSData.GetMileageCurrent(),
		shiftStructData.mileageGPSData.GetMileageEnd(),
		shiftStructData.mileageGPSData.GetMileageLoaded(),
		shiftStructData.mileageGPSData.GetMileageAtBeginningOfLoading(),
		shiftStructData.mileageGPSData.GetMileageEmpty(),
	).Scan(shifId)
	if err == nil {
		log.Debugf("объект смены записан в БД id смены: %d", shifId)
	}
	return err
}

// метод делает запрос в БД на добавление новой сессии в таблицу
func (s *StorageMessageHandler) makeRquestAddNewSession(sessionStructData sessionDataFromModule, objId, shifId int, sessionId *int) error {
	err := s.dbConn.QueryRowDB(addNewSession,
		shifId,
		objId,
		sessionStructData.mainData.GetDriverId(),
		sessionStructData.mainData.GetOffset(),
		sessionStructData.mainData.GetTimeStartSession(),
		sessionStructData.mainData.GetTimeUpdateSession(),
		sessionStructData.mainData.GetAvSpeed(),
		sessionStructData.engHoursData.GetEngHoursStart(),
		sessionStructData.engHoursData.GetEngHoursCurrent(),
		sessionStructData.engHoursData.GetEngHoursEnd(),
		sessionStructData.mileageData.GetMileageStart(),
		sessionStructData.mileageData.GetMileageCurrent(),
		sessionStructData.mileageData.GetMileageEnd(),
		sessionStructData.mileageData.GetMileageLoaded(),
		sessionStructData.mileageData.GetMileageAtBeginningOfLoading(),
		sessionStructData.mileageData.GetMileageEmpty(),
		sessionStructData.mileageGPSData.GetMileageStart(),
		sessionStructData.mileageGPSData.GetMileageCurrent(),
		sessionStructData.mileageGPSData.GetMileageEnd(),
		sessionStructData.mileageGPSData.GetMileageLoaded(),
		sessionStructData.mileageGPSData.GetMileageAtBeginningOfLoading(),
		sessionStructData.mileageGPSData.GetMileageEmpty(),
	).Scan(&sessionId)
	if err == nil {
		log.Debugf("объект cессии записан в БД id cессии: %d", sessionId)
	}
	return err
}

// метод делает запрос в БД на обновление смены в таблице
func (s *StorageMessageHandler) makeRequestUpdateShift(shiftStructData shiftDataFromModule) error {
	err := s.dbConn.ExecQuery(updateShift,
		shiftStructData.mainData.GetShiftId(),
		shiftStructData.mainData.GetShiftDateEnd(),
		shiftStructData.mainData.GetUpdatedTime(),
		shiftStructData.mainData.GetOffset(),
		shiftStructData.mainData.GetCurrentDriverId(),
		shiftStructData.mainData.GetStatusLoaded(),
		shiftStructData.engHoursData.GetEngHoursStart(),
		shiftStructData.engHoursData.GetEngHoursCurrent(),
		shiftStructData.engHoursData.GetEngHoursEnd(),
		shiftStructData.mileageData.GetMileageStart(),
		shiftStructData.mileageData.GetMileageCurrent(),
		shiftStructData.mileageData.GetMileageEnd(),
		shiftStructData.mileageData.GetMileageLoaded(),
		shiftStructData.mileageData.GetMileageAtBeginningOfLoading(),
		shiftStructData.mileageData.GetMileageEmpty(),
		shiftStructData.mileageGPSData.GetMileageStart(),
		shiftStructData.mileageGPSData.GetMileageCurrent(),
		shiftStructData.mileageGPSData.GetMileageEnd(),
		shiftStructData.mileageGPSData.GetMileageLoaded(),
		shiftStructData.mileageGPSData.GetMileageAtBeginningOfLoading(),
		shiftStructData.mileageGPSData.GetMileageEmpty(),
	)
	if err == nil {
		log.Debugf("объект смены обновлен в БД id смены: %d", shiftStructData.mainData.GetShiftId())
	}
	return err
}

func (s *StorageMessageHandler) makeRequestUpdateSession(sessionStructData sessionDataFromModule) error {
	err := s.dbConn.ExecQuery(updateDriverSession,
		sessionStructData.mainData.GetSessionId(),
		sessionStructData.mainData.GetOffset(),
		sessionStructData.mainData.GetTimeUpdateSession(),
		sessionStructData.mainData.GetAvSpeed(),
		sessionStructData.engHoursData.GetEngHoursStart(),
		sessionStructData.engHoursData.GetEngHoursCurrent(),
		sessionStructData.engHoursData.GetEngHoursEnd(),
		sessionStructData.mileageData.GetMileageStart(),
		sessionStructData.mileageData.GetMileageCurrent(),
		sessionStructData.mileageData.GetMileageEnd(),
		sessionStructData.mileageData.GetMileageLoaded(),
		sessionStructData.mileageData.GetMileageAtBeginningOfLoading(),
		sessionStructData.mileageData.GetMileageEmpty(),
		sessionStructData.mileageGPSData.GetMileageStart(),
		sessionStructData.mileageGPSData.GetMileageCurrent(),
		sessionStructData.mileageGPSData.GetMileageEnd(),
		sessionStructData.mileageGPSData.GetMileageLoaded(),
		sessionStructData.mileageGPSData.GetMileageAtBeginningOfLoading(),
		sessionStructData.mileageGPSData.GetMileageEmpty(),
	)
	if err == nil {
		log.Debugf("объект сессии обновлен в БД id сессии: %d", sessionStructData.mainData.GetSessionId())
	}
	return err
}
