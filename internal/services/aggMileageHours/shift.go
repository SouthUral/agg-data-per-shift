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
		engHoursData:    initNewEngHours(event),
		mileageData:     initNewMileageData(event),
		mileageGPSData:  initNewMileageData(event),
	}

	return newShift
}

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

// метод преобразует объект смены в json
func (s *ShiftObjData) conversionToJson() ([]byte, error) {
	trunsportStruct := RowShiftObjData{
		Id:              s.Id,
		NumShift:        s.NumShift,
		ShiftDateStart:  s.ShiftDateStart,
		ShiftDateEnd:    s.ShiftDateEnd,
		ShiftDate:       s.ShiftDate,
		UpdatedTime:     s.UpdatedTime,
		Offset:          s.Offset,
		CurrentDriverId: s.CurrentDriverId,
		Loaded:          s.Loaded,
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

	res, err := conversionToJson[RowShiftObjData](trunsportStruct)
	return res, err
}
