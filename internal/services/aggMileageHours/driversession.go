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
func (s sessionDriverData) GetEngHoursData() interface{} {
	return *s.engHoursData
}
func (s sessionDriverData) GetMileageData() interface{} {
	return *s.mileageData
}
func (s sessionDriverData) GetMileageGPSData() interface{} {
	return *s.mileageGPSData
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

	err = s.mileageData.loadingData(sData.GetMileageData())
	if err != nil {
		return err
	}
	err = s.mileageGPSData.loadingData(sData.GetMileageGPSData())
	if err != nil {
		return err
	}
	err = s.engHoursData.loadingData(sData.GetEngHoursData())

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
