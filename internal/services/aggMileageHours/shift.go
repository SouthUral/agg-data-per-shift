package aggmileagehours

import (
	"time"
)

// функция создает объект смены на основании данных из события.
func initNewShift(event *eventData, numShift int, dateShift time.Time, eventOffset int64) *ShiftObjData {
	newShift := &ShiftObjData{
		// Id: будет заполнено после записи данных в БД
		NumShift:        numShift,
		ShiftDateStart:  event.mesTime, // нужно сюда поставить дату начала смены, а не текущего события
		ShiftDateEnd:    event.mesTime, // нужно поставить дату конца смены, а не текущего события
		ShiftDate:       dateShift,
		UpdatedTime:     event.mesTime,
		Offset:          eventOffset,
		CurrentDriverId: event.numDriver,
		Loaded:          false, // при создании неизвестно является ли транспорт груженым
	}

	newShift.initAggDataFields(event)

	return newShift
}

// функция создает структуру ShiftObjData на основании данных из БД
func initNewShiftLoadingDBData(data interface{}) (*ShiftObjData, error) {
	newShift := &ShiftObjData{}
	err := newShift.loadingData(data)
	return newShift, err
}

// данные смены по объекту техники
type ShiftObjData struct {
	Id              int       `json:"id"`                // id текущей смены
	NumShift        int       `json:"num_shift"`         // номер смены (1/2....n)
	ShiftDateStart  time.Time `json:"shift_date_start"`  // время и дата начала смены (время первого события смены)
	ShiftDateEnd    time.Time `json:"shift_date_end"`    // время окончания смены (время последнего обновления)
	ShiftDate       time.Time `json:"shift_date"`        // текущая дата смены (смена может начинаться в другую дату)
	UpdatedTime     time.Time `json:"updated_time"`      // время обновления смены
	Offset          int64     `json:"event_offset"`      // offset события, которое последнее обновило состояние смены
	CurrentDriverId int       `json:"current_driver_id"` // id текущего водителя техники (может меняться в пределах смены)
	Loaded          bool      `json:"loaded"`            // находится ли техника в груженом состоянии
	aggData
}

func (s ShiftObjData) GetShiftId() int {
	return s.Id
}
func (s ShiftObjData) GetShiftNum() int {
	return s.NumShift
}
func (s ShiftObjData) GetShiftDateStart() time.Time {
	return s.ShiftDateStart
}
func (s ShiftObjData) GetShiftDateEnd() time.Time {
	return s.ShiftDateEnd
}
func (s ShiftObjData) GetShiftDate() time.Time {
	return s.ShiftDate
}
func (s ShiftObjData) GetUpdatedTime() time.Time {
	return s.UpdatedTime
}
func (s ShiftObjData) GetOffset() int64 {
	return s.Offset
}
func (s ShiftObjData) GetCurrentDriverId() int {
	return s.CurrentDriverId
}
func (s ShiftObjData) GetStatusLoaded() bool {
	return s.Loaded
}

func (s *ShiftObjData) setShiftId(id int) {
	s.Id = id
}

// метод для загрузки данных в структуру из интерфеса
func (s *ShiftObjData) loadingData(data interface{}) error {
	shiftData, err := typeСonversion[dataShiftFromStorage](data)
	if err != nil {
		return err
	}

	s.Id = shiftData.GetShiftId()
	s.NumShift = shiftData.GetShiftNum()
	s.ShiftDateStart = shiftData.GetShiftDateStart()
	s.ShiftDateEnd = shiftData.GetShiftDateEnd()
	s.ShiftDate = shiftData.GetShiftDate()
	s.UpdatedTime = shiftData.GetUpdatedTime()
	s.Offset = shiftData.GetOffset()
	s.CurrentDriverId = shiftData.GetCurrentDriverId()
	s.Loaded = shiftData.GetStatusLoaded()

	err = s.loadingDataFromInterface(
		shiftData.GetMileageData(),
		shiftData.GetMileageGPSData(),
		shiftData.GetEngHoursData(),
	)

	return err
}

// метод сравнивает номер и дату смены с номером и датой переданных в параметрах
func (s *ShiftObjData) checkDateNumCurrentShift(numShift int, dateShift time.Time) bool {
	return s.NumShift == numShift && comparingDates(s.ShiftDate, dateShift)
}

// метод создает новый объект смены на основании данных в старой смене
func (s *ShiftObjData) createNewShift(numShift int, dateShift, mesTime time.Time) *ShiftObjData {
	newShift := &ShiftObjData{
		// id смены не заполняется, т.к. его нужно получить из БД
		// updatedTime заполняется во время обновления данных
		ShiftDateStart: mesTime,
		NumShift:       numShift,
		ShiftDate:      dateShift,
		Loaded:         s.Loaded, // флаг загрузки переносится с предыдущей смены, т.к. техника может быть еще не разгружена
	}

	newShift.initNewAggDataFields(s.EngHoursData, s.MileageData, s.MileageGPSData)

	return newShift
}

func (s *ShiftObjData) updateShiftObjData(eventData *eventData, eventOffset int64, objLoaded bool) {
	// id не меняется (возвращается из БД)
	// numShift не меняется (задается при создании смены)
	// shiftDateStart не меняется (задается при создании смены)
	s.ShiftDateEnd = eventData.mesTime
	// shiftDate не меняется (задается при создании смены)
	s.UpdatedTime = eventData.mesTime
	s.Offset = eventOffset
	s.CurrentDriverId = eventData.numDriver
	s.Loaded = objLoaded
	s.updateDataFields(eventData, objLoaded)
}
