package aggmileagehours

import (
	"time"
)

// функция создает объект сессии на основании данных из события.
func initNewSession(event *eventData, eventOffset int64) *sessionDriverData {
	newShift := &sessionDriverData{
		// shiftId: будет заполнено после записи объекта в БД
		// sessionId: будет записано после записи объекта в БД
		driverId:          event.numDriver,
		offset:            eventOffset,
		timeStartSession:  event.mesTime,
		timeUpdateSession: event.mesTime,
		avSpeed:           event.avSpeed,
		engHoursData:      initNewEngHours(event),
		mileageData:       initNewMileageData(event),
		mileageGPSData:    initNewMileageData(event),
	}

	return newShift
}

// данные по сессии водителя в смене
type sessionDriverData struct {
	shiftId           int          // id смены, в которой находится сессия
	sessionId         int          // id сессии, берется из БД
	driverId          int          // id водителя
	offset            int64        // последний записанный offset
	timeStartSession  time.Time    // время старта сессии
	timeUpdateSession time.Time    // время обновления записи сессии
	avSpeed           float32      // средняя скорость водителя
	engHoursData      *engHours    // данные по моточасам за сессию
	mileageData       *mileageData // данные по пробегу за смену
	mileageGPSData    *mileageData // данные пробега по GPS за сессию
}

func (s *sessionDriverData) setSessionId(id int) {
	s.sessionId = id
}

func (s *sessionDriverData) setShiftId(id int) {
	s.shiftId = id
}

// метод для загрузки данных в структуру из интерфеса
func (s *sessionDriverData) loadingData(data []byte) error {
	unmData, err := decodingMesFromStorageToStruct[RowSessionObjData](data)
	if err != nil {
		return err
	}

	s.shiftId = unmData.ShiftId
	s.sessionId = unmData.SessionId
	s.driverId = unmData.DriverId
	s.offset = unmData.Offset
	s.timeStartSession = unmData.TimeStartSession
	s.timeUpdateSession = unmData.TimeUpdateSession
	s.avSpeed = unmData.AvSpeed

	s.engHoursData = &engHours{
		engHoursStart:   unmData.EngHoursStart,
		engHoursCurrent: unmData.EngHoursCurrent,
		engHoursEnd:     unmData.EngHoursEnd,
	}
	s.mileageData = &mileageData{
		mileageStart:                unmData.MileageStart,
		mileageEnd:                  unmData.MileageEnd,
		mileageCurrent:              unmData.MileageCurrent,
		mileageLoaded:               unmData.MileageLoaded,
		mileageAtBeginningOfLoading: unmData.MileageAtBeginningOfLoading,
		mileageEmpty:                unmData.MileageEmpty,
	}
	s.mileageGPSData = &mileageData{
		mileageStart:                unmData.MileageGPSStart,
		mileageEnd:                  unmData.MileageGPSEnd,
		mileageCurrent:              unmData.MileageGPSCurrent,
		mileageLoaded:               unmData.MileageGPSLoaded,
		mileageAtBeginningOfLoading: unmData.MileageGPSAtBeginningOfLoading,
		mileageEmpty:                unmData.MileageGPSEmpty,
	}

	return err
}

// метод сравнивает входящий параметр id водителя с текущим id водителя
func (s *sessionDriverData) checkDriverSession(idDriver int) bool {
	return s.driverId == idDriver
}

func (s *sessionDriverData) createNewDriverSession(driverId int, mesTime time.Time) *sessionDriverData {
	newDriverSession := &sessionDriverData{
		// shiftId записывается уже после добавления новой записи в БД??
		// sessionId возвращается из БД после создания новой записи
		// offset записывается при обновлении записи
		driverId:         driverId,
		timeStartSession: mesTime,
		// timeUpdateSession записывается при обновлении записи
		// avSpeed записывается при обновлении записи
		engHoursData:   s.engHoursData.createNewEngHours(),
		mileageData:    s.mileageData.createNewMileageData(),
		mileageGPSData: s.mileageGPSData.createNewMileageData(),
	}

	return newDriverSession
}

// метод для обновления информации о сессии водителя. Параметры:
//   - eventData: данные события;
//   - eventOffset: offset события;
//   - objLoaded: параметр обозначающий, сейчас машина едет груженой или нет.
func (s *sessionDriverData) updateSession(eventData *eventData, eventOffset int64, objLoaded bool) {
	s.offset = eventOffset
	s.timeUpdateSession = eventData.mesTime
	s.avSpeed = eventData.avSpeed
	s.engHoursData.updateEngHours(eventData.engineHours)
	s.mileageData.updateMileageData(eventData.mileage, objLoaded)
	s.mileageGPSData.updateMileageData(eventData.gpsMileage, objLoaded)
}

// метод преобразует объект смены в json
func (s *sessionDriverData) conversionToJson() ([]byte, error) {
	trunsportStruct := RowSessionObjData{
		ShiftId:           s.shiftId,
		SessionId:         s.sessionId,
		DriverId:          s.driverId,
		Offset:            s.offset,
		TimeStartSession:  s.timeStartSession,
		TimeUpdateSession: s.timeUpdateSession,
		AvSpeed:           s.avSpeed,

		EngHoursStart:   s.engHoursData.engHoursStart,
		EngHoursEnd:     s.engHoursData.engHoursEnd,
		EngHoursCurrent: s.engHoursData.engHoursCurrent,

		MileageStart:                s.mileageData.mileageStart,
		MileageEnd:                  s.mileageData.mileageEnd,
		MileageCurrent:              s.mileageData.mileageCurrent,
		MileageLoaded:               s.mileageData.mileageLoaded,
		MileageEmpty:                s.mileageData.mileageEmpty,
		MileageAtBeginningOfLoading: s.mileageData.mileageAtBeginningOfLoading,

		MileageGPSStart:                s.mileageGPSData.mileageStart,
		MileageGPSEnd:                  s.mileageGPSData.mileageEnd,
		MileageGPSCurrent:              s.mileageGPSData.mileageCurrent,
		MileageGPSLoaded:               s.mileageGPSData.mileageLoaded,
		MileageGPSEmpty:                s.mileageGPSData.mileageEmpty,
		MileageGPSAtBeginningOfLoading: s.mileageGPSData.mileageAtBeginningOfLoading,
	}

	res, err := conversionToJson[RowSessionObjData](trunsportStruct)
	return res, err
}
