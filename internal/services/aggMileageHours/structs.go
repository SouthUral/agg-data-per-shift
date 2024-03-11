package aggmileagehours

import (
	"sync"
	"time"
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

// метод добаления смены
// TODO: нужно сделать проверку что смена не пересекается с другими сменами
func (s *settingsDurationShifts) AddShiftSetting(numShift, shiftDuration int, startTimeShift time.Time) {
	s.mx.Lock()
	s.shifts[numShift] = settingShift{
		numShift:       numShift,
		startTimeShift: startTimeShift,
		shiftDuration:  shiftDuration,
	}
	s.mx.Unlock()
}

// метод для определения номера и даты смены
func (s *settingsDurationShifts) defineShift(dateEvent time.Time) (int, time.Time, error) {
	// определять смену нужно по текущей дате в событии
	var numShift int
	var dateShift time.Time
	var err error

	s.mx.RLock()
	for numShift, shiftSettings := range s.shifts {
		t := time.Date(dateEvent.Year(),
			dateEvent.Month(),
			dateEvent.Day(),
			shiftSettings.startTimeShift.Hour(),
			shiftSettings.startTimeShift.Minute(),
			shiftSettings.startTimeShift.Second(),
			shiftSettings.startTimeShift.Nanosecond(),
			dateEvent.Local().Location(),
		)
		startShift := t.Add(-time.Duration(s.offsetTimeShift) * time.Hour)
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

func initNewMileageData(mileage int) *mileageData {
	newMileageData := &mileageData{
		mileageStart:                mileage,
		mileageCurrent:              0,
		mileageEnd:                  mileage,
		mileageLoaded:               0, // значение 0 т.к. неизвестно в каком состоянии находится машина
		mileageAtBeginningOfLoading: 0, // значение 0 т.к. неизвестно в каком состоянии находится машина
		mileageEmpty:                0, // значение 0 т.к. неизвестно в каком состоянии находится машина

	}
	return newMileageData
}

// функция создает структуру mileageData на основании данных из БД
func initNewMileageDataLoadingDBData(data interface{}) (*mileageData, error) {
	newMileageData := &mileageData{}
	err := newMileageData.loadingData(data)
	return newMileageData, err
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

func (m mileageData) GetMileageStart() int {
	return m.mileageStart
}
func (m mileageData) GetMileageCurrent() int {
	return m.mileageCurrent
}
func (m mileageData) GetMileageEnd() int {
	return m.mileageEnd
}
func (m mileageData) GetMileageLoaded() int {
	return m.mileageLoaded
}
func (m mileageData) GetMileageAtBeginningOfLoading() int {
	return m.mileageAtBeginningOfLoading
}
func (m mileageData) GetMileageEmpty() int {
	return m.mileageEmpty
}

func (m *mileageData) loadingData(data interface{}) error {
	mData, err := typeСonversion[mileageDataInterface](data)
	if err != nil {
		return err
	}

	m.mileageStart = mData.GetMileageStart()
	m.mileageCurrent = mData.GetMileageCurrent()
	m.mileageEnd = mData.GetMileageEnd()
	m.mileageLoaded = mData.GetMileageLoaded()
	m.mileageAtBeginningOfLoading = mData.GetMileageAtBeginningOfLoading()
	m.mileageEmpty = mData.GetMileageEmpty()
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
			m.mileageLoaded = mileage - m.mileageAtBeginningOfLoading
		} else {
			// если mileageAtBeginningOfLoading не был установлен, то это первое событие в груженом состоянии,
			// нужно установить значение mileageAtBeginningOfLoading
			m.mileageAtBeginningOfLoading = mileage
		}
	} else {
		if m.mileageAtBeginningOfLoading != 0 {
			// если при обновлении флаг загрузки будет false, но поле mileageAtBeginningOfLoading не будет 0, то
			// это означает что текущее событие было событием разгрузки и пройденное расстояние до разгрузки входит в груженный пробег
			m.mileageLoaded = mileage - m.mileageAtBeginningOfLoading
			m.mileageAtBeginningOfLoading = 0
		}
	}
	m.mileageEmpty = m.mileageCurrent - m.mileageLoaded
}

func initNewEngHours(eventData *eventData) *engHours {
	engHours := &engHours{
		engHoursStart:   eventData.engineHours,
		engHoursEnd:     eventData.engineHours,
		engHoursCurrent: 0,
	}
	return engHours
}

// функция создает структуру engHours на основании данных из БД
func initEngHoursLoadingDBData(data interface{}) (*engHours, error) {
	newEngHours := &engHours{}
	err := newEngHours.loadingData(data)
	return newEngHours, err
}

// данные по моточасам за смену/сессию
type engHours struct {
	engHoursStart   float32 // моточасы на начало
	engHoursCurrent float32 // последняя обновленноая запись моточасов
	engHoursEnd     float32 // моточасы на конец
}

func (e engHours) GetEngHoursStart() float32 {
	return e.engHoursStart
}
func (e engHours) GetEngHoursCurrent() float32 {
	return e.engHoursCurrent
}
func (e engHours) GetEngHoursEnd() float32 {
	return e.engHoursEnd
}

func (e *engHours) loadingData(data interface{}) error {
	eData, err := typeСonversion[engHoursDataInterface](data)
	if err != nil {
		return err
	}

	e.engHoursStart = eData.GetEngHoursStart()
	e.engHoursEnd = eData.GetEngHoursEnd()
	e.engHoursCurrent = eData.GetEngHoursCurrent()
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
func (e *engHours) updateEngHours(engHours float32) {
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
		EngineHours float32 `json:"engine_hours"` // data -> engine_hours : моточасы
		AvSpeed     float32 `json:"s_av_speed"`   // data -> s_av_speed : средняя скорость водителя на технике
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
		avSpeed:     e.Data.AvSpeed,
	}

	return data, err
}

type eventData struct {
	typeEvent   string    // тип события
	objectID    int       // id техники
	mesTime     time.Time // время сообщения
	mileage     int       // пробег
	gpsMileage  int       // пробег по gps
	engineHours float32   // моточасы
	fioDriver   string    // ФИО водителя
	numDriver   int       // номер водителя
	avSpeed     float32   // средняя скрорость водителя на технике
}
