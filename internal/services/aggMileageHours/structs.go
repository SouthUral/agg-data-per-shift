package aggmileagehours

import (
	"fmt"
	"sync"
	"time"

	utils "agg-data-per-shift/pkg/utils"
)

func initSettingsDurationShifts(offsetTimeShift int) *settingsDurationShifts {
	res := &settingsDurationShifts{
		mx:              sync.RWMutex{},
		shifts:          make(map[int]settingShift),
		offsetTimeShift: offsetTimeShift,
	}

	return res
}

// настройки смены
type settingsDurationShifts struct {
	mx              sync.RWMutex
	shifts          map[int]settingShift
	offsetTimeShift int // времянное смещение смены
}

type settingShift struct {
	numShift       int       // номер смены
	startTimeShift time.Time // время старта смены
	shiftDuration  int       // продолжительность смены
}

// метод для определения номера и даты смены
func (s settingsDurationShifts) defineShift(dateEvent time.Time) (int, time.Time, error) {
	var numShift int
	var dateShift time.Time
	var err error

	s.mx.RLock()
	for numShift, shiftSettings := range s.shifts {
		startShift := shiftSettings.startTimeShift.Add(-time.Duration(s.offsetTimeShift) * time.Hour)
		endShift := startShift.Add(time.Duration(shiftSettings.shiftDuration) * time.Hour)
		if dateEvent.After(startShift) && dateEvent.Before(endShift) {
			// дата смены равна дате окончанию смены  т.к. смена может начинаться вечером в прошлый день
			return numShift, endShift, err
		}
	}
	defer s.mx.RUnlock()

	err = defineShiftError{}

	return numShift, dateShift, err
}

// данные по пробегу за смену/сессию
type mileageData struct {
	mileageStart                int // пробег на начало (смены/сессии)
	mileageCurrent              int // текущий пробег
	mileageEnd                  int // проебег на конец смены
	mileageLoaded               int // пробег в груженом состоянии
	mileageAtBeginningOfLoading int // пробег на начало последней погрузки (поле обнуляется после разгрузки)
	mileageEmpty                int // пробег в порожнем состоянии
}

// интерфейсный метод
func (m mileageData) GetMileageStart() int {
	return m.mileageStart
}

// интерфейсный метод
func (m mileageData) GetMileageCurrent() int {
	return m.mileageCurrent
}

// интерфейсный метод
func (m mileageData) GetMileageEnd() int {
	return m.mileageEnd
}

// интерфейсный метод
func (m mileageData) GetMileageLoaded() int {
	return m.mileageLoaded
}

// интерфейсный метод
func (m mileageData) GetMileageAtBeginningOfLoading() int {
	return m.mileageAtBeginningOfLoading
}

// интерфейсный метод
func (m mileageData) GetMileageEmpty() int {
	return m.mileageEmpty
}

// метод переброса данных из интерфейса в структуру
func (m *mileageData) loadingInterfaceData(interfaceData interface{}) error {
	mileageInterface, err := utils.TypeConversion[mileageDataInterface](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"mileageData"}, err)
		return err
	}

	m.mileageStart = mileageInterface.GetMileageStart()
	m.mileageEnd = mileageInterface.GetMileageEnd()
	m.mileageCurrent = mileageInterface.GetMileageCurrent()
	m.mileageLoaded = mileageInterface.GetMileageLoaded()
	m.mileageEmpty = mileageInterface.GetMileageEmpty()
	m.mileageAtBeginningOfLoading = mileageInterface.GetMileageAtBeginningOfLoading()

	return err
}

// метод для создания новой структуры mileageData на основании старой
func (m *mileageData) createNewMileageData() *mileageData {
	newMileageData := &mileageData{
		mileageStart:                m.mileageEnd,
		mileageEnd:                  m.mileageEnd,
		mileageAtBeginningOfLoading: m.mileageEnd, // значение обнулится, если окажется что объект не груженый
	}
	return newMileageData
}

// метод для обновления данных по моточасам
func (m *mileageData) updateMileageData(mileage int, objLoaded bool) {
	m.mileageEnd = mileage
	m.mileageCurrent = m.mileageEnd - m.mileageStart
	if objLoaded {
		if m.mileageAtBeginningOfLoading != 0 {
			m.mileageLoaded += mileage - m.mileageAtBeginningOfLoading
		} else {
			m.mileageAtBeginningOfLoading = mileage
		}
	} else {
		if m.mileageAtBeginningOfLoading != 0 {
			// если при обновлении флаг загрузки будет false, но поле mileageAtBeginningOfLoading не будет 0, то
			// это означает что текущее событие было событием разгрузки и пройденное расстояние до разгрузки входит в груженный пробег
			m.mileageLoaded += mileage - m.mileageAtBeginningOfLoading
			m.mileageAtBeginningOfLoading = 0
		}
	}
	m.mileageEmpty = m.mileageCurrent - m.mileageLoaded
}

// данные по моточасам за смену/сессию
type engHours struct {
	engHoursStart   float64 // моточасы на начало
	engHoursCurrent float64 // последняя обновленноая запись моточасов
	engHoursEnd     float64 // моточасы на конец
}

// Интерфейсный метод
func (e engHours) GetEngHoursStart() float64 {
	return e.engHoursStart
}

// Интерфейсный метод
func (e engHours) GetEngHoursCurrent() float64 {
	return e.engHoursCurrent
}

// Интерфейсный метод
func (e engHours) GetEngHoursEnd() float64 {
	return e.engHoursEnd
}

// метод переброса данных из интерфейса в структуру
func (e *engHours) loadingInterfaceData(interfaceData interface{}) error {
	engHoursInterface, err := utils.TypeConversion[engHoursDataInterface](interfaceData)
	if err != nil {
		err = fmt.Errorf("%w: %w", interfaceLoadingToStructError{"engHours"}, err)
		return err
	}

	e.engHoursStart = engHoursInterface.GetEngHoursStart()
	e.engHoursEnd = engHoursInterface.GetEngHoursEnd()
	e.engHoursCurrent = engHoursInterface.GetEngHoursCurrent()

	return err
}

// метод для создания новой структуры engHours на основании данных старой структуры
func (e *engHours) createNewEngHours() *engHours {
	newEngHours := &engHours{
		engHoursStart: e.engHoursEnd,
		engHoursEnd:   e.engHoursEnd,
	}
	return newEngHours
}

// метод для обновления данных по моточасам
func (e *engHours) updateEngHours(engHours float64) {
	e.engHoursEnd = engHours
	e.engHoursCurrent = e.engHoursEnd - e.engHoursStart
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
		avSpeed     float64 `json:"s_av_speed"`   // data -> s_av_speed : средняя скорость водителя на технике
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
		avSpeed:     e.Data.avSpeed,
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
	avSpeed     float64   // средняя скрорость водителя на технике
}

// стукрура содержащая сконвертированные интерфейсы ответа от модуля storage
type storageAnswerData struct {
	shiftData         *shiftObjData
	driverSessionData *sessionDriverData
}
