package aggmileagehours

import (
	"time"
)

// интерфейс сообщения от amqp
type incomingMessage interface {
	GetOffset() int64
	GetMsg() []byte
}

// интерфейс сообщения от модуля storage
type incomingMessageFromStorage interface {
	GetDataShift() interface{}         // возвращает json, который нужно привести к структуре shiftObjData
	GetDataDriverSession() interface{} // возвращает json, который нужно привести к структуре sessionDriverData
	GetError() error                   // возращает ошибку
}

// интерфейс для получения ответа от Storage на запрос добавления новый записей в таблицу смены и сессий
type responceStorageToAddNewShiftAndSession interface {
	GetShiftId() int
	GetSessionId() int
	GetError() error
}

// данные смены
type dataShiftFromStorage interface {
	GetShiftId() int  // возвращает Id смены (индентификатор в таблице БД)
	GetShiftNum() int // возвращает номер смены
	GetShiftDateStart() time.Time
	GetShiftDateEnd() time.Time
	GetShiftDate() time.Time   // возвращает дату смены
	GetUpdatedTime() time.Time // время последнего обновления (пока под вопросом)
	GetOffset() int64          // offset последнего события, которое было применено к смене
	GetCurrentDriverId() int
	GetStatusLoaded() bool // груженый или нет (показатель всегда на послденее обновление записи)

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за смену
	GetMileageData() interface{}    // получение интерфейса к данным пробега за смену
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за смену
}

// данные текущей сессии водителя
type dataDriverSessionFromStorage interface {
	GetShiftId() int                 // id смены, в которой находится сессия
	GetSessionId() int               // id сессии
	GetDriverId() int                // id водителя
	GetOffset() int64                // offset последнего события, которое было применено к сессии
	GetTimeStartSession() time.Time  // время начала сессии
	GetTimeUpdateSession() time.Time // время последнего обновления сессии

	GetAvSpeed() float32 // средняя скорость водителя в сессии

	GetEngHoursData() interface{}   // получение интерфейса к данным о моточасах за сессию
	GetMileageData() interface{}    // получение интерфейса к данным пробега за сессию
	GetMileageGPSData() interface{} // получение интерфейса к данным пробега по GPS за сессию
}

// интерфейс получения данных о моточасах
type engHoursDataInterface interface {
	GetEngHoursStart() float32   // моточасы на начало смены/сессии
	GetEngHoursCurrent() float32 // последний обновленный показатель моточасов
	GetEngHoursEnd() float32     // моточасы на конец смены/сессии
}

// интерфейс получения данных о пробеге
type mileageDataInterface interface {
	GetMileageStart() int   // пробег на начало смены/сессии
	GetMileageCurrent() int // пробег на последнее обновление записи
	GetMileageEnd() int     // пробег на конец смены/сессии
	GetMileageLoaded() int  // пробег груженым
	GetMileageEmpty() int   // пробег порожним
}
