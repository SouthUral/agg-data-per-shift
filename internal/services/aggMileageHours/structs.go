package aggmileagehours

import (
	"fmt"
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
	id               int         // id текущей смены
	numShift         int         // номер смены (1/2....n)
	currentShiftDate time.Time   // текущая дата смены
	updatedTime      time.Time   // время обновления смены
	offset           int         // offset события, которое последнее обновило состояние смены
	currentDriverId  int         // id текущего водителя техники (может меняться в пределах смены)
	loaded           bool        // находится ли техника в груженом состоянии
	engHoursData     engHours    // данные по моточасам за смену
	mileageData      mileageData // данные по пробегу за смену
	mileageGPSData   mileageData // данные по пробегу по GPS за смену
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
	s.currentShiftDate = dataShift.GetShiftDate()
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

// данные по пробегу за смену/сессию
type mileageData struct {
	mileageStart   int // пробег на начало (смены/сессии)
	mileageCurrent int // текущий пробег
	mileageEnd     int // проебег на конец смены
	mileageLoaded  int // пробег в груженом состоянии
	mileageEmpty   int // пробег в порожнем состоянии
}

// метод переброса данных из интерфейса в структуру
func (m *mileageData) loadingInterfaceData(interfaceData interface{}) error {
	mileageInterface, err := typeConversion[mileageDataInterface](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"mileageData"}, err)
		return err
	}

	m.mileageStart = mileageInterface.GetMileageStart()
	m.mileageEnd = mileageInterface.GetMileageEnd()
	m.mileageCurrent = mileageInterface.GetMileageCurrent()
	m.mileageLoaded = mileageInterface.GetMileageLoaded()
	m.mileageEmpty = mileageInterface.GetMileageEmpty()

	return err
}

// данные по моточасам за смену/сессию
type engHours struct {
	engHoursStart   float64 // моточасы на начало
	engHoursCurrent float64 // последняя обновленноая запись моточасов
	engHoursEnd     float64 // моточасы на конец
}

// метод переброса данных из интерфейса в структуру
func (e *engHours) loadingInterfaceData(interfaceData interface{}) error {
	engHoursInterface, err := typeConversion[engHoursDataInterface](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"engHours"}, err)
		return err
	}

	e.engHoursStart = engHoursInterface.GetEngHoursStart()
	e.engHoursEnd = engHoursInterface.GetEngHoursEnd()
	e.engHoursCurrent = engHoursInterface.GetEngHoursCurrent()

	return err
}

// данные по сессии водителя в смене
type sessionDriverData struct {
	sessionId         int         // id сессии, берется из БД
	offset            int         // последний записанный offset
	timeStartSession  time.Time   // время старта сессии
	timeUpdateSession time.Time   // время обновления записи сессии
	avSpeed           float64     // средняя скорость водителя
	engHoursData      engHours    // данные по моточасам за сессию
	mileageData       mileageData // данные по пробегу за смену
	mileageGPSData    mileageData // данные пробега по GPS за сессию
}

// метод для загрузки данных в структуру из интерфеса
func (s *sessionDriverData) loadingInterfaceData(interfaceData interface{}) error {
	dataDriverSession, err := typeConversion[dataDriverSessionFromStorage](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"sessionDriverData"}, err)
		return err
	}

	s.sessionId = dataDriverSession.GetSessionId()
	s.offset = dataDriverSession.GetOffset()
	s.timeStartSession = dataDriverSession.GetStartSession()
	s.timeUpdateSession = dataDriverSession.GetUpdatedTime()
	s.avSpeed = dataDriverSession.GetAvSpeed()

	err = s.engHoursData.loadingInterfaceData(dataDriverSession.GetEngHoursData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"sessionDriverData"}, err)
		return err
	}
	err = s.mileageData.loadingInterfaceData(dataDriverSession.GetMileageData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"sessionDriverData"}, err)
		return err
	}
	err = s.mileageGPSData.loadingInterfaceData(dataDriverSession.GetMileageGPSData())
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"sessionDriverData"}, err)
		return err
	}

	return err
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

// стукрура содержащая сконвертированные интерфейсы ответа от модуля storage
type storageAnswerData struct {
	shiftData         shiftObjData
	driverSessionData sessionDriverData
}
