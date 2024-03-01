package aggmileagehours

import (
	"time"
)

// данные смены по объекту техники
type ShiftObjData struct {
	Id              int          `json:"id"`                // id текущей смены
	NumShift        int          `json:"num_shift"`         // номер смены (1/2....n)
	ShiftDateStart  time.Time    `json:"shift_date_start"`  // время и дата начала смены (время первого события смены)
	ShiftDateEnd    time.Time    `json:"shift_date_end"`    // время окончания смены (время последнего обновления)
	ShiftDate       time.Time    `json:"shift_date"`        // текущая дата смены (смена может начинаться в другую дату)
	UpdatedTime     time.Time    `json:"updated_time"`      // время обновления смены
	Offset          int64        `json:"event_offset"`      // offset события, которое последнее обновило состояние смены
	CurrentDriverId int          `json:"current_driver_id"` // id текущего водителя техники (может меняться в пределах смены)
	Loaded          bool         `json:"loaded"`            // находится ли техника в груженом состоянии
	engHoursData    *engHours    // данные по моточасам за смену
	mileageData     *mileageData // данные по пробегу за смену
	mileageGPSData  *mileageData // данные по пробегу по GPS за смену
}

func (s *ShiftObjData) setShiftId(id int) {
	s.Id = id
}

// метод для загрузки данных в структуру из интерфеса
func (s *ShiftObjData) loadingData(data []byte) error {
	unmData, err := decodingMesFromStorageToStruct[RowShiftObjData](data)
	if err != nil {
		return err
	}

	s.Id = unmData.Id
	s.NumShift = unmData.NumShift
	s.ShiftDateStart = unmData.ShiftDateStart
	s.ShiftDateEnd = unmData.ShiftDateEnd
	s.ShiftDate = unmData.ShiftDate
	s.UpdatedTime = unmData.UpdatedTime
	s.Offset = unmData.Offset
	s.CurrentDriverId = unmData.CurrentDriverId
	s.Loaded = unmData.Loaded
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
		engHoursData:   s.engHoursData.createNewEngHours(),
		mileageData:    s.mileageData.createNewMileageData(),
		mileageGPSData: s.mileageGPSData.createNewMileageData(),
	}

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
	s.engHoursData.updateEngHours(eventData.engineHours)
	s.mileageData.updateMileageData(eventData.mileage, objLoaded)
	s.mileageGPSData.updateMileageData(eventData.gpsMileage, objLoaded)
}
