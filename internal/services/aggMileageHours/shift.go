package aggmileagehours

import (
	"fmt"
	"time"
)

// данные смены по объекту техники
type shiftObjData struct {
	id              int          // id текущей смены
	numShift        int          // номер смены (1/2....n)
	shiftDateStart  time.Time    // время и дата начала смены (время первого события смены)
	shiftDateEnd    time.Time    // время окончания смены (время последнего обновления)
	shiftDate       time.Time    // текущая дата смены (смена может начинаться в другую дату)
	updatedTime     time.Time    // время обновления смены
	offset          int64        // offset события, которое последнее обновило состояние смены
	currentDriverId int          // id текущего водителя техники (может меняться в пределах смены)
	loaded          bool         // находится ли техника в груженом состоянии
	engHoursData    *engHours    // данные по моточасам за смену
	mileageData     *mileageData // данные по пробегу за смену
	mileageGPSData  *mileageData // данные по пробегу по GPS за смену
}

// метод для загрузки данных в структуру из интерфеса
func (s *shiftObjData) loadingInterfaceData(interfaceData interface{}) error {
	dataShift, err := typeConversion[dataShiftFromStorage](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"shiftObjData"}, err)
		return err
	}

	s.id = dataShift.GetShiftId()
	s.numShift = dataShift.GetShiftNum()
	s.shiftDate = dataShift.GetShiftDate()
	s.updatedTime = dataShift.GetUpdatedTime()
	s.offset = dataShift.GetOffset()
	s.loaded = dataShift.GetStatusLoaded()

	err = s.engHoursData.loadingInterfaceData(dataShift.GetEngHoursData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"shiftObjData"}, err)
		return err
	}
	err = s.mileageData.loadingInterfaceData(dataShift.GetMileageData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"shiftObjData"}, err)
		return err
	}
	err = s.mileageGPSData.loadingInterfaceData(dataShift.GetMileageGPSData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"shiftObjData"}, err)
		return err
	}

	return err
}

// метод сравнивает номер и дату смены с номером и датой переданных в параметрах
func (s *shiftObjData) checkDateNumCurrentShift(numShift int, dateShift time.Time) bool {
	return s.numShift == numShift && comparingDates(s.shiftDate, dateShift)
}

// метод создает новый объект смены на основании данных в старой смене
func (s *shiftObjData) createNewShift(numShift int, dateShift, mesTime time.Time) *shiftObjData {
	newShift := &shiftObjData{
		// id смены не заполняется, т.к. его нужно получить из БД
		// updatedTime заполняется во время обновления данных
		shiftDateStart: mesTime,
		numShift:       numShift,
		shiftDate:      dateShift,
		loaded:         s.loaded, // флаг загрузки переносится с предыдущей смены, т.к. техника может быть еще не разгружена
		engHoursData:   s.engHoursData.createNewEngHours(),
		mileageData:    s.mileageData.createNewMileageData(),
		mileageGPSData: s.mileageGPSData.createNewMileageData(),
	}

	return newShift
}

func (s *shiftObjData) updateShiftObjData(eventData *eventData, eventOffset int64, objLoaded bool) {
	// id не меняется (возвращается из БД)
	// numShift не меняется (задается при создании смены)
	// shiftDateStart не меняется (задается при создании смены)
	s.shiftDateEnd = eventData.mesTime
	// shiftDate не меняется (задается при создании смены)
	s.updatedTime = eventData.mesTime
	s.offset = eventOffset
	s.currentDriverId = eventData.numDriver
	s.loaded = objLoaded
	s.engHoursData.updateEngHours(eventData.engineHours)
	s.mileageData.updateMileageData(eventData.mileage, objLoaded)
	s.mileageGPSData.updateMileageData(eventData.gpsMileage, objLoaded)
}
