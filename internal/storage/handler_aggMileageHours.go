package storage

import (
	"errors"
	"sync"

	utils "agg-data-per-shift/pkg/utils"

	log "github.com/sirupsen/logrus"
)

var (
	errConvertTypeError    = errors.New("error converting the interface to a structure")
	errConvertShiftError   = errors.New("shift conversion error")
	errConvertSessionError = errors.New("session conversion error")
)

type aggMileageAndHoursHandler struct {
	dbConn *PgConn
}

// метод обрабатывает сообщение от модуля aggMileageHours
//   - ctx общий контекст storage (прекращает работу модуля)
func (a *aggMileageAndHoursHandler) handlerMesAggMileageHours(message trunsportMes) {
	var response responceIn
	mes, err := utils.TypeConversion[mesFromAggMileageHours](message.GetMesage())
	if err != nil {
		err = utils.Wrapper(handlerMesAggMileageHoursError{}, err)
		log.Error(err)
		r := responceForAggMileageHours{}
		r.criticalErr = err
		response = r
	} else {
		response = a.processingMessage(mes)
	}

	message.GetChForResponse() <- response
}

func (a *aggMileageAndHoursHandler) processingMessage(mes mesFromAggMileageHours) responceIn {
	var response responceIn
	switch mes.GetType() {
	case restoreShiftDataPerObj:
		res := a.handlerRestoreShiftDataPerObj(mes.GetObjID())
		res.responseShift.criticalErr, res.responseShift.err = handlingErrors(res.responseShift.err)
		res.responseSession.criticalErr, res.responseSession.err = handlingErrors(res.responseSession.err)
		log.Infof("Ответ по восстановлению состояния отправлен, ObjID: %d", mes.GetObjID())
		response = res
	default:
		res := a.processingRequestsToAddOrUpdate(mes)
		res.responseShift.criticalErr, res.responseShift.err = handlingErrors(res.responseShift.err)
		res.responseSession.criticalErr, res.responseSession.err = handlingErrors(res.responseSession.err)
		response = res
	}

	return response
}

// метод производит два ассинхронных запроса на получение строк из БД
func (a *aggMileageAndHoursHandler) handlerRestoreShiftDataPerObj(objId int) responceForAggMileageHours {
	defer log.Infof("Закончена обработка запроса на восстановления данных для объекта: %d", objId)
	var responce responceForAggMileageHours
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		res, err := a.dbConn.QueryDB(getLastObjShift, objId)
		if err != nil {
			responce.responseShift.err = err
			return
		}
		responce.responseShift.data, responce.responseShift.err = converQuery[RowShiftObjData](res)
	}()

	go func() {
		defer wg.Done()
		res, err := a.dbConn.QueryDB(getLastObjSession, objId)
		if err != nil {
			responce.responseShift.err = err
			return
		}
		responce.responseSession.data, responce.responseSession.err = converQuery[RowSessionObjData](res)
	}()

	wg.Wait()
	return responce
}

// метод обрабатывает запросы на модуля агрегации на обновление или добавление записей в таблицы
func (a *aggMileageAndHoursHandler) processingRequestsToAddOrUpdate(mes mesFromAggMileageHours) responceAggMileageHoursAddNewShiftAndSession {
	var response responceAggMileageHoursAddNewShiftAndSession

	shift, session, err := a.convertDataShiftAndSession(mes.GetShiftData(), mes.GetSessionData())
	if err != nil {
		response.criticalErr = err
		return response
	}
	switch mes.GetType() {
	case addNewShiftAndSession:
		log.Debugf("Добавление новых записей смены и сессии для объекта: %d", mes.GetObjID())
		response.responseShift, response.responseSession = a.handlerAddNewShiftAndSession(mes.GetObjID(), shift, session)
	case updateShiftAndAddNewSession:
		log.Debugf("Добавление новой записи сессии, обновление записи смены для объекта : %d", mes.GetObjID())
		response.responseShift, response.responseSession = a.handlerUpdateShiftAndAddNewSession(mes.GetObjID(), shift, session)
	case updateShiftAndSession:
		log.Debugf("Обновление записи сессии, обновление записи смены для объекта : %d", mes.GetObjID())
		response.responseShift, response.responseSession = a.handlerUpdateShiftAndSession(shift, session)
	}

	return response
}

// статический метод для конвертации данных из интерфейсов в струкруты shiftDataFromModule и sessionDataFromModule
func (a *aggMileageAndHoursHandler) convertDataShiftAndSession(shiftData, sessionData interface{}) (shiftDataFromModule, sessionDataFromModule, error) {
	shiftStructData := shiftDataFromModule{}
	sessionStructData := sessionDataFromModule{}

	err := shiftStructData.loadData(shiftData)
	if err != nil {
		err = utils.Wrapper(errConvertTypeError, utils.Wrapper(errConvertShiftError, err))
		return shiftStructData, sessionStructData, err
	}
	err = sessionStructData.loadData(sessionData)
	if err != nil {
		err = utils.Wrapper(errConvertTypeError, utils.Wrapper(errConvertSessionError, err))
		return shiftStructData, sessionStructData, err
	}

	return shiftStructData, sessionStructData, err
}

func (a *aggMileageAndHoursHandler) handlerAddNewShiftAndSession(objId int, shiftStructData shiftDataFromModule, sessionStructData sessionDataFromModule) (responceDataFromDB[int], responceDataFromDB[int]) {
	var respShift, respSession responceDataFromDB[int]

	respShift.err = a.makeRquestAddNewShift(shiftStructData, objId, &respShift.data)
	if respShift.err != nil {
		return respShift, respSession
	}
	respSession.err = a.makeRquestAddNewSession(sessionStructData, objId, respShift.data, &respSession.data)

	return respShift, respSession
}

func (a *aggMileageAndHoursHandler) handlerUpdateShiftAndAddNewSession(objId int, shiftStructData shiftDataFromModule, sessionStructData sessionDataFromModule) (responceDataFromDB[int], responceDataFromDB[int]) {
	var respShift, respSession responceDataFromDB[int]
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		respSession.err = a.makeRquestAddNewSession(sessionStructData, objId, shiftStructData.mainData.GetShiftId(), &respSession.data)
		wg.Done()
	}()

	go func() {
		respShift.err = a.makeRequestUpdateShift(shiftStructData)
		wg.Done()
	}()

	wg.Wait()
	return respShift, respSession
}

func (a *aggMileageAndHoursHandler) handlerUpdateShiftAndSession(shiftStructData shiftDataFromModule, sessionStructData sessionDataFromModule) (responceDataFromDB[int], responceDataFromDB[int]) {
	var respShift, respSession responceDataFromDB[int]
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		respShift.err = a.makeRequestUpdateShift(shiftStructData)
		wg.Done()
	}()

	go func() {
		respSession.err = a.makeRequestUpdateSession(sessionStructData)
		wg.Done()
	}()

	wg.Wait()
	return respShift, respSession
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
