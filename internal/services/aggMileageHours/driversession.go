package aggmileagehours

import (
	"time"
)

// функция создает объект сессии на основании данных из события.
func initNewSession(event *eventData, eventOffset int64) *sessionDriverData {
	newSession := &sessionDriverData{
		// shiftId: будет заполнено после записи объекта в БД
		// sessionId: будет записано после записи объекта в БД
		driverId:          event.numDriver,
		offset:            eventOffset,
		timeStartSession:  event.mesTime,
		timeUpdateSession: event.mesTime,
		avSpeed:           event.avSpeed,
	}

	newSession.initAggDataFields(event)

	return newSession
}

// функция создает новую сессию на основании данных из БД
func initNewSessionLoadingDBData(data interface{}) (*sessionDriverData, error) {
	newSession := &sessionDriverData{}
	err := newSession.loadingData(data)
	return newSession, err
}

// данные по сессии водителя в смене
type sessionDriverData struct {
	shiftId           int       // id смены, в которой находится сессия
	sessionId         int       // id сессии, берется из БД
	driverId          int       // id водителя
	offset            int64     // последний записанный offset
	timeStartSession  time.Time // время старта сессии
	timeUpdateSession time.Time // время обновления записи сессии
	avSpeed           float32   // средняя скорость водителя
	aggData
}

func (s sessionDriverData) GetShiftId() int {
	return s.shiftId
}
func (s sessionDriverData) GetSessionId() int {
	return s.sessionId
}
func (s sessionDriverData) GetDriverId() int {
	return s.driverId
}
func (s sessionDriverData) GetOffset() int64 {
	return s.offset
}
func (s sessionDriverData) GetTimeStartSession() time.Time {
	return s.timeStartSession
}
func (s sessionDriverData) GetTimeUpdateSession() time.Time {
	return s.timeUpdateSession
}
func (s sessionDriverData) GetAvSpeed() float32 {
	return s.avSpeed
}

func (s *sessionDriverData) setSessionId(id int) {
	s.sessionId = id
}

func (s *sessionDriverData) setShiftId(id int) {
	s.shiftId = id
}

// метод для загрузки данных в структуру из интерфеса
func (s *sessionDriverData) loadingData(data interface{}) error {
	sData, err := typeСonversion[dataDriverSessionFromStorage](data)
	if err != nil {
		return err
	}

	s.shiftId = sData.GetShiftId()
	s.sessionId = sData.GetSessionId()
	s.driverId = sData.GetDriverId()
	s.offset = sData.GetOffset()
	s.timeStartSession = sData.GetTimeStartSession()
	s.timeUpdateSession = sData.GetTimeUpdateSession()
	s.avSpeed = sData.GetAvSpeed()

	err = s.loadingDataFromInterface(
		sData.GetMileageData(),
		sData.GetMileageGPSData(),
		sData.GetEngHoursData(),
	)

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
	}

	newDriverSession.initNewAggDataFields(s.EngHoursData, s.MileageData, s.MileageGPSData)

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
	s.updateDataFields(eventData, objLoaded)
}
