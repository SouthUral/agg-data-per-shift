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
	GetDataDriverSession() interface{} // возвращает интерфейс, который потом нужно привести к интерфейсу
}

// данные смены (используются при восстановлении состояния)
type dataShiftFromStorage interface {
	GetShiftId() int           // возвращает Id смены (индентификатор в таблице БД)
	GetOffset() int            // offset последнего события, которое было применено к смене
	GetUpdatedTime() time.Time // время последнего обновления (пока под вопросом)
	GetShiftDate() time.Time   // возвращает дату смены

	GetEngHoursStart() float64   // моточасы на начало смены
	GetEngHoursCurrent() float64 // показатель моточасов на последнее обновление записи
	GetEngHoursEnd() float64     // моточасы на конец смены

	GetMileageStart() int   // пробег на начало смены
	GetMileageCurrent() int // пробег на последнее обновление записи
	GetMileageEnd() int     // пробег на конец смены

	GetMileageLoaded() int // пробег груженым

	GetMileageGPSStart() int   // пробег по GPS начало смены
	GetMileageGPSCurrent() int // пробег по GPS на последнее обновление записи
	GetMileageGPSEnd() int     // пробег по GPS на конец смены

	GetMileageGPSLoaded() int // пробег груженым по GPS

	GetStatusLoaded() bool // груженый или нет (показатель всегда на послденее обновление записи)
}

// данные текущей сессии водителя, в текущей смене (последней обновленной смене)
type dataDriverSessionFromStorage interface {
	GetSessionId() int          // id сессии
	GetOffset() int             // offset последнего события, которое было применено к сессии
	GetStartSession() time.Time // время начала сессии
	GetUpdatedTime() time.Time  // время последнего обновления сессии

	GetAvSpeed() float64 // средняя скорость водителя в сессии

	GetEngHoursStartSession() float64 // моточасы на начало сессии
	GetEngHoursCurrent() float64      // последний обновленный показатель моточасов
	GetEngHoursEndSession() float64   // моточасы на конец сессии

	GetMileageStartSession() int // пробег на начало сессии
	GetMileageCurrent() int      // последний обновленный пробег
	GetMileageEndSession() int   //  пробег на конец сессии

	GetMileageLoaded() int // пробег груженым во время сессии

	GetMileageGPSStartSession() int // пробег по GPS на начало сессии
	GetMileageGPSCurrent() int      // последний обновленный пробег по GPS
	GetMileageGPSEndSession() int   //  пробег по GPS на конец сессии

	GetMileageGPSLoaded() int // пробег по GPS груженым во время сессии
}
