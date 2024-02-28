package aggmileagehours

import (
	"time"
)

// события для отправки в горутину агрегации
type eventForAgg struct {
	offset    int64
	eventData *eventData
}

func initShiftObjTransportData(dataShift shiftObjData) shiftObjTransportData {
	res := shiftObjTransportData{}
	res.convertData(dataShift)
	return res
}

// структура для отправки в модуль storage
type shiftObjTransportData struct {
	id              int         // id текущей смены
	numShift        int         // номер смены (1/2....n)
	shiftDateStart  time.Time   // время и дата начала смены (время первого события смены)
	shiftDateEnd    time.Time   // время окончания смены (время последнего обновления)
	shiftDate       time.Time   // текущая дата смены (смена может начинаться в другую дату)
	updatedTime     time.Time   // время обновления смены
	offset          int64       // offset события, которое последнее обновило состояние смены
	currentDriverId int         // id текущего водителя техники (может меняться в пределах смены)
	loaded          bool        // находится ли техника в груженом состоянии
	engHoursData    engHours    // данные по моточасам за смену
	mileageData     mileageData // данные по пробегу за смену
	mileageGPSData  mileageData // данные по пробегу по GPS за смену
}

func (s *shiftObjTransportData) convertData(dataShift shiftObjData) {
	s.id = dataShift.id
	s.numShift = dataShift.numShift
	s.shiftDateStart = dataShift.shiftDateStart
	s.shiftDateEnd = dataShift.shiftDateEnd
	s.shiftDate = dataShift.shiftDate
	s.updatedTime = dataShift.updatedTime
	s.offset = dataShift.offset
	s.currentDriverId = dataShift.currentDriverId
	s.loaded = dataShift.loaded
	s.engHoursData = *dataShift.engHoursData
	s.mileageData = *dataShift.mileageData
	s.mileageGPSData = *dataShift.mileageGPSData
}

func (s shiftObjTransportData) GetId() int {
	return s.id
}

func (s shiftObjTransportData) GetNumShift() int {
	return s.numShift
}

func (s shiftObjTransportData) GetShiftDateStart() time.Time {
	return s.shiftDateStart
}

func (s shiftObjTransportData) GetShiftDateEnd() time.Time {
	return s.shiftDateEnd
}

func (s shiftObjTransportData) GetShiftDate() time.Time {
	return s.shiftDate
}

func (s shiftObjTransportData) GetUpdatedTime() time.Time {
	return s.updatedTime
}

func (s shiftObjTransportData) GetOffset() int64 {
	return s.offset
}

func (s shiftObjTransportData) GetCurrentDriverId() int {
	return s.currentDriverId
}

func (s shiftObjTransportData) GetLoaded() bool {
	return s.loaded
}

func (s shiftObjTransportData) GetEngHoursData() interface{} {
	return s.engHoursData
}

func (s shiftObjTransportData) GetMileageData() interface{} {
	return s.mileageData
}

func (s shiftObjTransportData) GetMileageGPSData() interface{} {
	return s.mileageGPSData
}

func initSessionDriverTransportData(dataSession sessionDriverData) sessionDriverTransportData {
	res := sessionDriverTransportData{}
	res.convertData(dataSession)
	return res
}

// структура с данными по сессии водителей для отправки в storage
type sessionDriverTransportData struct {
	sessionId         int         // id сессии, берется из БД
	driverId          int         // id водителя
	offset            int64       // последний записанный offset
	timeStartSession  time.Time   // время старта сессии
	timeUpdateSession time.Time   // время обновления записи сессии
	avSpeed           float64     // средняя скорость водителя
	engHoursData      engHours    // данные по моточасам за сессию
	mileageData       mileageData // данные по пробегу за смену
	mileageGPSData    mileageData // данные пробега по GPS за сессию
}

func (s *sessionDriverTransportData) convertData(dataSession sessionDriverData) {
	s.sessionId = dataSession.sessionId
	s.driverId = dataSession.driverId
	s.offset = dataSession.offset
	s.timeStartSession = dataSession.timeStartSession
	s.timeUpdateSession = dataSession.timeUpdateSession
	s.avSpeed = dataSession.avSpeed
	s.engHoursData = *dataSession.engHoursData
	s.mileageData = *dataSession.mileageData
	s.mileageGPSData = *dataSession.mileageGPSData
}

func (s sessionDriverTransportData) GetSessionId() int {
	return s.sessionId
}

func (s sessionDriverTransportData) GetDriverId() int {
	return s.driverId
}

func (s sessionDriverTransportData) GetOffset() int64 {
	return s.offset
}

func (s sessionDriverTransportData) GetTimeStartSession() time.Time {
	return s.timeStartSession
}

func (s sessionDriverTransportData) GetTimeUpdateSession() time.Time {
	return s.timeUpdateSession
}

func (s sessionDriverTransportData) GetAvSpeed() float64 {
	return s.avSpeed
}

func (s sessionDriverTransportData) GetEngHoursData() interface{} {
	return s.engHoursData
}

func (s sessionDriverTransportData) GetMileageData() interface{} {
	return s.mileageData
}

func (s sessionDriverTransportData) GetMileageGPSData() interface{} {
	return s.mileageGPSData
}

type mesForStorage struct {
	typeMes         string
	objectID        int
	shiftInitData   shiftObjTransportData
	sessionInitData sessionDriverTransportData
}

// метод для получения типа сообщения
func (m mesForStorage) GetType() string {
	return m.typeMes
}

// метод возвращает objID техники, для восстановления состояния
func (m mesForStorage) GetObjID() int {
	return m.objectID
}

// метод возвращает начальные данные для смены (данные для добавления новых данных в БД)
func (m mesForStorage) GetShiftData() interface{} {
	return m.shiftInitData

}

// метод возвращает начальные данные для сессии (данные для добавления новых данных в БД)
func (m mesForStorage) GetSessionData() interface{} {
	return m.sessionInitData
}

// транспортная структура (универсальный интерфейс)
type transportStruct struct {
	sender         string           // модуль отправитель сообщения
	mesage         mesForStorage    // сообщение отправителя
	reverseChannel chan interface{} // канал для отправки ответа от модуля storage
}

func (t transportStruct) GetSender() string {
	return t.sender
}

func (t transportStruct) GetMesage() interface{} {
	return t.mesage
}

// метод для отправки ответа от модуля storage
func (t transportStruct) SendAnswer(answer interface{}) {
	t.reverseChannel <- answer
}
