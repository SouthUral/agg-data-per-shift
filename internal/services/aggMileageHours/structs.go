package aggmileagehours

import (
	"time"
)

// настройки смены
type settingsDurationShifts struct {
	shifts          map[int]settingShift
	offsetTimeShift int // времянное смещение смены
}

type settingShift struct {
	numShift       int       // номер смены
	startTimeShift time.Time // время старта смены
	shiftDuration  int       // продолжительность смены
}

// данные смены по объекту техники
type shiftObjData struct {
	id                  int         // id текущей смены
	numShift            int         // номер смены (1/2....n)
	currentShiftDate    time.Time   // текущая дата смены
	currentDriverId     int         // id текущего водителя техники (может меняться в пределах смены)
	loaded              bool        // находится ли техника в груженом состоянии
	motoHoursStartShift float64     // моточасы на начало смены
	motoHoursCurrent    float64     // текущие моточасы
	mileageData         mileageData // данные по пробегу за смену
	mileageGPSData      mileageData // данные по пробегу по GPS за смену
}

// данные по пробегу за смену
type mileageData struct {
	mileageStartShift int // пробег на начало смены
	mileageCurrent    int // текущий пробег
	mileageLoaded     int // пробег в груженом состоянии
	mileageEmpty      int // пробег в порожнем состоянии
}

// данные которые получили из события
type rawEventData struct {
	EventInfo struct {
		Const string `json:"const"`
	} `json:"event_info"` // event_info -> const : тип события
	Data struct {
		Mileage     int     `json:"mileage"`      // data -> mileage : пробег
		GpsMileage  int     `json:"gps_mileage"`  // data -> gps_mileage : пробег по gps
		EngineHours float64 `json:"engine_hours"` // data -> engine_hours : моточасы
	} `json:"data"`
	EventData struct {
		DriverInfo struct {
			Fio    string `json:"fio"`
			TabNum int    `json:"tab_num,string"`
		} `json:"driver_info"`
	} `json:"event_data"`
	ObjectID int    `json:"object_id"` // object_id : id техники
	MesTime  string `json:"mes_time"`  // mes_time : время сообщения
}

func (e rawEventData) getDecryptedData() (*eventData, error) {
	messTime, err := timeConversion(e.MesTime)
	data := &eventData{
		typeEvent:   e.EventInfo.Const,
		objectID:    e.ObjectID,
		mesTime:     messTime,
		mileage:     e.Data.Mileage,
		gpsMileage:  e.Data.GpsMileage,
		engineHours: e.Data.EngineHours,
		fioDriver:   e.EventData.DriverInfo.Fio,
		numDriver:   e.EventData.DriverInfo.TabNum,
	}

	return data, err
}

type eventData struct {
	typeEvent   string    // тип события
	objectID    int       // id техники
	mesTime     time.Time // время сообщения
	mileage     int       // пробег
	gpsMileage  int       // пробег по gps
	engineHours float64   // моточасы
	fioDriver   string    // ФИО водителя
	numDriver   int       // номер водителя
}
