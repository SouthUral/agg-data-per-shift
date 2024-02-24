package aggmileagehours

import (
	"time"
)

type incomingMessage interface {
	GetOffset() int64
	GetMsg() []byte
}

// интерфейс сообщения от модуля storage
type incomingMessageFromStorage interface {
	GetDataShift() interface{}         // возвращает интерфейс, который потом нужно привести к интерфейсу dataShiftFromStorage
	GetDataDriverSession() interface{} // возвращает интерфейс, который потом нужно привести к интерфейсу dataDriverSessionFromStorage
}

// данные смены (используются при восстановлении состояния)
type dataShiftFromStorage interface {
	GetShiftId() int           // возвращает Id смены (индентификатор в таблице БД)
	GetShiftNum() int          // возвращает номер смены
	GetOffset() int            // offset последнего события, которое было применено к смене
	GetUpdatedTime() time.Time // время последнего обновления (пока под вопросом)
	GetShiftDate() time.Time   // возвращает дату смены
	GetStatusLoaded() bool     // груженый или нет (показатель всегда на послденее обновление записи)

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за смену
	GetMileageData() interface{}    // получение интерфейса к данным пробега за смену
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за смену
}

// данные текущей сессии водителя, в текущей смене (последней обновленной смене)
type dataDriverSessionFromStorage interface {
	GetSessionId() int          // id сессии
	GetOffset() int             // offset последнего события, которое было применено к сессии
	GetStartSession() time.Time // время начала сессии
	GetUpdatedTime() time.Time  // время последнего обновления сессии

	GetAvSpeed() float64 // средняя скорость водителя в сессии

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за сессию
	GetMileageData() interface{}    // получение интерфейса к данным пробега за сессию
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за сессию
}

// интерфейс получения данных о моточасах
type engHoursDataInterface interface {
	GetEngHoursStart() float64   // моточасы на начало смены/сессии
	GetEngHoursCurrent() float64 // последний обновленный показатель моточасов
	GetEngHoursEnd() float64     // моточасы на конец смены/сессии
}

// интерфейс получения данных о пробеге
type mileageDataInterface interface {
	GetMileageStart() int   // пробег на начало смены/сессии
	GetMileageCurrent() int // пробег на последнее обновление записи
	GetMileageEnd() int     // пробег на конец смены/сессии
	GetMileageLoaded() int  // пробег груженым
	GetMileageEmpty() int   // пробег порожним
}
