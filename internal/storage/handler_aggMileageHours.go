package storage

import (
	"context"
	"fmt"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

type aggMileageAndHoursHandler struct {
	dbConn *PgConn
}

// метод обрабатывает сообщение от модуля aggMileageHours
func (a *aggMileageAndHoursHandler) handlerMesAggMileageHours(ctx context.Context, message trunsportMes) {
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
		response := a.handlerRestoreShiftDataPerObj(ctx, mes.GetObjID())
		// нужно обработать ошибки
		// возвращется две ошибки, нужно обработать каждую, каждая ошибка проверяется на тип
		typeCriticalErr, err := handlingErrors(response.responseSession.err, response.responseShift.err)
		switch typeCriticalErr {
		case criticalError:
			log.Error(err)
			// TODO: завершить работу
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
		shiftId, sessionId, err := a.handlerAddNewShiftAndSession(mes.GetObjID(), mes.GetShiftData(), mes.GetSessionData())
		// возможно придется добавить обработку ошибок
		if err != nil {
			log.Error(err)
			// TODO: завершить работу
			return
		}
		message.GetChForResponse() <- responceAggMileageHoursAddNewShiftAndSession{
			shiftId:   shiftId,
			sessionId: sessionId,
		}
	case updateShiftAndAddNewSession:
		log.Debugf("Добавление новой записи сессии, обновление записи смены для объекта : %d", mes.GetObjID())
		sessionId, err := a.handlerUpdateShiftAndAddNewSession(mes.GetObjID(), mes.GetShiftData(), mes.GetSessionData())
		if err != nil {
			log.Error(err)
			// TODO: завершить работу
			return
		}
		message.GetChForResponse() <- responceAggMileageHoursAddNewShiftAndSession{
			sessionId: sessionId,
		}
	case updateShiftAndSession:
		err := a.handlerUpdateShiftAndSession(mes.GetShiftData(), mes.GetSessionData())
		if err != nil {
			log.Error(err)
			// TODO: завершить работу
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
func (a *aggMileageAndHoursHandler) handlerRestoreShiftDataPerObj(ctx context.Context, objId int) responceShiftSession {
	defer log.Infof("Закончена обработка запроса на восстановления данных для объекта: %d", objId)
	var counterResponse int
	var result responceShiftSession
	numResponse := 2 // количество ответов, которые нужно получить

	chShift := makeRequestAndProcess[RowShiftObjData](a.dbConn, getLastObjShift, objId)
	chSession := makeRequestAndProcess[RowSessionObjData](a.dbConn, getLastObjSession, objId)

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

func (a *aggMileageAndHoursHandler) handlerAddNewShiftAndSession(objId int, shiftData, sessionData interface{}) (int, int, error) {
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
	err = a.makeRquestAddNewShift(shiftStructData, objId, &shifId)
	if err != nil {
		return shifId, sessionId, err
	}
	err = a.makeRquestAddNewSession(sessionStructData, objId, shifId, &sessionId)
	if err != nil {
		return shifId, sessionId, err
	}
	return shifId, sessionId, err
}

func (a *aggMileageAndHoursHandler) handlerUpdateShiftAndAddNewSession(objId int, shiftData, sessionData interface{}) (int, error) {
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
	err = a.makeRquestAddNewSession(sessionStructData, objId, shiftStructData.mainData.GetShiftId(), &sessionId)
	if err != nil {
		return sessionId, err
	}
	err = a.makeRequestUpdateShift(shiftStructData)
	if err != nil {
		return sessionId, err
	}
	return sessionId, err
}

func (a *aggMileageAndHoursHandler) handlerUpdateShiftAndSession(shiftData, sessionData interface{}) error {
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
	err = a.makeRequestUpdateShift(shiftStructData)
	if err != nil {
		return err
	}
	err = a.makeRequestUpdateSession(sessionStructData)
	return err
}

// метод делает запрос в БД на добавлении новой смены в таблицу
func (a *aggMileageAndHoursHandler) makeRquestAddNewShift(shiftStructData shiftDataFromModule, objId int, shifId *int) error {
	err := a.dbConn.QueryRowDB(addNewShift,
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
		shiftStructData.mileageData.GetMileageEmpty(),
		shiftStructData.mileageGPSData.GetMileageStart(),
		shiftStructData.mileageGPSData.GetMileageCurrent(),
		shiftStructData.mileageGPSData.GetMileageEnd(),
		shiftStructData.mileageGPSData.GetMileageLoaded(),
		shiftStructData.mileageGPSData.GetMileageEmpty(),
	).Scan(shifId)
	if err == nil {
		log.Debugf("объект смены записан в БД id смены: %d", *shifId)
	}
	return err
}

// метод делает запрос в БД на добавление новой сессии в таблицу
func (a *aggMileageAndHoursHandler) makeRquestAddNewSession(sessionStructData sessionDataFromModule, objId, shifId int, sessionId *int) error {
	err := a.dbConn.QueryRowDB(addNewSession,
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
		sessionStructData.mileageData.GetMileageEmpty(),
		sessionStructData.mileageGPSData.GetMileageStart(),
		sessionStructData.mileageGPSData.GetMileageCurrent(),
		sessionStructData.mileageGPSData.GetMileageEnd(),
		sessionStructData.mileageGPSData.GetMileageLoaded(),
		sessionStructData.mileageGPSData.GetMileageEmpty(),
	).Scan(sessionId)
	if err == nil {
		log.Debugf("объект cессии записан в БД id cессии: %d", *sessionId)
	}
	return err
}

// метод делает запрос в БД на обновление смены в таблице
func (a *aggMileageAndHoursHandler) makeRequestUpdateShift(shiftStructData shiftDataFromModule) error {
	err := a.dbConn.ExecQuery(updateShift,
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
		shiftStructData.mileageData.GetMileageEmpty(),
		shiftStructData.mileageGPSData.GetMileageStart(),
		shiftStructData.mileageGPSData.GetMileageCurrent(),
		shiftStructData.mileageGPSData.GetMileageEnd(),
		shiftStructData.mileageGPSData.GetMileageLoaded(),
		shiftStructData.mileageGPSData.GetMileageEmpty(),
	)
	if err == nil {
		log.Debugf("объект смены обновлен в БД id смены: %d", shiftStructData.mainData.GetShiftId())
	}
	return err
}

func (a *aggMileageAndHoursHandler) makeRequestUpdateSession(sessionStructData sessionDataFromModule) error {
	err := a.dbConn.ExecQuery(updateDriverSession,
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
		sessionStructData.mileageData.GetMileageEmpty(),
		sessionStructData.mileageGPSData.GetMileageStart(),
		sessionStructData.mileageGPSData.GetMileageCurrent(),
		sessionStructData.mileageGPSData.GetMileageEnd(),
		sessionStructData.mileageGPSData.GetMileageLoaded(),
		sessionStructData.mileageGPSData.GetMileageEmpty(),
	)
	if err == nil {
		log.Debugf("объект сессии обновлен в БД id сессии: %d", sessionStructData.mainData.GetSessionId())
	}
	return err
}

// функция отправляет запрос в БД, получает ответ, конвертирует его в переданный тип, и из типа конвертирует его в json.
// Функция обрабатывает ответ в одну строку.
func makeRequestAndProcess[T RowSessionObjData | RowShiftObjData](dbConn *PgConn, request string, objectId int) chan responceDataFromDB[T] {
	responseCh := make(chan responceDataFromDB[T])
	go func() {
		response, err := dbConn.QueryDB(request, objectId)
		if err != nil {
			log.Error(err)
			responseCh <- responceDataFromDB[T]{err: err}
			return
		}

		sessionData, err := converQuery[T](response)
		if err != nil {
			log.Error(err)
			responseCh <- responceDataFromDB[T]{err: err}
			return
		}

		responseCh <- responceDataFromDB[T]{data: sessionData, err: err}
		defer response.Close()
	}()
	return responseCh
}
